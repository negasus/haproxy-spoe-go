package agent

import (
	"context"
	"net"

	"github.com/negasus/haproxy-spoe-go/logger"
	"github.com/negasus/haproxy-spoe-go/request"
	"github.com/negasus/haproxy-spoe-go/worker"
)

func New(handler func(*request.Request), logger logger.Logger) *Agent {
	agent := &Agent{
		handler: handler,
		logger:  logger,
	}

	return agent
}

type Agent struct {
	handler func(*request.Request)
	logger  logger.Logger
}

// Serve accepts connections on the listener and handles each in a new goroutine.
// It is equivalent to ServeContext with context.Background().
func (agent *Agent) Serve(listener net.Listener) error {
	return agent.ServeContext(context.Background(), listener)
}

// ServeContext accepts connections on the listener and handles each in a new goroutine
// until ctx is done or the listener returns a non-temporary error.
//
// When ctx is done the listener is closed and ctx.Err() is returned.
// Connections already accepted are not interrupted.
func (agent *Agent) ServeContext(ctx context.Context, listener net.Listener) error {
	// context.Background() and similar are never done, so no watcher is needed
	if ctx.Done() != nil {
		done := make(chan struct{})
		stopped := make(chan struct{})

		go func() {
			defer close(stopped)
			select {
			case <-ctx.Done():
				listener.Close()
			case <-done:
			}
		}()

		// Make sure the watcher has exited before returning
		defer func() {
			close(done)
			<-stopped
		}()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}

		go worker.Handle(conn, agent.handler, agent.logger)
	}
}
