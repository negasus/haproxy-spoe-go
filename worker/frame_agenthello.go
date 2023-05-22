package worker

import (
	"bytes"
	"fmt"

	"github.com/negasus/haproxy-spoe-go/frame"
)

func (w *worker) sendAgentHello(haproxyHello *frame.Frame) error {
	var err error
	var frameSize, n int

	agentHello := frame.AcquireFrame()
	defer frame.ReleaseFrame(agentHello)

	agentHello.Type = frame.TypeAgentHello
	agentHello.FrameID = haproxyHello.FrameID
	agentHello.StreamID = haproxyHello.StreamID

	agentHello.KV.Add("version", "2.0")
	agentHello.KV.Add("max-frame-size", haproxyHello.MaxFrameSize)
	agentHello.KV.Add("capabilities", capabilities)

	buf := bytes.NewBuffer(make([]byte, 0))

	frameSize, err = agentHello.Encode(buf)
	if err != nil {
		return fmt.Errorf("marshaling error: %v", err)
	}

	n, err = w.conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error write to connection: %v", err)
	}
	if n != frameSize {
		return fmt.Errorf("write unexpected bytes count %d, expect %d", n, frameSize)
	}

	return nil
}
