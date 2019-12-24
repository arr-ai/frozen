package frozen

import (
	"testing"

	"github.com/marcelocantos/hash"
	"github.com/stretchr/testify/assert"
)

type intWithBadHash int

var _ hash.Hashable = intWithBadHash(0)

func (i intWithBadHash) Hash(seed uintptr) uintptr {
	return hash.Int(int(i)%100, seed)
}

func TestNodeBadHash(t *testing.T) {
	const N = 100000
	var b SetBuilder
	for i := 0; i < N; i += 10 {
		b.Add(intWithBadHash(i))
	}

	for i := 0; i < N; i += 10 {
		assert.True(t, b.Has(intWithBadHash(i)))
	}
	for i := N; i < 2*N; i += 10 {
		assert.False(t, b.Has(intWithBadHash(i)))
	}
}

func TestNodeRemoveCollider(t *testing.T) {
	var b SetBuilder
	b.Add(intWithBadHash(100))
	b.Add(intWithBadHash(200))
	b.Remove(intWithBadHash(100))
	assert.True(t, b.Has(intWithBadHash(200)))
}

func TestNodeAddNearlyCollider(t *testing.T) {
	var b SetBuilder
	h100 := newHasher(intWithBadHash(100), 0)
	near100 := intWithBadHash(0)
	for ; newHasher(near100, 0) == h100 || newHasher(near100, 0).hash() != h100.hash(); near100++ {
	}
	t.Logf("near100=%o", near100)

	b.Add(intWithBadHash(100))
	b.Add(intWithBadHash(200))
	b.Add(near100)
	assert.True(t, b.Has(intWithBadHash(100)))
	assert.True(t, b.Has(intWithBadHash(200)))
}
