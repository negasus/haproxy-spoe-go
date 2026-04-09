package worker

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"

	"github.com/negasus/haproxy-spoe-go/frame"
	"github.com/negasus/haproxy-spoe-go/logger"
	"github.com/negasus/haproxy-spoe-go/request"
)

const (
	capabilities = "pipelining,async"
)

// Handle listen connection and process frames
func Handle(conn net.Conn, handler func(*request.Request), logger logger.Logger) {
	w := &worker{
		conn:    conn,
		handler: handler,
		logger:  logger,
	}

	if err := w.run(); err != nil {
		logger.Errorf("handle worker: %v", err)
	}
}

type worker struct {
	conn     net.Conn
	ready    bool
	engineID string
	handler  func(*request.Request)

	logger logger.Logger

	wg sync.WaitGroup
}

func (w *worker) close() {
	if err := w.conn.Close(); err != nil {
		w.logger.Errorf("close connection: %v", err)
	}
}

func (w *worker) run() error {

	defer func() {
		// Wait for all in-flight notify handlers to finish before closing conn
		w.wg.Wait()
		w.close()
	}()

	var f *frame.Frame

	buf := bufio.NewReader(w.conn)

	for {
		f = frame.AcquireFrame()

		if err := f.Read(buf); err != nil {
			frame.ReleaseFrame(f)
			if isConnectionClose(err) {
				return nil
			}
			return fmt.Errorf("error read frame: %v", err)
		}

		switch f.Type {
		case frame.TypeHaproxyHello:

			if w.ready {
				return fmt.Errorf("worker already ready, but got HaproxyHello frame")
			}

			if err := w.sendAgentHello(f); err != nil {
				frame.ReleaseFrame(f)
				return fmt.Errorf("error send AgentHello frame: %v", err)
			}

			if f.Healthcheck {
				frame.ReleaseFrame(f)
				return nil
			}

			w.engineID = f.EngineID

			w.ready = true
			continue

		case frame.TypeHaproxyDisconnect:
			if !w.ready {
				return fmt.Errorf("worker not ready, but got HaproxyDisconnect frame")
			}

			if err := w.sendAgentDisconnect(f, 0, "connection closed by server"); err != nil {
				return fmt.Errorf("error send AgentDisconnect frame: %v", err)
			}
			frame.ReleaseFrame(f)
			return nil

		case frame.TypeNotify:
			if !w.ready {
				return fmt.Errorf("worker not ready, but got Notify frame")
			}

			w.wg.Add(1)
			go w.processNotifyFrame(f)

		default:
			w.logger.Errorf("unexpected frame type: %v", f.Type)
		}
	}
}

// isConnectionClose reports whether err indicates the peer closed the
// connection. HAProxy 3.x's mux_spop tears down TCP connections without
// sending a DISCONNECT frame when all SPOP streams are done, resulting
// in ECONNRESET or EPIPE instead of io.EOF.
func isConnectionClose(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	var netErr *net.OpError
	return errors.As(err, &netErr) && !netErr.Temporary()
}
