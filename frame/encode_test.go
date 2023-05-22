package frame

import (
	"bytes"
	"encoding/binary"
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
	if err != nil {
		t.Fatalf("expect err is nil, got %v", err)
	}
	bufBytes := buf.Bytes()
	encodedFrameSize := int(binary.BigEndian.Uint32(bufBytes[0:4]))
	if frameSize-4 != encodedFrameSize {
		t.Fatal("wrong frame size")
	}
	if string(bufBytes[13:17]) != "key1" {
		t.Fatal("expect key1")
	}
	if string(bufBytes[19:23]) != "val1" {
		t.Fatal("expect val1")
	}
	if string(bufBytes[24:28]) != "key2" {
		t.Fatal("expect key2")
	}
	if string(bufBytes[30:34]) != "val2" {
		t.Fatal("expect val1")
	}
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
