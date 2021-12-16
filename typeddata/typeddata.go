package typeddata

import (
	"fmt"
	"github.com/negasus/haproxy-spoe-go/varint"
	"net"
	"reflect"

	"github.com/pkg/errors"
)

const (
	// TypeNull const for TypedData type
	TypeNull byte = 0
	// TypeBoolean const for TypedData type
	TypeBoolean byte = 1
	// TypeInt32 const for TypedData type
	TypeInt32 byte = 2
	// TypeUInt32 const for TypedData type
	TypeUInt32 byte = 3
	// TypeInt64 const for TypedData type
	TypeInt64 byte = 4
	// TypeUInt64 const for TypedData type
	TypeUInt64 byte = 5
	// TypeIPv4 const for TypedData type
	TypeIPv4 byte = 6
	// TypeIPv6 const for TypedData type
	TypeIPv6 byte = 7
	// TypeString const for TypedData type
	TypeString byte = 8
	// TypeBinary const for TypedData type
	TypeBinary byte = 9
)

//ErrEmptyBuffer describe error, if passed empty buffer for decoding
var ErrEmptyBuffer = errors.New("empty buffer for decode")

//ErrDecodingBufferTooSmall describe error for too small decoding buffer
var ErrDecodingBufferTooSmall = errors.New("decoding buffer too small")

//Encode variable to TypedData value
//returns filled buffer, count of bytes and error
func Encode(data interface{}, buf []byte) ([]byte, int, error) {
	var n int

	switch v := data.(type) {
	case nil:
		buf = append(buf, TypeNull)
		return buf, 1, nil

	case bool:
		var b byte = 0x11
		if !v {
			b = 0x10
		}
		buf = append(buf, b)
		return buf, 1, nil

	case int32:
		buf = append(buf, TypeInt32)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case uint32:
		buf = append(buf, TypeUInt32)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case int:
		buf = append(buf, TypeInt64)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case int64:
		buf = append(buf, TypeInt64)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case uint:
		buf = append(buf, TypeUInt64)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case uint64:
		buf = append(buf, TypeUInt64)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(v))
		buf = append(buf, b[:i]...)
		return buf, i + 1, nil

	case string:
		n = 1
		buf = append(buf, TypeString)
		b := make([]byte, 8)
		i := varint.PutUvarint(b, uint64(len(v)))
		n += i
		n += len(v)
		buf = append(buf, b[:i]...)
		buf = append(buf, v...)
		return buf, n, nil

	case []byte:
		buf = append(buf, TypeBinary)
		buf = append(buf, v...)
		return buf, len(v) + 1, nil
	}

	return nil, 0, fmt.Errorf("type not supported for encode to TypedData: %s", reflect.TypeOf(data).String())
}

//Decode TypedData value
//Returns decoded variable, bytes count and error
func Decode(buf []byte) (data interface{}, n int, err error) {
	if len(buf) == 0 {
		err = ErrEmptyBuffer
		return
	}

	f := buf[0] >> 4
	t := buf[0] & 0x0F
	buf = buf[1:]
	n = 1

	switch t {
	case TypeNull:
		return

	case TypeBoolean:
		data = f&0x01 > 0
		return

	case TypeInt32:
		i, l := varint.Uvarint(buf)
		n += l
		data = int32(i)
		return

	case TypeUInt32:
		i, l := varint.Uvarint(buf)
		n += l
		data = uint32(i)
		return

	case TypeInt64:
		i, l := varint.Uvarint(buf)
		n += l
		data = int64(i)
		return

	case TypeUInt64:
		i, l := varint.Uvarint(buf)
		n += l
		data = uint64(i)
		return

	case TypeIPv4:
		data = net.IP(buf[:4])
		n += 4
		return

	case TypeIPv6:
		data = net.IP(buf[:16])
		n += 16
		return

	case TypeString:
		sLen, i := varint.Uvarint(buf)
		n += i
		buf = buf[i:]
		if len(buf) < int(sLen) {
			err = ErrDecodingBufferTooSmall
			return
		}
		data = string(buf[:sLen])
		n += int(sLen)
		return

	case TypeBinary:
		dataLen, i := varint.Uvarint(buf)
		n += i
		buf = buf[i:]
		if len(buf) < int(dataLen) {
			err = ErrDecodingBufferTooSmall
			return
		}
		data = buf[:dataLen]
		n += int(dataLen)
		return
	}

	return nil, n, fmt.Errorf("type %d not supported for decode from TypedData", t)
}
