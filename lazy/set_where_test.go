package lazy_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
	"github.com/arr-ai/frozen/lazy"
)

func TestSetWhereEmpty(t *testing.T) {
	t.Parallel()

	f := lazy.Frozen(frozen.Set[any]{})

	test.True(t, f.IsEmpty())
	assertFastIsEmpty(t, f)
	test.Equal(t, 0, f.Count())
	assertFastCountEqual(t, 0, f)
	test.Equal(t, 0, f.CountUpTo(0))
	test.Equal(t, 0, f.CountUpTo(1))
	assertFastCountUpToEqual(t, 0, f, 0)
	assertFastCountUpToEqual(t, 0, f, 1)
	test.True(t, f.Equal(lazy.Frozen(frozen.Set[any]{})))
	test.True(t, f.EqualSet(lazy.Frozen(frozen.Set[any]{})))
	test.False(t, f.EqualSet(lazy.Frozen(frozen.NewSet[any](1))))
	test.NotEqual(t, 0, f.Hash(0))
	test.False(t, f.Has(3))
	assertFastNotHas(t, f, 3)
	test.True(t, f.IsSubsetOf(lazy.Frozen(frozen.Set[any]{})))
	test.False(t, f.Range().Next())
	assertFastIsEmpty(t, f.Where(func(any) bool { return true }))
	assertFastNotIsEmpty(t, f.With(2))
	assertFastIsEmpty(t, f.Without(2))
	assertFastIsEmpty(t, f.With(2).Without(2))
	assertFastIsEmpty(t, f.Map(func(any) any { return 42 }))
	test.True(t, f.Union(lazy.Frozen(frozen.Set[any]{})).IsEmpty())
	assertFastIsEmpty(t, f.Intersection(lazy.Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastIsEmpty(t, f.Intersection(lazy.Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastIsEmpty(t, f.Powerset())
}
