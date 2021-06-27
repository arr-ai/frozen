package frozen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

func TestKeyValueString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "1:2", KV(1, 2).String())
}

func TestMapEmpty(t *testing.T) {
	t.Parallel()

	var m Map
	assert.True(t, m.IsEmpty())
	m = m.With(1, 2)
	assert.False(t, m.IsEmpty())
	m = m.Without(NewSet(1))
	assert.True(t, m.IsEmpty())
	assert.Panics(t, func() { m.MustGet(1) })
}

func TestMapWithWithout(t *testing.T) {
	t.Parallel()

	var m Map
	for i := 0; i < 15; i++ {
		m = m.With(i, i*i)
	}
	assert.NotPanics(t, func() { m.MustGet(14) })
	assert.Panics(t, func() { m.MustGet(15) })
	assert.Equal(t, 15, m.Count())
	for i := 0; i < 10; i++ {
		assert.Equal(t, i*i, m.MustGet(i))
	}

	m = m.Without(Iota2(5, 10))

	assert.Equal(t, 10, m.Count())
	for i := 0; i < 15; i++ {
		switch {
		case i < 5:
			assertMapHas(t, m, i, i*i)
		case i < 10:
			assertMapNotHas(t, m, i)
		default:
			assertMapHas(t, m, i, i*i)
		}
	}
}

// Special case for a bug found when testing Nest.
func TestMapNestBug(t *testing.T) {
	t.Parallel()

	// Original data:
	//   (aa: {(a: 11), (a: 10)}):{(c: 3)}
	//   (aa: {(a: 11), (a: 10)}):{(c: 1)}
	a := KV(
		NewMap(
			KV("aa", NewSet(
				NewMap(KV("a", 11)),
				NewMap(KV("a", 10)),
			)),
		),
		NewSet(NewMap(KV("c", 3))),
	)
	require.Contains(t,
		[]string{
			"(aa: {(a: 10), (a: 11)}):{(c: 3)}",
			"(aa: {(a: 11), (a: 10)}):{(c: 3)}",
		}, a.String())
	b := KV(
		NewMap(
			KV("aa", NewSet(
				NewMap(KV("a", 11)),
				NewMap(KV("a", 10)),
			)),
		),
		NewSet(NewMap(KV("c", 1))),
	)
	require.Contains(t,
		[]string{
			"(aa: {(a: 10), (a: 11)}):{(c: 1)}",
			"(aa: {(a: 11), (a: 10)}):{(c: 1)}",
		}, b.String())
	assert.Equal(t, a.Hash(0) == b.Hash(0), KeyEqual(a, b))

	// The bug actually caused an endless loop, but there's not way to assert
	// for that
	NewMap(a).Update(NewMap(b))
}

func TestMapRedundantWithWithout(t *testing.T) {
	t.Parallel()

	var m Map
	for i := 0; i < 35; i++ {
		m = m.With(i, i*i)
	}
	for i := 10; i < 25; i++ {
		m = m.Without(Iota2(10, 25))
	}
	for i := 5; i < 15; i++ {
		m = m.With(i, i*i*i)
	}
	for i := 20; i < 30; i++ {
		m = m.Without(Iota2(20, 30))
	}

	for i := 0; i < 35; i++ {
		switch {
		case i < 5:
			assertMapHas(t, m, i, i*i)
		case i < 15:
			assertMapHas(t, m, i, i*i*i)
		case i < 30:
			assertMapNotHas(t, m, i)
		default:
			assertMapHas(t, m, i, i*i)
		}
	}
}

func TestNewMapFromGoMap(t *testing.T) {
	t.Parallel()

	N := hugeCollectionSize()
	m := make(map[interface{}]interface{}, N)
	for i := 0; i < N; i++ {
		m[i] = i * i
	}

	fm := NewMapFromGoMap(m)
	expected := NewMapFromKeys(Iota(N), func(k interface{}) interface{} {
		return k.(int) * k.(int)
	})
	assert.True(t, fm.Equal(expected))
}

func TestMapGetElse(t *testing.T) {
	t.Parallel()

	var m Map
	assert.Equal(t, 10, m.GetElse(1, 10))
	assert.Equal(t, 30, m.GetElse(3, 30))
	m = m.With(1, 2)
	assert.Equal(t, 2, m.GetElse(1, 10))
	assert.Equal(t, 30, m.GetElse(3, 30))
	m = m.With(3, 4)
	assert.Equal(t, 2, m.GetElse(1, 10))
	assert.Equal(t, 4, m.GetElse(3, 30))
	m = m.Without(NewSet(1))
	assert.Equal(t, 10, m.GetElse(1, 10))
	assert.Equal(t, 4, m.GetElse(3, 30))
	m = m.Without(NewSet(3))
	assert.Equal(t, 10, m.GetElse(1, 10))
	assert.Equal(t, 30, m.GetElse(3, 30))
}

