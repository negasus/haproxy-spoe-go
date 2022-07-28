package request

import (
	"sync"

	"github.com/negasus/haproxy-spoe-go/action"
	"github.com/negasus/haproxy-spoe-go/message"
)

var requestPool = sync.Pool{
	New: func() interface{} {
		return newRequest()
	},
}

type Request struct {
	EngineID string
	StreamID uint64
	FrameID  uint64
	Messages *message.Messages
	Actions  action.Actions
}

func newRequest() *Request {
	m := &Request{
		Messages: message.NewMessages(),
		Actions:  make(action.Actions, 0, 1),
	}

	return m
}

func AcquireRequest() *Request {
	m := requestPool.Get()
	if m == nil {
		return newRequest()
	}

	return m.(*Request)
}

func ReleaseRequest(m *Request) {
	m.Reset()
	requestPool.Put(m)
}

func (req *Request) Reset() {

	req.Messages.Reset()
	req.Actions.Reset()

	req.EngineID = ""
	req.StreamID = 0
	req.FrameID = 0
}
