package frame

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/negasus/haproxy-spoe-go/varint"
)

func (f *Frame) Read(src io.Reader) error {
	var n int
	var err error

	n, err = io.ReadFull(src, f.tmp[:])
	if err != nil {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("error read frame size, %v", err)
	}

	f.Len = binary.BigEndian.Uint32(f.tmp[0:4])
	f.Type = Type(f.tmp[4])

	// Drop packet that doesn't have defined frame type early, before allocating any buffers
	// that way spurious connections (say someone calling curl on port) won't cause it to
	// allocate gigabytes of RAM
	switch f.Type {
	case TypeHaproxyHello, TypeHaproxyDisconnect, TypeNotify, TypeAgentHello, TypeAgentDisconnect, TypeAgentAck:
	default:
		return fmt.Errorf("unexpected frame type %d", f.Type)
	}

	buf := make([]byte, f.Len-1)

	n, err = io.ReadFull(src, buf)
	if err != nil {
		return fmt.Errorf("error read frame, %v", err)
	}

	if uint32(n) != f.Len-1 {
		return fmt.Errorf("unexpected frame length %d, expect %d", n, f.Len)
	}

	f.Flags = binary.BigEndian.Uint32(buf[0:4])
	buf = buf[4:]

	f.StreamID, n = varint.Uvarint(buf)
	buf = buf[n:]

	f.FrameID, n = varint.Uvarint(buf)
	buf = buf[n:]

	switch f.Type {
	case TypeHaproxyHello, TypeHaproxyDisconnect:
		if err = f.KV.Unmarshal(buf); err != nil {
			return err
		}
		if v, ok := f.KV.Get("healthcheck"); ok && v.(bool) {
			f.Healthcheck = true
		}
		if v, ok := f.KV.Get("max-frame-size"); ok {
			f.MaxFrameSize = v.(uint32)
		}
		if v, ok := f.KV.Get("engine-id"); ok {
			f.EngineID = v.(string)
		}

	case TypeNotify:
		err = f.Messages.Decode(buf)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unexpected frame type %d", f.Type)
	}

	return nil
}
