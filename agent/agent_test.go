package agent

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/logger"
	"github.com/negasus/haproxy-spoe-go/request"
)

// countingListener counts calls to Close
type countingListener struct {
	net.Listener
	closes atomic.Int32
}

func (l *countingListener) Close() error {
	l.closes.Add(1)
	return l.Listener.Close()
}

func newListener(t *testing.T) *countingListener {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	return &countingListener{Listener: l}
}

func newAgent() *Agent {
	return New(func(*request.Request) {}, logger.NewNop())
}

func waitResult(t *testing.T, errCh <-chan error) error {
	t.Helper()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return")
		return nil
	}
}

func TestServeContext_Cancel(t *testing.T) {
	l := newListener(t)
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- newAgent().ServeContext(ctx, l)
	}()

	cancel()

	if err := waitResult(t, errCh); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := l.closes.Load(); got != 1 {
		t.Fatalf("expected listener to be closed once, got %d", got)
	}
}

func TestServe_ListenerClosed(t *testing.T) {
	l := newListener(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- newAgent().Serve(l)
	}()

	// Make sure Serve is accepting before closing the listener
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	conn.Close()

	l.Close()

	if err := waitResult(t, errCh); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected net.ErrClosed, got %v", err)
	}
}

func TestServeContext_WatcherStopsOnReturn(t *testing.T) {
	l := newListener(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- newAgent().ServeContext(ctx, l)
	}()

	l.Close()

	if err := waitResult(t, errCh); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected net.ErrClosed, got %v", err)
	}

	// The watcher has exited, so cancelling must not close the listener again
	cancel()

	if got := l.closes.Load(); got != 1 {
		t.Fatalf("expected listener to be closed once, got %d", got)
	}
}
