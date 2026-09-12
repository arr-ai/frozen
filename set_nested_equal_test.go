package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
)

// The h0 content hash is an XOR of element hashes, and a nested collection's
// Hash128 is its h0. XOR is associative, so collections with the same
// flattened elements but different partitions share an h0. Equal must never
// treat a matching h0 as proof of equality.
func TestSetEqualNestedPartitionsDiffer(t *testing.T) {
	t.Parallel()

	a := frozen.NewSet(frozen.NewSet(1, 2), frozen.NewSet(3))
	b := frozen.NewSet(frozen.NewSet(1), frozen.NewSet(2, 3))
	if a.Hash128() != b.Hash128() {
		t.Log("hashes differ; the equality check below is not exercising an h0 match")
	}
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
