package agent

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/frame"
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

func sendFrame(t *testing.T, c net.Conn, f *frame.Frame) {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	if _, err := f.Encode(buf); err != nil {
		t.Fatalf("encode frame: %v", err)
	}
	if _, err := c.Write(buf.Bytes()); err != nil {
		t.Fatalf("write frame: %v", err)
	}
}

// readFrameType reads one frame and returns its type.
// frame.Read returns an error for Agent* frames after consuming the payload,
// so only the type is checked.
func readFrameType(t *testing.T, r *bufio.Reader) frame.Type {
	t.Helper()

	f := frame.AcquireFrame()
	defer frame.ReleaseFrame(f)

	_ = f.Read(r)

	return f.Type
}

func TestServeContext_AcceptedConnectionSurvivesCancel(t *testing.T) {
	l := newListener(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- newAgent().ServeContext(ctx, l)
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Bound the whole exchange so a regression fails rather than hangs
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	reader := bufio.NewReader(conn)

	// Complete the hello exchange so the connection is known to be handled by a worker
	hello := frame.AcquireFrame()
	hello.Type = frame.TypeHaproxyHello
	hello.KV.Add("supported-versions", "2.0")
	hello.KV.Add("max-frame-size", uint32(16*1024))
	hello.KV.Add("capabilities", "")
	sendFrame(t, conn, hello)
	frame.ReleaseFrame(hello)

	if got := readFrameType(t, reader); got != frame.TypeAgentHello {
		t.Fatalf("unexpected response type: got %v, want AgentHello", got)
	}

	cancel()

	if err := waitResult(t, errCh); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// The accepted connection must still be served after ServeContext returns
	disc := frame.AcquireFrame()
	disc.Type = frame.TypeHaproxyDisconnect
	disc.KV.Add("status-code", uint32(0))
	disc.KV.Add("message", "normal")
	sendFrame(t, conn, disc)
	frame.ReleaseFrame(disc)

	if got := readFrameType(t, reader); got != frame.TypeAgentDisconnect {
		t.Fatalf("unexpected response type: got %v, want AgentDisconnect", got)
	}
}
