package agent

import (
	"net"

	"github.com/github/haproxy-spoe-go/logger"
	"github.com/github/haproxy-spoe-go/request"
	"github.com/github/haproxy-spoe-go/worker"
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

func (agent *Agent) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}

		go worker.Handle(conn, agent.handler, agent.logger)
	}
}
