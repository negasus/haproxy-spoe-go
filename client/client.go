package client

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/negasus/haproxy-spoe-go/frame"
	_ "github.com/negasus/haproxy-spoe-go/request"
	"io"
	"net"
)

type Client struct {
	conn   net.Conn
	reader io.Reader
}

func NewClient(conn net.Conn) Client {
	return Client{conn: conn, reader: bufio.NewReader(conn)}
}

func (c *Client) Init() error {
	f := frame.AcquireFrame()
	defer frame.ReleaseFrame(f)
	f.Type = frame.TypeHaproxyHello
	f.StreamID = 0
	f.FrameID = 0
	f.KV.Add("supported-versions", "2")
	f.KV.Add("max-frame-size", uint32(16*1024))
	f.KV.Add("capabilities", "")

	err := c.send(f)
	if err != nil {
		return err
	}

	responseFrame := frame.AcquireFrame()
	defer frame.ReleaseFrame(responseFrame)
	responseFrame.Read(c.reader)
	// todo read frame

	return nil

}

func (c *Client) send(f *frame.Frame) error {
	buf := bytes.NewBuffer(make([]byte, 0))
	n, err := f.Encode(buf)
	if err != nil {
		return err
	}
	n, err = c.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buf.Len() {
		return fmt.Errorf("size mismatch")
	}
	return nil
}

func (c *Client) Notify() error {
	f := frame.AcquireFrame()
	defer frame.ReleaseFrame(f)
	f.Type = frame.TypeNotify
	f.StreamID = 1
	f.FrameID = 1

	err := c.send(f)
	if err != nil {
		return err
	}

	responseFrame := frame.AcquireFrame()
	defer frame.ReleaseFrame(responseFrame)
	responseFrame.Read(c.reader)

	return nil

}
func (c *Client) Stop() error {
	f := frame.AcquireFrame()
	defer frame.ReleaseFrame(f)
	f.Type = frame.TypeHaproxyDisconnect
	f.StreamID = 0
	f.FrameID = 0
	f.KV.Add("status-code", uint32(0))
	f.KV.Add("message", "normal")

	err := c.send(f)
	if err != nil {
		return err
	}

	responseFrame := frame.AcquireFrame()
	defer frame.ReleaseFrame(responseFrame)
	responseFrame.Read(c.reader)

	return nil

}