func TestMapGetElseFunc(t *testing.T) {
	t.Parallel()

	val := func(i int) func() interface{} {
		return func() interface{} {
			return i
		}
	}
	var m Map
	assert.Equal(t, 10, m.GetElseFunc(1, val(10)))
	assert.Equal(t, 30, m.GetElseFunc(3, val(30)))
	m = m.With(1, 2)
	assert.Equal(t, 2, m.GetElseFunc(1, val(10)))
	assert.Equal(t, 30, m.GetElseFunc(3, val(30)))
	m = m.With(3, 4)
	assert.Equal(t, 2, m.GetElseFunc(1, val(10)))
	assert.Equal(t, 4, m.GetElseFunc(3, val(30)))
	m = m.Without(NewSet(1))
	assert.Equal(t, 10, m.GetElseFunc(1, val(10)))
	assert.Equal(t, 4, m.GetElseFunc(3, val(30)))
	m = m.Without(NewSet(3))
	assert.Equal(t, 10, m.GetElseFunc(1, val(10)))
	assert.Equal(t, 30, m.GetElseFunc(3, val(30)))
}

func TestMapKeys(t *testing.T) { //nolint:dupl
	t.Parallel()

	var m Map
	test.AssertSetEqual(t, Set{}, m.Keys())
	m = m.With(1, 2)
	test.AssertSetEqual(t, NewSet(1), m.Keys())
	m = m.With(3, 4)
	test.AssertSetEqual(t, NewSet(1, 3), m.Keys())
	m = m.Without(NewSet(1))
	test.AssertSetEqual(t, NewSet(3), m.Keys())
	m = m.Without(NewSet(3))
	test.AssertSetEqual(t, Set{}, m.Keys())
}

func TestMapValues(t *testing.T) { //nolint:dupl
	t.Parallel()

	var m Map
	test.AssertSetEqual(t, Set{}, m.Values())
	m = m.With(1, 2)
	test.AssertSetEqual(t, NewSet(2), m.Values())
	m = m.With(3, 4)
	test.AssertSetEqual(t, NewSet(2, 4), m.Values())
	m = m.Without(NewSet(1))
	test.AssertSetEqual(t, NewSet(4), m.Values())
	m = m.Without(NewSet(3))
	test.AssertSetEqual(t, Set{}, m.Values())
}

func TestMapProject(t *testing.T) {
	t.Parallel()

	var m Map
	assertMapEqual(t, Map{}, m.Project(Set{}))
	assertMapEqual(t, Map{}, m.Project(NewSet(1)))
	assertMapEqual(t, Map{}, m.Project(NewSet(3)))
	assertMapEqual(t, Map{}, m.Project(NewSet(1, 3)))
	m = m.With(1, 2)
	assertMapEqual(t, Map{}, m.Project(Set{}))
	assertMapEqual(t, NewMap(KV(1, 2)), m.Project(NewSet(1)))
	assertMapEqual(t, Map{}, m.Project(NewSet(3)))
	assertMapEqual(t, NewMap(KV(1, 2)), m.Project(NewSet(1, 3)))
	m = m.With(3, 4)
	assertMapEqual(t, Map{}, m.Project(Set{}))
	assertMapEqual(t, NewMap(KV(1, 2)), m.Project(NewSet(1)))
	assertMapEqual(t, NewMap(KV(3, 4)), m.Project(NewSet(3)))
	assertMapEqual(t, NewMap(KV(1, 2), KV(3, 4)), m.Project(NewSet(1, 3)))
	m = m.Without(NewSet(1))
	assertMapEqual(t, Map{}, m.Project(Set{}))
	assertMapEqual(t, Map{}, m.Project(NewSet(1)))
	assertMapEqual(t, NewMap(KV(3, 4)), m.Project(NewSet(3)))
	assertMapEqual(t, NewMap(KV(3, 4)), m.Project(NewSet(1, 3)))
	m = m.Without(NewSet(3))
	assertMapEqual(t, Map{}, m.Project(Set{}))
	assertMapEqual(t, Map{}, m.Project(NewSet(1)))
	assertMapEqual(t, Map{}, m.Project(NewSet(3)))
	assertMapEqual(t, Map{}, m.Project(NewSet(1, 3)))
}

