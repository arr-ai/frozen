package hash

import (
	"testing"
)

func TestStringAndBytesUseDistinctSalts(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"", "a", "ab", "abcdefgh", "0123456789abcdef", "0123456789abcdefX"} {
		if String(s) == Bytes([]byte(s)) {
			t.Errorf("%q: String and Bytes hashed alike", s)
		}
	}
}

func TestMixBytesCoversEveryByte(t *testing.T) {
	t.Parallel()
	// Overlapping reads must still see a difference in the last byte of a
	// 16-byte key and in the first byte of a 4-byte key.
	if String("0123456789abcdef") == String("0123456789abcdeX") {
		t.Error("16-byte strings that differ in the last byte hashed alike")
	}
	if String("abcd") == String("abcX") {
		t.Error("4-byte strings that differ in the last byte hashed alike")
	}
	if String("abcd") == String("Xbcd") {
		t.Error("4-byte strings that differ in the first byte hashed alike")
	}
	if String("ab") == String("ba") {
		t.Error("swapped 2-byte strings hashed alike")
	}
}

func TestMixBytesMixesLength(t *testing.T) {
	t.Parallel()
	if Bytes([]byte{1}) == Bytes([]byte{1, 0}) {
		t.Error("[]byte{1} and []byte{1, 0} hashed alike")
	}
	if String("a") == String("a\x00") {
		t.Error("\"a\" and \"a\\x00\" hashed alike")
	}
	if String("") == String("\x00") {
		t.Error("empty string and NUL hashed alike")
	}
}

func TestBytesNilAndEmptyHashAlike(t *testing.T) {
	t.Parallel()
	var empty []byte
	if Bytes(nil) != Bytes(empty) {
		t.Error("nil and empty []byte hashed differently")
	}
}

func TestAnyDispatchesBytes(t *testing.T) {
	t.Parallel()
	b := []byte("xy")
	if Any(b) != Bytes(b) {
		t.Error("Any([]byte) did not match Bytes")
	}
	if Any("xy") != String("xy") {
		t.Error("Any(string) did not match String")
	}
}
