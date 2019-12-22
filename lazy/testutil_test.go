//nolint:unused,deadcode,unparam
package lazy

import (
	"testing"

	"github.com/marcelocantos/frozen"
	"github.com/stretchr/testify/assert"
)

func extraArgs(msgAndArgs []interface{}, msg string, args ...interface{}) []interface{} {
	if len(msgAndArgs) == 0 {
		return append([]interface{}{msg}, args...)
	}
	return append(append([]interface{}{msgAndArgs[0].(string) + msg}, msgAndArgs[1:]...), args...)
}

func assertFastIsEmpty(t *testing.T, a Set) bool {
	empty, ok := a.FastIsEmpty()
	return assert.True(t, ok) && assert.True(t, empty)
}

func assertFastNotIsEmpty(t *testing.T, a Set) bool {
	empty, ok := a.FastIsEmpty()
	return assert.True(t, ok) && assert.False(t, empty)
}

func assertEqualSet(t *testing.T, expected, s Set, msgAndArgs ...interface{}) bool {
	return assert.True(t, expected.EqualSet(s),
		extraArgs(msgAndArgs, "\nexpected=%v\nactual  =%v", expected.Freeze(), s.Freeze())...)
}

func assertNotEqualSet(t *testing.T, expected, s Set, msgAndArgs ...interface{}) bool {
	return assert.False(t, expected.EqualSet(s),
		extraArgs(msgAndArgs, "\nunexpected=%v\nactual    =%v", expected.Freeze(), s.Freeze())...)
}

func assertFastCountEqual(t *testing.T, expected int, a Set) bool {
	count, ok := a.FastCount()
	return assert.True(t, ok) && assert.Equal(t, expected, count)
}

func assertFastCountNotEqual(t *testing.T, expected int, a Set) bool {
	count, ok := a.FastCount()
	return assert.True(t, ok) && assert.NotEqual(t, expected, count)
}

func assertFastCountUpToEqual(t *testing.T, expected int, a Set, limit int) bool {
	count, ok := a.FastCountUpTo(limit)
	return assert.True(t, ok) && assert.Equal(t, expected, count)
}

func assertFastCountUpToNotEqual(t *testing.T, expected int, a Set, limit int) bool {
	count, ok := a.FastCountUpTo(limit)
	return assert.True(t, ok) && assert.NotEqual(t, expected, count)
}

func assertFastHas(t *testing.T, a Set, el interface{}) bool {
	equal, ok := a.FastHas(el)
	return assert.True(t, ok) && assert.True(t, equal)
}

func assertFastNotHas(t *testing.T, a Set, el interface{}) bool {
	equal, ok := a.FastHas(el)
	return assert.True(t, ok) && assert.False(t, equal)
}

func assertRangeEmits(t *testing.T, expected frozen.Set, a Set) bool {
	var b frozen.SetBuilder
	for i := a.Range(); i.Next(); {
		v := i.Value()
		if assert.False(t, b.Has(v), "duplicate element: %v", v) {
			b.Add(v)
		}
	}
	return assert.True(t, expected.EqualSet(b.Finish()))
}

func extractInt(i interface{}) int {
	switch x := i.(type) {
	case int:
		return x
	case Set:
		return x.Count()
	default:
		panic("cannot extract int")
	}
}

func assertSetOps(t *testing.T, golden frozen.Set, s Set) { //nolint:funlen
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

	assert.Equal(t, golden.Equal(frozen.Set{}), s.Equal(Frozen(frozen.Set{})))
	assert.Equal(t, golden.EqualSet(frozen.Set{}), s.EqualSet(Frozen(frozen.Set{})))
	assert.False(t, golden.EqualSet(frozen.NewSet(1)), s.EqualSet(Frozen(frozen.NewSet(1))))

	assert.NotEqual(t, 0, s.Hash(0))

	for i := 0; i < 10; i++ {
		assert.Equal(t, golden.Has(i), s.Has(i), "i=%v", i)
	}

	assert.True(t, s.IsSubsetOf(fgolden))
	assert.True(t, s.IsSubsetOf(fgolden.With(42)))
	assert.True(t, fgolden.IsSubsetOf(s))
	assert.False(t, fgolden.With(42).IsSubsetOf(s))

	assertRangeEmits(t, golden, s)
	for i, pred := range []func(interface{}) bool{
		func(_ interface{}) bool { return false },
		func(_ interface{}) bool { return true },
		func(i interface{}) bool { return extractInt(i)%2 == 0 },
		func(i interface{}) bool { return extractInt(i) < 3 },
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

	for i, m := range []func(interface{}) interface{}{
		func(_ interface{}) interface{} { return 42 },
		func(i interface{}) interface{} { return i },
		func(i interface{}) interface{} { return 2 * extractInt(i) },
		func(i interface{}) interface{} { return extractInt(i) / 2 },
	} {
		assertEqualSet(t, Frozen(golden.Map(m)), s.Map(m), "i=%v", i)
	}

	for i, u := range []frozen.Set{
		{},
		frozen.NewSet(1, 2, 3),
		frozen.NewSet(1, 2, 3, 4),
		frozen.NewSet(4, 5),
	} {
		assertEqualSet(t, Frozen(golden.Union(u)), s.Union(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.Intersection(u)), s.Intersection(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.Difference(u)), s.Difference(Frozen(u)), "i=%v", i)
		assertEqualSet(t, Frozen(golden.SymmetricDifference(u)), s.SymmetricDifference(Frozen(u)), "i=%v u=%v", i, u)
	}

	assert.Equal(t, 1<<golden.Count(), s.Powerset().Count())
}
