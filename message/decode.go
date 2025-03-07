package message

import (
	"github.com/github/haproxy-spoe-go/varint"
)

func (m *Messages) Decode(buf []byte) error {
	for {
		if len(buf) == 0 {
			break
		}

		message := AcquireMessage()

		messageNameLen, n := varint.Uvarint(buf)
		buf = buf[n:]
		message.Name = string(buf[:messageNameLen])
		buf = buf[messageNameLen:]

		nbArgs := int(buf[0])
		buf = buf[1:]

		n, err := message.KV.UnmarshalNB(buf, nbArgs)

		if err != nil {
			return err
		}

		buf = buf[n:]

		*m = append(*m, message)
	}

	return nil
}
