package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/github/haproxy-spoe-go/varint"
)

func (f *Frame) Encode(dest io.Writer) (n int, err error) {
	buf := bytes.Buffer{}

	buf.WriteByte(byte(f.Type))

	binary.BigEndian.PutUint32(f.tmp[:], f.Flags)

	buf.Write(f.tmp[0:4])

	n = varint.PutUvarint(f.varintBuf[:], f.StreamID)
	buf.Write(f.varintBuf[:n])

	n = varint.PutUvarint(f.varintBuf[:], f.FrameID)
	buf.Write(f.varintBuf[:n])

	var payload []byte

	switch f.Type {
	case TypeAgentHello, TypeAgentDisconnect, TypeHaproxyHello, TypeHaproxyDisconnect:
		payload, err = f.KV.Bytes()
		if err != nil {
			return
		}

	case TypeAgentAck:
		if f.Actions != nil {
			for _, act := range f.Actions {
				payload, err = act.Marshal(payload)
				if err != nil {
					return
				}
			}
		}
	case TypeNotify:
		if len(*f.Messages) > 0 {
			err = fmt.Errorf("encoding Notify frame with Message isn't handled yet")
			return

		}
	default:
		err = fmt.Errorf("unexpected frame type %d", f.Type)
		return
	}

	buf.Write(payload)

	binary.BigEndian.PutUint32(f.tmp[:], uint32(buf.Len()))

	n, err = dest.Write(f.tmp[0:4])
	if err != nil || n != 4 {
		return 0, fmt.Errorf("error write frameSize. writes %d, expect %d, err: %v", n, len(f.tmp), err)
	}

	n, err = dest.Write(buf.Bytes())
	if err != nil || n != buf.Len() {
		return 0, fmt.Errorf("error write frame. writes %d, expect %d, err: %v", n, len(f.tmp), err)
	}

	return 4 + buf.Len(), nil
}
