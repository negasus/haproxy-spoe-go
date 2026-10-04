package typeddata

import (
	"bytes"
	"math"
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

// Regression test for issue #10: values above ~2^53 need a 10-byte varint
// buffer, not 8. A representative current-day Unix nano timestamp is the
// easiest realistic trigger.
func TestEncode_Int64_NanoTimestamp(t *testing.T) {
	v := int64(1700000000000000000)
	buf, n, err := Encode(v, make([]byte, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("n (%d) does not match len(buf) (%d)", n, len(buf))
	}
	decoded, _, err := Decode(buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	got, ok := decoded.(int64)
	if !ok {
		t.Fatalf("decoded type = %T, want int64", decoded)
	}
	if got != v {
		t.Fatalf("round-trip mismatch: got %d, want %d", got, v)
	}
}

// Regression test for issue #10: math.MaxUint64 needs the full 10-byte
// varint range.
func TestEncode_UInt64_Max(t *testing.T) {
	v := uint64(math.MaxUint64)
	buf, n, err := Encode(v, make([]byte, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("n (%d) does not match len(buf) (%d)", n, len(buf))
	}
	decoded, _, err := Decode(buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	got, ok := decoded.(uint64)
	if !ok {
		t.Fatalf("decoded type = %T, want uint64", decoded)
	}
	if got != v {
		t.Fatalf("round-trip mismatch: got %d, want %d", got, v)
	}
}

// Regression test for issue #10: negative int32 sign-extends to ~MaxUint64
// when widened to uint64, so it also needs the full 10-byte varint range.
func TestEncode_Int32_Negative(t *testing.T) {
	v := int32(-1)
	buf, n, err := Encode(v, make([]byte, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("n (%d) does not match len(buf) (%d)", n, len(buf))
	}
	decoded, _, err := Decode(buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	got, ok := decoded.(int32)
	if !ok {
		t.Fatalf("decoded type = %T, want int32", decoded)
	}
	if got != v {
		t.Fatalf("round-trip mismatch: got %d, want %d", got, v)
	}
}
