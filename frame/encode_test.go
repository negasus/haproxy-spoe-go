package frame

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFrame_Write(t *testing.T) {
	f := NewFrame()
	f.Type = TypeAgentDisconnect
	f.FrameID = 123
	f.StreamID = 456
	f.KV.Add("key1", "val1")
	f.KV.Add("key2", "val2")
	buf := &bytes.Buffer{}
	frameSize, err := f.Encode(buf)
	require.Nil(t, err)
	bufBytes := buf.Bytes()
	encodedFrameSize := int(binary.BigEndian.Uint32(bufBytes[0:4]))
	assert.Equal(t, frameSize-4, encodedFrameSize, "frame size")
	assert.Equal(t, "key1", string(bufBytes[13:17]))
	assert.Equal(t, "val1", string(bufBytes[19:23]))
	assert.Equal(t, "key2", string(bufBytes[24:28]))
	assert.Equal(t, "val2", string(bufBytes[30:34]))

}

func BenchmarkFrame_Encode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := NewFrame()
		f.Type = TypeAgentDisconnect
		f.FrameID = 123
		f.StreamID = 456
		f.KV.Add("key1", "val1")
		f.KV.Add("key2", "val2")
		buf := &bytes.Buffer{}
		_, _ = f.Encode(buf)
	}
}
