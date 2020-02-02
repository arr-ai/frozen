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

	const N = 1000
	arr := make([]int, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, i)
	}
	set := NewIntSet(arr...)
	for i := 0; i < N; i++ {
		assert.True(t, set.Has(i), i)
	}
}
