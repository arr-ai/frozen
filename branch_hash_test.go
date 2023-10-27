package frozen_test

import (
	"testing"

	"github.com/arr-ai/hash"

	. "github.com/arr-ai/frozen"
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
	var b SetBuilder[intWithBadHash]
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

	var b SetBuilder[intWithBadHash]
	b.Add(intWithBadHash(100))
	b.Add(intWithBadHash(200))
	b.Remove(intWithBadHash(100))
	test.True(t, b.Has(intWithBadHash(200)))
}