func TestMapWhere(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 2), KV(3, 4), KV(4, 5), KV(6, 7))
	assertMapEqual(t, NewMap(), m.Where(func(_, _ interface{}) bool { return false }))
	assertMapEqual(t, m, m.Where(func(_, _ interface{}) bool { return true }))
	assertMapEqual(t,
		NewMap(KV(4, 5), KV(6, 7)),
		m.Where(func(k, _ interface{}) bool { return k.(int)%2 == 0 }),
	)
	assertMapEqual(t,
		NewMap(KV(1, 2), KV(3, 4)),
		m.Where(func(_, v interface{}) bool { return v.(int)%2 == 0 }),
	)
}

func TestMapMap(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 2), KV(3, 4), KV(4, 5), KV(6, 7))
	assertMapEqual(t,
		NewMap(KV(1, 3), KV(3, 7), KV(4, 9), KV(6, 13)),
		m.Map(func(k, v interface{}) interface{} { return k.(int) + v.(int) }),
	)
}

func TestMapReduce(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 2), KV(3, 4), KV(4, 5), KV(6, 7))
	assert.Equal(t,
		1*2+3*4+4*5+6*7,
		m.Reduce(func(acc, k, v interface{}) interface{} { return acc.(int) + k.(int)*v.(int) }, 0),
	)
}

func TestMapUpdate(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(3, 4), KV(4, 5), KV(1, 2))
	n := NewMap(KV(3, 4), KV(4, 7), KV(6, 7))
	assertMapEqual(t, NewMap(KV(1, 2), KV(3, 4), KV(4, 7), KV(6, 7)), m.Update(n))
	lotsa := Iota(5).Powerset()
	plus := func(n int) func(interface{}) interface{} {
		return func(key interface{}) interface{} { return n + key.(int) }
	}
	for i := lotsa.Range(); i.Next(); {
		s := i.Value().(Set)
		a := NewMapFromKeys(s, plus(0))
		for j := lotsa.Range(); j.Next(); {
			u := j.Value().(Set)
			b := NewMapFromKeys(u, plus(10))
			actual := a.Update(b)
			expected := NewMapFromKeys(s.Union(u), func(key interface{}) interface{} {
				if u.Has(key) {
					return 10 + key.(int)
				}
				return key
			})
			if !assertMapEqual(t, expected, actual) {
				// log.Print("a:           ", a)
				// log.Print("b:           ", b)
				// log.Print("a.Update(b): ", actual)
				// a = NewMapFromKeys(s, plus(0))
				// b = NewMapFromKeys(u, plus(10))
				// actual = a.Update(b)
				// a.Update(b)
				return
			}
		}
	}
}

func TestMapHashAndEqual(t *testing.T) {
	t.Parallel()

	maps := []Map{
		{},
		NewMap(KV(1, 2)),
		NewMap(KV(1, 3)),
		NewMap(KV(3, 4)),
		NewMap(KV(3, 5)),
		NewMap(KV(1, 2), KV(3, 4)),
		NewMap(KV(1, 2), KV(3, 5)),
		NewMap(KV(1, 3), KV(3, 4)),
		NewMap(KV(1, 3), KV(3, 5)),
		NewMap(KV(NewSet(1, 2), NewSet(3, 4))),
		NewMap(KV(NewSet(1, 2), NewSet(3, 5))),
	}
	for i, a := range maps {
		for j, b := range maps {
			assert.Equal(t, i == j, a.Equal(b), "i=%v j=%v a=%v b=%v", i, j, a, b)
			assert.Equal(t, i == j, a.Hash(0) == b.Hash(0),
				"i=%v j=%v a=%v b=%v a.Hash()=%v b.Hash()=%v", i, j, a, b, a.Hash(0), b.Hash(0))
		}
		assert.False(t, a.Equal(42))
	}
}

func TestMapString(t *testing.T) {
	t.Parallel()

	var m Map
	assert.Equal(t, "()", m.String())
	m = m.With(1, 2)
	assert.Equal(t, "(1: 2)", m.String())
	m = m.With(3, 4)
	assert.Contains(t, []string{"(1: 2, 3: 4)", "(3: 4, 1: 2)"}, m.String())
}

func TestMapRange(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 2), KV(3, 4), KV(4, 5), KV(6, 7))
	output := map[int]int{}
	for i := m.Range(); i.Next(); {
		k, v := i.Entry()
		output[k.(int)] = v.(int)
	}

	assert.Equal(t, map[int]int{1: 2, 3: 4, 4: 5, 6: 7}, output)
}

