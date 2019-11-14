package message

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDecode(t *testing.T) {
	mess := NewMessages()

	// Two message with 2 and 1 values in KV
	buf := []byte{
		0x03, 'F', 'o', 'o', 0x02, 0x03, 'B', 'a', 'r', 0x09, 0x03, 0x10, 0x20, 0x30, 0x04, 'V', 'a', 'l', 's', 0x08, 0x03, 'B', 'a', 'z',
		0x02, 'U', 'I', 0x01, 0x03, 'F', 'e', 'e', 0x02, 0x0A,
	}

	err := mess.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, 2, mess.Len())

	// First message
	m, err := mess.GetByIndex(0)
	require.NoError(t, err)
	assert.Equal(t, "Foo", m.Name)

	v, ok := m.KV.Get("Bar")
	require.True(t, ok)
	require.Equal(t, []byte{0x10, 0x20, 0x30}, v)

	v, ok = m.KV.Get("Vals")
	require.True(t, ok)
	require.Equal(t, "Baz", v)

	// Second message
	m, err = mess.GetByIndex(1)
	require.NoError(t, err)
	assert.Equal(t, "UI", m.Name)

	v, ok = m.KV.Get("Fee")
	require.True(t, ok)
	require.Equal(t, int32(10), v)

}
