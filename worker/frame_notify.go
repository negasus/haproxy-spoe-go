package worker

import (
	"bytes"
	"github.com/negasus/haproxy-spoe-go/frame"
	"github.com/negasus/haproxy-spoe-go/request"
	"log"
)

func (w *worker) processNotifyFrame(f *frame.Frame) {

	defer frame.ReleaseFrame(f)

	var err error
	var n int

	req := request.AcquireRequest()
	defer request.ReleaseRequest(req)

	req.StreamID = f.StreamID
	req.FrameID = f.FrameID
	req.EngineID = w.engineID
	req.Messages = f.Messages

	w.handler(req)

	ackFrame := frame.AcquireFrame()
	defer frame.ReleaseFrame(ackFrame)

	ackFrame.Type = frame.TypeAgentAck
	ackFrame.StreamID = f.StreamID
	ackFrame.FrameID = f.FrameID
	ackFrame.Actions = req.Actions

	buf := bytes.NewBuffer(make([]byte, 0))
	n, err = ackFrame.Encode(buf)
	if err != nil {
		log.Printf("error marshal ack frame: %v", err)
		return
	}

	n, err = w.conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("error write ack frame: %v", err)
		return
	}

	if n != buf.Len() {
		log.Printf("write wrong data count %d, expect %d", n, buf.Len())
	}
}
