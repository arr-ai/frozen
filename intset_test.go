package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntSetEmpty(t *testing.T) {
	t.Parallel()

	var s IntSet
	assert.True(t, s.IsEmpty())
}

func TestNewIntSet(t *testing.T) {
	t.Parallel()

	n := 1000
	set := intSetIota(n)
	for i := 0; i < n; i++ {
		assert.True(t, set.Has(i), i)
	}
}

func TestIntSetIter(t *testing.T) {
	t.Parallel()

	n := 1000
	container := Iota(n)
	set := intSetIota(1000)
	for i := set.Range(); i.Next(); {
		assert.True(t, container.Has(i.Value()), i.Value())
	}
}

func intSetIota(n int) IntSet {
	arr := make([]int, 0, n)
	for i := 0; i < n; i++ {
		arr = append(arr, i)
	}
	return NewIntSet(arr...)
}
