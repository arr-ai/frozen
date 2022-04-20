package lazy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arr-ai/frozen"
	. "github.com/arr-ai/frozen/lazy"
)

func TestSetWhereEmpty(t *testing.T) {
	t.Parallel()

	f := Frozen(frozen.Set[any]{})

	assert.True(t, f.IsEmpty())
	assertFastIsEmpty(t, f)
	assert.Equal(t, 0, f.Count())
	assertFastCountEqual(t, 0, f)
	assert.Equal(t, 0, f.CountUpTo(0))
	assert.Equal(t, 0, f.CountUpTo(1))
	assertFastCountUpToEqual(t, 0, f, 0)
	assertFastCountUpToEqual(t, 0, f, 1)
	assert.True(t, f.Equal(Frozen(frozen.Set[any]{})))
	assert.True(t, f.EqualSet(Frozen(frozen.Set[any]{})))
	assert.False(t, f.EqualSet(Frozen(frozen.NewSet[any](1))))
	assert.NotEqual(t, 0, f.Hash(0))
	assert.False(t, f.Has(3))
	assertFastNotHas(t, f, 3)
	assert.True(t, f.IsSubsetOf(Frozen(frozen.Set[any]{})))
	assert.False(t, f.Range().Next())
	assertFastIsEmpty(t, f.Where(func(_ any) bool { return true }))
	assertFastNotIsEmpty(t, f.With(2))
	assertFastIsEmpty(t, f.Without(2))
	assertFastIsEmpty(t, f.With(2).Without(2))
	assertFastIsEmpty(t, f.Map(func(_ any) any { return 42 }))
	assert.True(t, f.Union(Frozen(frozen.Set[any]{})).IsEmpty())
	assertFastIsEmpty(t, f.Intersection(Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastIsEmpty(t, f.Intersection(Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastIsEmpty(t, f.Powerset())
}
