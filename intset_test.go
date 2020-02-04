package frozen

import (
	"testing"
	"sort"

	"github.com/stretchr/testify/assert"
)

func TestIntSetEmpty(t *testing.T) {
	t.Parallel()

	var s IntSet
	assert.True(t, s.IsEmpty())
	assert.False(t, s.Has(0))
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
	container := makeIntArray(0, n, 1)
	set := intSetIota(n)
	arr := make([]int, 0, n)
	for i := set.Range(); i.Next(); {
		arr = append(arr, i.Value())
	}
	sort.Ints(arr)
	assert.Equal(t, container, arr)
}

func TestIntSetHas(t *testing.T) {
	t.Parallel()

	set := intSetIota(100)
	assert.True(t, set.Has(99))
	assert.False(t, set.Has(100))
}

func TestIntSetWith(t *testing.T) {
	t.Parallel()

	n := 1000
	set := intSetIota(n)
	arr := makeIntArray(n, 2*n, 1)
	for i := n; i < 2*n; i++ {
		assert.False(t, set.Has(i))
	}

	set = set.With(arr...)

	for i := n; i < 2*n; i++ {
		assert.True(t, set.Has(i))
	}
	assert.Equal(t, 2*n, set.count)
}

func TestIntSetWithout(t *testing.T) {
	t.Parallel()

	n := 1024
	set := intSetIota(n)
	for i := 0; i < n; i++ {
		assert.True(t, set.Has(i))
	}

	set = set.Without(makeIntArray(n/2, n, 1)...)
	assert.Equal(t, n/2, set.count)
	for i := n / 2; i < n; i++ {
		assert.False(t, set.Has(i), i)
	}

	for i := set.data.Range(); i.Next(); {
		assert.NotEqual(t, emptyBlock, i.Value().(cellBlock), i.Key())
	}
}

func intSetIota(n int) IntSet {
	arr := make([]int, 0, n)
	for i := 0; i < n; i++ {
		arr = append(arr, i)
	}
	return NewIntSet(arr...)
}

func makeIntArray(start, stop, step int) []int {
	arr := generateSortedIntArray(start, stop, step)
	intArr := make([]int, 0, len(arr))
	for _, i := range arr {
		intArr = append(intArr, i.(int))
	}
	return intArr
}
