package lazy

import (
	"testing"

	"github.com/marcelocantos/frozen"
	"github.com/stretchr/testify/assert"
)

func TestSetWhereEmpty(t *testing.T) {
	t.Parallel()

	f := Frozen(frozen.Set{})

	assert.True(t, f.IsEmpty())
	assertFastIsEmpty(t, f)
	assert.Equal(t, 0, f.Count())
	assertFastCountEqual(t, 0, f)
	assert.Equal(t, 0, f.CountUpTo(0))
	assert.Equal(t, 0, f.CountUpTo(1))
	assertFastCountUpToEqual(t, 0, f, 0)
	assertFastCountUpToEqual(t, 0, f, 1)
	assert.True(t, f.Equal(Frozen(frozen.Set{})))
	assert.True(t, f.EqualSet(Frozen(frozen.Set{})))
	assert.False(t, f.EqualSet(Frozen(frozen.NewSet(1))))
	assert.NotEqual(t, 0, f.Hash(0))
	assert.False(t, f.Has(3))
	assertFastNotHas(t, f, 3)
	assert.True(t, f.IsSubsetOf(Frozen(frozen.Set{})))
	assert.False(t, f.Range().Next())
	assertFastIsEmpty(t, f.Where(func(_ interface{}) bool { return true }))
	assertFastNotIsEmpty(t, f.With(2))
	assertFastIsEmpty(t, f.Without(2))
	assertFastIsEmpty(t, f.With(2).Without(2))
	assertFastIsEmpty(t, f.Map(func(_ interface{}) interface{} { return 42 }))
	assert.True(t, f.Union(Frozen(frozen.Set{})).IsEmpty())
	assertFastIsEmpty(t, f.Intersection(Frozen(frozen.NewSet(1, 2, 3))))
	assertFastIsEmpty(t, f.Intersection(Frozen(frozen.NewSet(1, 2, 3))))
	assertFastIsEmpty(t, f.Powerset())
}