func TestMapMergeSameValueType(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 1), KV(2, 2), KV(3, 3))
	n := NewMap(KV(1, 2), KV(2, 3), KV(4, 4))
	resolve := func(key, a, b interface{}) interface{} {
		return key.(int) + a.(int) + b.(int)
	}
	expected := NewMap(KV(1, 4), KV(2, 7), KV(3, 3), KV(4, 4))
	result := m.Merge(n, resolve)
	assert.True(t, expected.Equal(result))
}

func TestMapMergeDifferentValueType(t *testing.T) {
	t.Parallel()

	m := NewMap(KV(1, 1), KV(2, 2), KV(3, 5))
	n := NewMap(KV(1, "A"), KV(2, 'K'), KV(3, byte(1)), KV(4, 4))
	resolve := func(key, a, b interface{}) interface{} {
		switch v := b.(type) {
		case string:
			return v
		case rune:
			return string(v)
		case []byte:
			return string(v)
		default:
			return a
		}
	}
	expected := NewMap(KV(1, "A"), KV(2, "K"), KV(3, 5), KV(4, 4))

	assert.True(t, expected.Equal(m.Merge(n, resolve)))
}

func TestMapMergeEmptyMap(t *testing.T) {
	t.Parallel()

	empty := NewMap()
	nonEmpty := NewMap(KV("doesn't", "matter"), KV(1, 2), KV(2, 3))

	assert.True(t, nonEmpty.Equal(empty.Merge(nonEmpty, func(key, a, b interface{}) interface{} { return a })))
	assert.True(t, nonEmpty.Equal(nonEmpty.Merge(empty, func(key, a, b interface{}) interface{} { return a })))
}

var prepopMapInt = memoizePrepop(func(n int) interface{} {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInt(b *testing.B, n int) {
	b.Helper()

	m := prepopMapInt(n).(map[int]int)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInt0(b *testing.B) {
	benchmarkInsertMapInt(b, 0)
}

func BenchmarkInsertMapInt1k(b *testing.B) {
	benchmarkInsertMapInt(b, 1<<10)
}

func BenchmarkInsertMapInt1M(b *testing.B) {
	benchmarkInsertMapInt(b, 1<<20)
}

var prepopMapInterface = memoizePrepop(func(n int) interface{} {
	m := make(map[interface{}]interface{}, n)
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInterface(b *testing.B, n int) {
	b.Helper()

	m := prepopMapInterface(n).(map[interface{}]interface{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInterface0(b *testing.B) {
	benchmarkInsertMapInterface(b, 0)
}

func BenchmarkInsertMapInterface1k(b *testing.B) {
	benchmarkInsertMapInterface(b, 1<<10)
}

func BenchmarkInsertMapInterface1M(b *testing.B) {
	benchmarkInsertMapInterface(b, 1<<20)
}

var prepopFrozenMap = memoizePrepop(func(n int) interface{} {
	var m Map
	for i := 0; i < n; i++ {
		m = m.With(i, i*i)
	}
	return m
})

func benchmarkInsertFrozenMap(b *testing.B, n int) {
	b.Helper()

	m := prepopFrozenMap(n).(Map)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m.With(i, i*i)
	}
}

func BenchmarkInsertFrozenMap0(b *testing.B) {
	benchmarkInsertFrozenMap(b, 0)
}

func BenchmarkInsertFrozenMap1k(b *testing.B) {
	benchmarkInsertFrozenMap(b, 1<<10)
}

func BenchmarkInsertFrozenMap1M(b *testing.B) {
	benchmarkInsertFrozenMap(b, 1<<20)
}

func benchmarkMergeFrozenMap(b *testing.B, limit int) {
	b.Helper()

	plus := func(_, a, b interface{}) interface{} { return a.(int) + b.(int) }
	mapToTest := NewMapFromKeys(Iota(limit), func(key interface{}) interface{} {
		return key
	})
	for i := 0; i < b.N; i++ {
		mapToTest.Merge(mapToTest, plus)
	}
}

func BenchmarkMergeFrozenMap10(b *testing.B) {
	benchmarkMergeFrozenMap(b, 10)
}

func BenchmarkMergeFrozenMap100(b *testing.B) {
	benchmarkMergeFrozenMap(b, 100)
}

func BenchmarkMergeFrozenMap1k(b *testing.B) {
	benchmarkMergeFrozenMap(b, 1000)
}

func BenchmarkMergeFrozenMap100k(b *testing.B) {
	benchmarkMergeFrozenMap(b, 100000)
}

func BenchmarkMergeFrozenMap1M(b *testing.B) {
	benchmarkMergeFrozenMap(b, 1000000)
}
