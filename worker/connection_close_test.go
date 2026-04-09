package worker

import (
	"errors"
	"io"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/client"
	"github.com/negasus/haproxy-spoe-go/request"
)

func TestIsConnectionClose_EOF(t *testing.T) {
	if !isConnectionClose(io.EOF) {
		t.Error("io.EOF should be treated as connection close")
	}
}

func TestIsConnectionClose_UnexpectedEOF(t *testing.T) {
	if !isConnectionClose(io.ErrUnexpectedEOF) {
		t.Error("io.ErrUnexpectedEOF should be treated as connection close")
	}
}

func TestIsConnectionClose_ECONNRESET(t *testing.T) {
	if !isConnectionClose(syscall.ECONNRESET) {
		t.Error("ECONNRESET should be treated as connection close")
	}
}

func TestIsConnectionClose_EPIPE(t *testing.T) {
	if !isConnectionClose(syscall.EPIPE) {
		t.Error("EPIPE should be treated as connection close")
	}
}

func TestIsConnectionClose_WrappedECONNRESET(t *testing.T) {
	err := &net.OpError{
		Op:  "read",
		Net: "tcp",
		Err: syscall.ECONNRESET,
	}
	if !isConnectionClose(err) {
		t.Error("wrapped ECONNRESET in net.OpError should be treated as connection close")
	}
}

func TestIsConnectionClose_RandomError(t *testing.T) {
	if isConnectionClose(errors.New("something unexpected")) {
		t.Error("arbitrary errors should not be treated as connection close")
	}
}

type recordingLogger struct {
	messages []string
}

func (l *recordingLogger) Errorf(format string, args ...interface{}) {
	l.messages = append(l.messages, format)
}

func TestWorker_ConnectionResetDoesNotLogError(t *testing.T) {
	clientConn, server := net.Pipe()

	log := &recordingLogger{}
	done := make(chan struct{})

	go func() {
		Handle(server, func(r *request.Request) {}, log)
		close(done)
	}()

	spoe := client.NewClient(clientConn)
	if err := spoe.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Close the client side abruptly (simulates HAProxy RST)
	clientConn.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after connection close")
	}

	for _, msg := range log.messages {
		if msg == "handle worker: %v" {
			t.Errorf("connection close should not produce an error log, got: %v", log.messages)
		}
	}
}
