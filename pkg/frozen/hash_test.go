package frozen

import "testing"

func TestHash64(t *testing.T) {
	if hash(uint64(0)) == 0 {
		t.Error()
	}
}

func TestHash64String(t *testing.T) {
	if hash("hello") == 0 {
		t.Error()
	}
}
