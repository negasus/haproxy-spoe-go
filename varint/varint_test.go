package varint

import (
	"testing"
)

func TestPutUVarint(t *testing.T) {
	var n int

	buf := make([]byte, 10)

	n = PutUvarint(buf, 239)
	if n != 1 {
		t.Fatalf("n must be 1, got %d", n)
	}
	if buf[0] != 0xEF {
		t.Fatal("wrong buf value")
	}

	n = PutUvarint(buf, 240)
	if n != 2 {
		t.Fatalf("n must be 2, got %d", n)
	}
	if buf[0] != 0xF0 {
		t.Fatal("wrong buf value")
	}
	if buf[1] != 0x00 {
		t.Fatal("wrong buf value")
	}

	n = PutUvarint(buf, 256)
	if n != 2 {
		t.Fatalf("n must be 2, got %d", n)
	}
	if buf[0] != 0xF0 {
		t.Fatal("wrong buf value")
	}
	if buf[1] != 0x01 {
		t.Fatal("wrong buf value")
	}

	n = PutUvarint(buf, 2287)
	if n != 2 {
		t.Fatalf("n must be 2, got %d", n)
	}
	if buf[0] != 0xFF {
		t.Fatal("wrong buf value")
	}
	if buf[1] != 0x7F {
		t.Fatal("wrong buf value")
	}

	n = PutUvarint(buf, 2289)
	if n != 3 {
		t.Fatalf("n must be 3, got %d", n)
	}
	if buf[0] != 0xF1 {
		t.Fatal("wrong buf value")
	}
	if buf[1] != 0x80 {
		t.Fatal("wrong buf value")
	}
	if buf[2] != 0x00 {
		t.Fatal("wrong buf value")
	}
}

func TestGetUVarint(t *testing.T) {
	var n uint64
	var c int

	n, c = Uvarint([]byte{0xF0})
	if n != uint64(0) {
		t.Fatalf("n must be 0, got %d", n)
	}
	if c != -1 {
		t.Fatalf("c must be -1, got %d", c)
	}

	n, c = Uvarint([]byte{0xEF})
	if n != uint64(239) {
		t.Fatalf("n must be 239, got %d", n)
	}
	if c != 1 {
		t.Fatalf("c must be 1, got %d", c)
	}

	n, c = Uvarint([]byte{0xF1, 0x00})
	if n != uint64(241) {
		t.Fatalf("n must be 241, got %d", n)
	}
	if c != 2 {
		t.Fatalf("c must be 2, got %d", c)
	}

	n, c = Uvarint([]byte{0xF0, 0x01})
	if n != uint64(256) {
		t.Fatalf("n must be 256, got %d", n)
	}
	if c != 2 {
		t.Fatalf("c must be 2, got %d", c)
	}

	n, c = Uvarint([]byte{0xFF, 0x7F})
	if n != uint64(2287) {
		t.Fatalf("n must be 2287, got %d", n)
	}
	if c != 2 {
		t.Fatalf("c must be 2, got %d", c)
	}

	n, c = Uvarint([]byte{0xF1, 0x80, 0x00})
	if n != uint64(2289) {
		t.Fatalf("n must be 2289, got %d", n)
	}
	if c != 3 {
		t.Fatalf("c must be 3, got %d", c)
	}
}

func TestLoop(t *testing.T) {
	buf := make([]byte, 10)

	for i := 0; i < 1e6; i++ {
		PutUvarint(buf, uint64(i))
		n, _ := Uvarint(buf)
		if n != uint64(i) {
			t.Fatal("unexpected n value")
		}
	}
}
