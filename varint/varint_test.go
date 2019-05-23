package varint

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPutUVarint(t *testing.T) {
	var n int

	buf := make([]byte, 10)

	n = PutUvarint(buf, 239)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte(0xEF), buf[0])

	n = PutUvarint(buf, 240)
	assert.Equal(t, 2, n)
	assert.Equal(t, byte(0xF0), buf[0])
	assert.Equal(t, byte(0x00), buf[1])

	n = PutUvarint(buf, 256)
	assert.Equal(t, 2, n)
	assert.Equal(t, byte(0xF0), buf[0])
	assert.Equal(t, byte(0x01), buf[1])

	n = PutUvarint(buf, 2287)
	assert.Equal(t, 2, n)
	assert.Equal(t, byte(0xFF), buf[0])
	assert.Equal(t, byte(0x7F), buf[1])

	n = PutUvarint(buf, 2289)
	assert.Equal(t, 3, n)
	assert.Equal(t, byte(0xF1), buf[0])
	assert.Equal(t, byte(0x80), buf[1])
	assert.Equal(t, byte(0x00), buf[2])
}

func TestGetUVarint(t *testing.T) {
	var n uint64
	var c int

	n, c = Uvarint([]byte{0xF0})
	assert.Equal(t, uint64(0), n)
	assert.Equal(t, -1, c)

	n, c = Uvarint([]byte{0xEF})
	assert.Equal(t, uint64(239), n)
	assert.Equal(t, 1, c)

	n, c = Uvarint([]byte{0xF1, 0x00})
	assert.Equal(t, uint64(241), n)
	assert.Equal(t, 2, c)

	n, c = Uvarint([]byte{0xF0, 0x01})
	assert.Equal(t, uint64(256), n)
	assert.Equal(t, 2, c)

	n, c = Uvarint([]byte{0xFF, 0x7F})
	assert.Equal(t, uint64(2287), n)
	assert.Equal(t, 2, c)

	n, c = Uvarint([]byte{0xF1, 0x80, 0x00})
	assert.Equal(t, uint64(2289), n)
	assert.Equal(t, 3, c)
}

func TestLoop(t *testing.T) {
	buf := make([]byte, 10)

	for i := 0; i < 1e6; i++ {
		PutUvarint(buf, uint64(i))
		n, _ := Uvarint(buf)
		assert.Equal(t, uint64(i), n)
	}
}
