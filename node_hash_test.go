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
	var sb SetBuilder
	for i := 0; i < N; i += 10 {
		sb.Add(intWithBadHash(i))
	}

	for i := 0; i < N; i += 10 {
		assert.True(t, sb.Has(intWithBadHash(i)))
	}
	for i := N; i < 2*N; i += 10 {
		assert.False(t, sb.Has(intWithBadHash(i)))
	}
}
