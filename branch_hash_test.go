package frozen_test

import (
	"testing"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen"

	"github.com/arr-ai/frozen/internal/pkg/test"
)

type intWithBadHash int

var _ hash.Hashable = intWithBadHash(0)

func (i intWithBadHash) Hash(seed uintptr) uintptr {
	return hash.Int(int(i)%100, seed)
}

func TestBadHash(t *testing.T) {
	t.Parallel()

	const N = 10000
	var b frozen.SetBuilder[intWithBadHash]
	for i := 0; i < N; i += 10 {
		b.Add(intWithBadHash(i))
	}

	for i := 0; i < N; i += 10 {
		test.True(t, b.Has(intWithBadHash(i)))
	}
	for i := N; i < 2*N; i += 10 {
		test.False(t, b.Has(intWithBadHash(i)))
	}
}

func TestRemoveCollider(t *testing.T) {
	t.Parallel()

	var b frozen.SetBuilder[intWithBadHash]
	b.Add(intWithBadHash(100))
	b.Add(intWithBadHash(200))
	b.Remove(intWithBadHash(100))
	test.True(t, b.Has(intWithBadHash(200)))
}

// constantHash is an int that always hashes to 0, forcing all elements into
// the same bucket at every tree level. This triggers splitLeaf's recursive
// path down to maxSplitDepth.
type constantHash int

var _ hash.Hashable = constantHash(0)

func (constantHash) Hash(uintptr) uintptr { return 0 }

func TestPathologicalCollision(t *testing.T) {
	t.Parallel()

	// 20 elements with identical hashes overflows maxLeafLen (8) and forces
	// splitLeaf to recurse to maxSplitDepth, creating a deeply nested tree.
	const N = 20

	// Build via SetBuilder (exercises Add/AddFast path).
	var b frozen.SetBuilder[constantHash]
	for i := 0; i < N; i++ {
		b.Add(constantHash(i))
	}
	s := b.Finish()

	test.Equal(t, N, s.Count())
	for i := 0; i < N; i++ {
		test.True(t, s.Has(constantHash(i)))
	}
	for i := N; i < 2*N; i++ {
		test.False(t, s.Has(constantHash(i)))
	}

	// With on the pathological set (exercises withFastBatched through deep branches).
	s2 := s.With(constantHash(N))
	test.Equal(t, N+1, s2.Count())
	test.True(t, s2.Has(constantHash(N)))

	// Without on the pathological set.
	s3 := s.Without(constantHash(0))
	test.Equal(t, N-1, s3.Count())
	test.False(t, s3.Has(constantHash(0)))

	// Original set unchanged.
	test.Equal(t, N, s.Count())
	test.True(t, s.Has(constantHash(0)))

	// Union of two pathological sets.
	var b2 frozen.SetBuilder[constantHash]
	for i := N / 2; i < N+N/2; i++ {
		b2.Add(constantHash(i))
	}
	s4 := b2.Finish()
	union := s.Union(s4)
	test.Equal(t, N+N/2, union.Count())

	// Intersection of two pathological sets.
	inter := s.Intersection(s4)
	test.Equal(t, N/2, inter.Count())
}
