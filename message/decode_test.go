package message

import (
	"bytes"
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
	if err != nil {
		t.Fatal("unexpected error")
	}
	if mess.Len() != 2 {
		t.Fatalf("mess.Len must be 2, got %d", mess.Len())
	}

	// First message
	m, err := mess.GetByIndex(0)
	if err != nil {
		t.Fatal("unexpected error")
	}
	if m.Name != "Foo" {
		t.Fatalf("m.Name must be Foo, got %s", m.Name)
	}

	v, ok := m.KV.Get("Bar")
	if !ok {
		t.Fatal("ok is not true")
	}
	if !bytes.Equal([]byte{0x10, 0x20, 0x30}, v.([]byte)) {
		t.Fatal("invalid result")
	}

	v, ok = m.KV.Get("Vals")
	if !ok {
		t.Fatal("ok is not true")
	}
	if v != "Baz" {
		t.Fatalf("v must be Baz, got %s", v)
	}

	// Second message
	m, err = mess.GetByIndex(1)
	if err != nil {
		t.Fatal("unexpected error")
	}
	if m.Name != "UI" {
		t.Fatalf("m.Name must be UI, got %s", m.Name)
	}

	v, ok = m.KV.Get("Fee")
	if !ok {
		t.Fatal("ok is not true")
	}
	if v != int32(10) {
		t.Fatalf("v must be int32(10), got %d", v)
	}
}
