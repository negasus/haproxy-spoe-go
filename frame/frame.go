package frame

import (
	"github.com/negasus/haproxy-spoe-go/action"
	"github.com/negasus/haproxy-spoe-go/message"
	"github.com/negasus/haproxy-spoe-go/payload/kv"
	"sync"
)

type Type byte

const (
	TypeUnset             Type = 0x00
	TypeHaproxyHello      Type = 0x01
	TypeHaproxyDisconnect Type = 0x02
	TypeNotify            Type = 0x03
	TypeAgentHello        Type = 0x65
	TypeAgentDisconnect   Type = 0x66
	TypeAgentAck          Type = 0x67
)

var framePool = sync.Pool{
	New: func() interface{} {
		return NewFrame()
	},
}

func AcquireFrame() *Frame {
	return framePool.Get().(*Frame)
}

func ReleaseFrame(frame *Frame) {
	frame.Reset()
	framePool.Put(frame)
}

//Frame describe frame struct
type Frame struct {
	Len          uint32
	Type         Type
	Flags        uint32
	EngineID     string
	StreamID     uint64
	FrameID      uint64
	Healthcheck  bool
	MaxFrameSize uint32
	KV           *kv.KV
	Messages     *message.Messages
	Actions      *action.Actions

	tmp       []byte
	varintBuf []byte
}

// NewFrame creates and returns new Frame
// Fin byte in Flags already set
func NewFrame() *Frame {
	f := &Frame{
		Flags:     0x01,
		KV:        kv.AcquireKV(),
		Messages:  message.NewMessages(),
		Actions:   action.NewActions(),
		tmp:       make([]byte, 4),
		varintBuf: make([]byte, 10),
	}

	return f
}

func (f *Frame) Reset() {
	f.Len = 0
	f.Type = 0
	f.Flags = 0x01
	f.EngineID = ""
	f.StreamID = 0
	f.FrameID = 0
	f.Healthcheck = false
	f.MaxFrameSize = 0

	// we want to acquire a new Actions as they are shared with Request
	f.Actions = action.NewActions()
	f.Messages.Reset()
	f.KV.Reset()
}

//IsFin returns true, if frame has flag 'FIN'
func (f *Frame) IsFin() bool {
	return f.Flags&0x01 > 0
}

//IsAbort returns true, if frame has flag 'ABORT'
func (f *Frame) IsAbort() bool {
	return f.Flags&0x02 > 0
}
