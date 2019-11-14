package typeddata

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncode_Nil(t *testing.T) {
	buf, n, err := Encode(nil, make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, len(buf))
	assert.Equal(t, byte(0x00), buf[0])
}

func TestEncode_Bool(t *testing.T) {
	buf, n, err := Encode(false, make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, len(buf))
	assert.Equal(t, byte(0x10), buf[0])

	buf, n, err = Encode(true, make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, len(buf))
	assert.Equal(t, byte(0x11), buf[0])
}

func TestEncode_Int32(t *testing.T) {
	buf, n, err := Encode(int32(100500), make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, len(buf))
	assert.Equal(t, []byte{0x02, 0xF4, 0xFA, 0x2F}, buf)
}

func TestEncode_Binary(t *testing.T) {
	buf, n, err := Encode([]byte{0x10, 0x20, 0x30}, make([]byte, 0))
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, len(buf))
	assert.Equal(t, []byte{0x09, 0x10, 0x20, 0x30}, buf)
}

// todo: tests
