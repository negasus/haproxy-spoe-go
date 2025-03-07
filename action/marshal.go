package action

import (
	"fmt"

	"github.com/github/haproxy-spoe-go/typeddata"
	"github.com/github/haproxy-spoe-go/varint"
)

func (action *Action) Marshal(buf []byte) ([]byte, error) {
	var nb byte

	switch action.Type {
	case TypeSetVar:
		nb = nbVarsSetVar
	case TypeUnsetVar:
		nb = nbVarsUnsetVar
	default:
		return nil, fmt.Errorf("unexpected action type: %v", action.Type)
	}

	buf = append(buf, byte(action.Type))
	buf = append(buf, nb)
	buf = append(buf, byte(action.Scope))

	b := make([]byte, 10)
	n := varint.PutUvarint(b, uint64(len(action.Name)))

	buf = append(buf, b[:n]...)
	buf = append(buf, action.Name...)

	valueBuf, n, err := typeddata.Encode(action.Value, make([]byte, 0))
	if err != nil {
		return nil, err
	}

	buf = append(buf, valueBuf[:n]...)

	return buf, nil
}
