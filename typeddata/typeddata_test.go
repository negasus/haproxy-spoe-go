package typeddata

import (
	"bytes"
	"testing"
)

func TestEncode_Nil(t *testing.T) {
	buf, n, err := Encode(nil, make([]byte, 0))
	if err != nil {
		t.Fatal("unexpected error")
	}
	if n != 1 {
		t.Fatalf("n must be 1, got %d", n)
	}
	if len(buf) != 1 {
		t.Fatalf("buf len must be 1, got %d", len(buf))
	}
	if buf[0] != 0x00 {
		t.Fatalf("invalid buf value")
	}
}

func TestEncode_Bool(t *testing.T) {
	buf, n, err := Encode(false, make([]byte, 0))
	if err != nil {
		t.Fatal("unexpected error")
	}
	if n != 1 {
		t.Fatalf("n must be 1, got %d", n)
	}
	if len(buf) != 1 {
		t.Fatalf("buf len must be 1, got %d", len(buf))
	}
	if buf[0] != 0x01 {
		t.Fatalf("invalid buf value")
	}

	buf, n, err = Encode(true, make([]byte, 0))
	if err != nil {
		t.Fatal("unexpected error")
	}
	if n != 1 {
		t.Fatalf("n must be 1, got %d", n)
	}
	if len(buf) != 1 {
		t.Fatalf("buf len must be 1, got %d", len(buf))
	}
	if buf[0] != 0x11 {
		t.Fatalf("invalid buf value")
	}
}

func TestEncode_Int32(t *testing.T) {
	buf, n, err := Encode(int32(100500), make([]byte, 0))
	if err != nil {
		t.Fatal("unexpected error")
	}
	if n != 4 {
		t.Fatalf("n must be 4, got %d", n)
	}
	if len(buf) != 4 {
		t.Fatalf("buf len must be 4, got %d", len(buf))
	}
	if !bytes.Equal(buf, []byte{0x02, 0xF4, 0xFA, 0x2F}) {
		t.Fatalf("invalid buf value")
	}
}

func TestEncode_Binary(t *testing.T) {
	buf, n, err := Encode([]byte{0x10, 0x20, 0x30}, make([]byte, 0))
	if err != nil {
		t.Fatal("unexpected error")
	}
	if n != 5 {
		t.Fatalf("n must be 4, got %d", n)
	}
	if len(buf) != 5 {
		t.Fatalf("buf len must be 4, got %d", len(buf))
	}
	if !bytes.Equal(buf, []byte{0x09, 0x03, 0x10, 0x20, 0x30}) {
		t.Fatalf("invalid buf value")
	}
}
