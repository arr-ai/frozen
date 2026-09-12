package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen/v2"
)

// The h0 content hash is an XOR of element hashes. Equal must never treat a
// matching h0 as proof of equality: user-supplied hashes need only satisfy
// equal ⇒ equal-hash, and XOR is not injective.
func TestSetEqualNestedPartitionsDiffer(t *testing.T) {
	t.Parallel()

	a := frozen.NewSet(frozen.NewSet(1, 2), frozen.NewSet(3))
	b := frozen.NewSet(frozen.NewSet(1), frozen.NewSet(2, 3))
	if a.Equal(b) {
		t.Errorf("%v.Equal(%v) reported true", a, b)
	}

	m1 := frozen.NewSet(
		frozen.NewMap(frozen.KV("a", 1), frozen.KV("b", 2)),
		frozen.NewMap(frozen.KV("c", 3)),
	)
	m2 := frozen.NewSet(
		frozen.NewMap(frozen.KV("a", 1)),
		frozen.NewMap(frozen.KV("b", 2), frozen.KV("c", 3)),
	)
	if m1.Equal(m2) {
		t.Errorf("%v.Equal(%v) reported true", m1, m2)
	}
}

func TestSetEqualSameRootShortCircuits(t *testing.T) {
	t.Parallel()

	s := frozen.Iota(1 << 10)
	if !s.Equal(s) {
		t.Error("a set must equal itself")
	}
	if !s.Equal(frozen.Iota(1 << 10)) {
		t.Error("independently built sets with the same elements must be equal")
	}
}
