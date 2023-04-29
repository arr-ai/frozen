//nolint:deadcode
package lazy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arr-ai/frozen"
	. "github.com/arr-ai/frozen/lazy"
)

func extraArgs(msgAndArgs []any, msg string, args ...any) []any {
	if len(msgAndArgs) == 0 {
		return append([]any{msg}, args...)
	}
	return append(append([]any{msgAndArgs[0].(string) + msg}, msgAndArgs[1:]...), args...)
}

func assertFastIsEmpty(t *testing.T, a Set) bool {
	t.Helper()

	empty, ok := a.FastIsEmpty()
	return assert.True(t, ok) && assert.True(t, empty)
}

func assertFastNotIsEmpty(t *testing.T, a Set) bool {
	t.Helper()

	empty, ok := a.FastIsEmpty()
	return assert.True(t, ok) && assert.False(t, empty)
}

func assertEqualSet(t *testing.T, expected, s Set, msgAndArgs ...any) bool {
	t.Helper()

	return assert.True(t, expected.EqualSet(s),
		extraArgs(msgAndArgs, "\nexpected=%v\nactual  =%v", expected.Freeze(), s.Freeze())...)
}

func assertNotEqualSet(t *testing.T, expected, s Set, msgAndArgs ...any) bool {
	t.Helper()

	return assert.False(t, expected.EqualSet(s),
		extraArgs(msgAndArgs, "\nunexpected=%v\nactual    =%v", expected.Freeze(), s.Freeze())...)
}

func assertFastCountEqual(t *testing.T, expected int, a Set) bool {
	t.Helper()

	count, ok := a.FastCount()
	return assert.True(t, ok) && assert.Equal(t, expected, count)
}

func assertFastCountUpToEqual(t *testing.T, expected int, a Set, limit int) bool {
	t.Helper()

	count, ok := a.FastCountUpTo(limit)
	return assert.True(t, ok) && assert.Equal(t, expected, count)
}

func assertFastHas(t *testing.T, a Set, el any) bool {
	t.Helper()

	equal, ok := a.FastHas(el)
	return assert.True(t, ok) && assert.True(t, equal)
}

func assertFastNotHas(t *testing.T, a Set, el any) bool {
	t.Helper()

	equal, ok := a.FastHas(el)
	return assert.True(t, ok) && assert.False(t, equal)
}

func assertRangeEmits(t *testing.T, expected frozen.Set[any], a Set) bool {
	t.Helper()

	var b frozen.SetBuilder[any]
	for i := a.Range(); i.Next(); {
		v := i.Value()
		if assert.False(t, b.Has(v), "duplicate element: %v", v) {
			b.Add(v)
		}
	}
	return assert.True(t, expected.Equal(b.Finish()))
}

func extractInt(i any) int {
	switch x := i.(type) {
	case int:
		return x
	case Set:
		return x.Count()
	default:
		panic("cannot extract int")
	}
}

func assertSetOps(t *testing.T, golden frozen.Set[any], s Set) { //nolint:funlen
	t.Helper()

	count := golden.Count()
	fgolden := Frozen(golden)

	assert.Equal(t, golden.IsEmpty(), s.IsEmpty())

	assert.Equal(t, count, s.Count())

	assertEqualSet(t, fgolden, s)
	assertEqualSet(t, s, fgolden)

	assertNotEqualSet(t, Frozen(golden.With(42)), s)
	assertNotEqualSet(t, s, Frozen(golden.With(42)))

	assert.Equal(t, 0, s.CountUpTo(0))
	if count > 0 {
		assert.Equal(t, count-1, s.CountUpTo(count-1))
	}
	assert.Equal(t, count, s.CountUpTo(count))
	assert.Equal(t, count, s.CountUpTo(count+1))

	assert.Equal(t, golden.Equal(frozen.Set[any]{}), s.Equal(Frozen(frozen.Set[any]{})))
	assert.Equal(t, golden.Equal(frozen.Set[any]{}), s.EqualSet(Frozen(frozen.Set[any]{})))
	assert.False(t, golden.Equal(frozen.NewSet[any](1)), s.EqualSet(Frozen(frozen.NewSet[any](1))))

	assert.NotEqual(t, 0, s.Hash(0))

	for i := 0; i < 10; i++ {
		assert.Equal(t, golden.Has(i), s.Has(i), "i=%v", i)
	}

	assert.True(t, s.IsSubsetOf(fgolden))
	assert.True(t, s.IsSubsetOf(fgolden.With(42)))
	assert.True(t, fgolden.IsSubsetOf(s))
	assert.False(t, fgolden.With(42).IsSubsetOf(s))

	assertRangeEmits(t, golden, s)
	for i, pred := range []func(any) bool{
		func(any) bool { return false },
		func(any) bool { return true },
		func(i any) bool { return extractInt(i)%2 == 0 },
		func(i any) bool { return extractInt(i) < 3 },
	} {
		expected := Frozen(golden.Where(pred))
		actual := s.Where(pred)
		assertEqualSet(t, expected, actual, "i=%v", i)
	}

	assert.False(t, s.With(2).IsEmpty())

	assertEqualSet(t, Frozen(golden.Without(2)), s.Without(2))
	assertEqualSet(t, Frozen(golden.With(2).Without(2)), s.With(2).Without(2))
	assertEqualSet(t, Frozen(golden.Without(42)), s.Without(42))
	assertEqualSet(t, Frozen(golden.With(42).Without(42)), s.With(42).Without(42))

	for i, m := range []func(any) any{
		func(any) any { return 42 },
		func(i any) any { return i },
		func(i any) any { return 2 * extractInt(i) },
		func(i any) any { return extractInt(i) / 2 },
	} {
		assertEqualSet(t, Frozen(frozen.SetMap(golden, m)), s.Map(m), "i=%v", i)
	}

	for i, u := range []frozen.Set[any]{
		{},
		frozen.NewSet[any](1, 2, 3),
		frozen.NewSet[any](1, 2, 3, 4),
		frozen.NewSet[any](4, 5),
	} {
		assertEqualSet(t, Frozen(golden.Union(u)), s.Union(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.Intersection(u)), s.Intersection(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.Difference(u)), s.Difference(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.SymmetricDifference(u)), s.SymmetricDifference(Frozen(u)), "i=%v u=%v", i, u)
	}

	assert.Equal(t, 1<<uint(golden.Count()), s.Powerset().Count())
}
