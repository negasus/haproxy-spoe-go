package worker

import (
	"bytes"
	"fmt"
	"github.com/negasus/haproxy-spoe-go/frame"
)

func (w *worker) sendAgentDisconnect(f *frame.Frame, statusCode uint32, message string) error {
	var frameSize, n int
	var err error

	agentDisconnectFrame := frame.AcquireFrame()
	defer frame.ReleaseFrame(agentDisconnectFrame)

	agentDisconnectFrame.Type = frame.TypeAgentDisconnect
	agentDisconnectFrame.FrameID = f.FrameID
	agentDisconnectFrame.StreamID = f.StreamID
	agentDisconnectFrame.KV.Add("status-code", statusCode)
	agentDisconnectFrame.KV.Add("message", message)

	buf := &bytes.Buffer{}
	frameSize, err = agentDisconnectFrame.Encode(buf)
	if err != nil {
		return err
	}

	n, err = w.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != frameSize {
		return fmt.Errorf("write unexpected bytes count %d, expect %d", n, frameSize)
	}

	return nil
}
