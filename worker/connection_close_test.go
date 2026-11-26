package worker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/client"
	"github.com/negasus/haproxy-spoe-go/frame"
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

// faultConn injects transport errors while preserving the real frame parser.
type faultConn struct {
	net.Conn
	input    *bytes.Reader
	readErr  error
	writeErr error
	closed   bool
}

func (c *faultConn) Read(p []byte) (int, error) {
	if c.input != nil && c.input.Len() > 0 {
		return c.input.Read(p)
	}
	if c.readErr != nil {
		return 0, c.readErr
	}
	return 0, io.EOF
}
func (c *faultConn) Write(p []byte) (int, error) {
	if c.writeErr != nil {
		return 0, c.writeErr
	}
	return len(p), nil
}
func (c *faultConn) Close() error { c.closed = true; return nil }

func TestWorkerDisconnectWriteErrors(t *testing.T) {
	unexpected := errors.New("unexpected write failure")
	for _, network := range []string{"unix", "tcp"} {
		for _, tc := range []struct {
			name    string
			err     error
			wantLog bool
		}{
			{"broken pipe", syscall.EPIPE, false},
			{"reset", syscall.ECONNRESET, false},
			{"closed", net.ErrClosed, false},
			{"unexpected", unexpected, true},
		} {
			t.Run(network+"/"+tc.name, func(t *testing.T) {
				var input bytes.Buffer
				hello := frame.AcquireFrame()
				hello.Type = frame.TypeHaproxyHello
				hello.KV.Add("supported-versions", "2")
				hello.KV.Add("max-frame-size", uint32(16384))
				hello.KV.Add("capabilities", "")
				if _, err := hello.Encode(&input); err != nil {
					t.Fatal(err)
				}
				frame.ReleaseFrame(hello)
				disconnect := frame.AcquireFrame()
				disconnect.Type = frame.TypeHaproxyDisconnect
				disconnect.KV.Add("status-code", uint32(0))
				if _, err := disconnect.Encode(&input); err != nil {
					t.Fatal(err)
				}
				frame.ReleaseFrame(disconnect)
				// Allow the hello response, then fail the disconnect response.
				c := &disconnectFaultConn{faultConn: faultConn{input: bytes.NewReader(input.Bytes())}, err: &net.OpError{Op: "write", Net: network, Err: tc.err}}
				log := &recordingLogger{}
				Handle(c, func(*request.Request) {}, log)
				if (len(log.messages) > 0) != tc.wantLog {
					t.Fatalf("logs = %v, want error log = %v", log.messages, tc.wantLog)
				}
				if c.writes != 2 || !c.closed {
					t.Fatalf("writes = %d, closed = %v", c.writes, c.closed)
				}
			})
		}
	}
}

type disconnectFaultConn struct {
	faultConn
	writes int
	err    error
}

func (c *disconnectFaultConn) Write(p []byte) (int, error) {
	c.writes++
	if c.writes == 2 {
		return 0, c.err
	}
	return len(p), nil
}

func TestWorkerReadConnectionResetDoesNotLogError(t *testing.T) {
	for _, prefix := range [][]byte{nil, {0, 0, 0, 7, byte(frame.TypeHaproxyDisconnect)}} {
		t.Run(fmt.Sprintf("prefix_%d", len(prefix)), func(t *testing.T) {
			c := &faultConn{input: bytes.NewReader(prefix), readErr: &net.OpError{Op: "read", Net: "unix", Err: syscall.ECONNRESET}}
			log := &recordingLogger{}
			Handle(c, func(*request.Request) {}, log)
			if len(log.messages) != 0 {
				t.Fatalf("unexpected logs: %v", log.messages)
			}
			if !c.closed {
				t.Fatal("connection was not closed")
			}
		})
	}
}
