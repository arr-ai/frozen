package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
	testset "github.com/arr-ai/frozen/internal/pkg/test/set"
)

type mapIntInt = frozen.Map[int, int]

func TestKeyValueString(t *testing.T) {
	t.Parallel()

	test.Equal(t, "1:2", frozen.KV(1, 2).String())
}

func TestMapEmpty(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	test.True(t, m.IsEmpty())
	m = m.With(1, 2)
	test.False(t, m.IsEmpty())
	m = m.Without(1)
	test.True(t, m.IsEmpty())
	test.Panic(t, func() { m.MustGet(1) })
}

func TestMap2(t *testing.T) {
	t.Parallel()

	a := frozen.NewMap(frozen.KV(1, 10), frozen.KV(2, 20))
	b := frozen.NewMap(frozen.KV(2, 20), frozen.KV(1, 10))
	test.True(t, a.Equal(b))

	var mb frozen.MapBuilder[frozen.Map[int, int], int]
	mb.Put(a, 100)
	mb.Put(a, 101)
	test.Equal(t, 1, mb.Finish().Count())
}

func TestMapWithWithout(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	for i := 0; i < 15; i++ {
		m = m.With(i, i*i)
	}
	test.NoPanic(t, func() { m.MustGet(14) })
	test.Panic(t, func() { m.MustGet(15) })
	test.Equal(t, 15, m.Count())
	for i := 0; i < 10; i++ {
		test.Equal(t, i*i, m.MustGet(i))
	}

	m = mapWithoutKeys(m, frozen.Iota2(5, 10))

	test.Equal(t, 10, m.Count())
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

func TestMapHas(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap[int, int]()
	test.False(t, m.Has(0))

	m = frozen.NewMap(frozen.KV(0, 1))
	test.True(t, m.Has(0))
	test.False(t, m.Has(1))
}

func TestMapGet(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(
		frozen.KV(0, 1),
		frozen.KV(1, 1),
		frozen.KV(2, 2),
		frozen.KV(3, 3),
		frozen.KV(4, 5),
	)
	test.Equal(t, 1, m.MustGet(0))
	test.Equal(t, 1, m.MustGet(1))
	test.Equal(t, 2, m.MustGet(2))
	test.Equal(t, 3, m.MustGet(3))
	test.Equal(t, 5, m.MustGet(4))
}

func mapWithoutKeys[K, V any](m frozen.Map[K, V], keys frozen.Set[K]) frozen.Map[K, V] {
	for r := keys.Range(); r.Next(); {
		m = m.Without(r.Value())
	}
	return m
}

//nolint:cyclop
func TestMapRedundantWithWithout(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	for i := 0; i < 35; i++ {
		m = m.With(i, i*i)
	}
	for i := 10; i < 25; i++ {
		m = mapWithoutKeys(m, frozen.Iota2(10, 25))
	}
	for i := 5; i < 15; i++ {
		m = m.With(i, i*i*i)
		if !assertMapHas(t, m, i, i*i*i) {
			return
		}
	}
	for i := 20; i < 30; i++ {
		m = mapWithoutKeys(m, frozen.Iota2(20, 30))
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

func TestMapNewMapFromGoMap(t *testing.T) {
	t.Parallel()

	N := hugeCollectionSize()
	m := make(map[int]int, N)
	for i := 0; i < N; i++ {
		m[i] = i * i
	}

	fm := frozen.NewMapFromGoMap(m)
	expected := frozen.NewMapFromKeys(frozen.Iota(N), func(k int) int {
		return k * k
	})
	test.True(t, fm.Equal(expected))
}

func TestMapGetElse(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	test.Equal(t, 10, m.GetElse(1, 10))
	test.Equal(t, 30, m.GetElse(3, 30))
	m = m.With(1, 2)
	test.Equal(t, 2, m.GetElse(1, 10))
	test.Equal(t, 30, m.GetElse(3, 30))
	m = m.With(3, 4)
	test.Equal(t, 2, m.GetElse(1, 10))
	test.Equal(t, 4, m.GetElse(3, 30))
	m = m.Without(1)
	test.Equal(t, 10, m.GetElse(1, 10))
	test.Equal(t, 4, m.GetElse(3, 30))
	m = m.Without(3)
	test.Equal(t, 10, m.GetElse(1, 10))
	test.Equal(t, 30, m.GetElse(3, 30))
}

func TestMapGetElseFunc(t *testing.T) {
	t.Parallel()

	val := func(i int) func() int {
		return func() int {
			return i
		}
	}
	var m mapIntInt
	test.Equal(t, 10, m.GetElseFunc(1, val(10)))
	test.Equal(t, 30, m.GetElseFunc(3, val(30)))
	m = m.With(1, 2)
	test.Equal(t, 2, m.GetElseFunc(1, val(10)))
	test.Equal(t, 30, m.GetElseFunc(3, val(30)))
	m = m.With(3, 4)
	test.Equal(t, 2, m.GetElseFunc(1, val(10)))
	test.Equal(t, 4, m.GetElseFunc(3, val(30)))
	m = m.Without(1)
	test.Equal(t, 10, m.GetElseFunc(1, val(10)))
	test.Equal(t, 4, m.GetElseFunc(3, val(30)))
	m = m.Without(3)
	test.Equal(t, 10, m.GetElseFunc(1, val(10)))
	test.Equal(t, 30, m.GetElseFunc(3, val(30)))
}

func TestMapKeys(t *testing.T) { //nolint:dupl
	t.Parallel()

	var m mapIntInt
	testset.AssertSetEqual(t, frozen.Set[int]{}, m.Keys())
	m = m.With(1, 2)
	testset.AssertSetEqual(t, frozen.NewSet(1), m.Keys())
	m = m.With(3, 4)
	testset.AssertSetEqual(t, frozen.NewSet(1, 3), m.Keys())
	m = m.Without(1)
	testset.AssertSetEqual(t, frozen.NewSet(3), m.Keys())
	m = m.Without(3)
	testset.AssertSetEqual(t, frozen.Set[int]{}, m.Keys())
}

func TestMapValues(t *testing.T) { //nolint:dupl
	t.Parallel()

	var m mapIntInt
	testset.AssertSetEqual(t, frozen.Set[int]{}, m.Values())
	m = m.With(1, 2)
	testset.AssertSetEqual(t, frozen.NewSet(2), m.Values())
	m = m.With(3, 4)
	testset.AssertSetEqual(t, frozen.NewSet(2, 4), m.Values())
	m = m.Without(1)
	testset.AssertSetEqual(t, frozen.NewSet(4), m.Values())
	m = m.Without(3)
	testset.AssertSetEqual(t, frozen.Set[int]{}, m.Values())
}

func TestMapProject(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	assertMapEqual(t, mapIntInt{}, m.Project())
	assertMapEqual(t, mapIntInt{}, m.Project(1))
	assertMapEqual(t, mapIntInt{}, m.Project(3))
	assertMapEqual(t, mapIntInt{}, m.Project(1, 3))
	m = m.With(1, 2)
	assertMapEqual(t, mapIntInt{}, m.Project())
	assertMapEqual(t, frozen.NewMap(frozen.KV(1, 2)), m.Project(1))
	assertMapEqual(t, mapIntInt{}, m.Project(3))
	assertMapEqual(t, frozen.NewMap(frozen.KV(1, 2)), m.Project(1, 3))
	m = m.With(3, 4)
	assertMapEqual(t, mapIntInt{}, m.Project())
	assertMapEqual(t, frozen.NewMap(frozen.KV(1, 2)), m.Project(1))
	assertMapEqual(t, frozen.NewMap(frozen.KV(3, 4)), m.Project(3))
	assertMapEqual(t, frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4)), m.Project(1, 3))
	m = m.Without(1)
	assertMapEqual(t, mapIntInt{}, m.Project())
	assertMapEqual(t, mapIntInt{}, m.Project(1))
	assertMapEqual(t, frozen.NewMap(frozen.KV(3, 4)), m.Project(3))
	assertMapEqual(t, frozen.NewMap(frozen.KV(3, 4)), m.Project(1, 3))
	m = m.Without(3)
	assertMapEqual(t, mapIntInt{}, m.Project())
	assertMapEqual(t, mapIntInt{}, m.Project(1))
	assertMapEqual(t, mapIntInt{}, m.Project(3))
	assertMapEqual(t, mapIntInt{}, m.Project(1, 3))
}

func TestMapWhere(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4), frozen.KV(4, 5), frozen.KV(6, 7))
	assertMapEqual(t, frozen.NewMap[int, int](), m.Where(func(_, _ int) bool { return false }))
	assertMapEqual(t, m, m.Where(func(_, _ int) bool { return true }))
	assertMapEqual(t,
		frozen.NewMap(frozen.KV(4, 5), frozen.KV(6, 7)),
		m.Where(func(k, _ int) bool { return k%2 == 0 }),
	)
	assertMapEqual(t,
		frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4)),
		m.Where(func(_, v int) bool { return v%2 == 0 }),
	)
}

func TestMapMap(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4), frozen.KV(4, 5), frozen.KV(6, 7))
	assertMapEqual(t,
		frozen.NewMap(frozen.KV(1, 3), frozen.KV(3, 7), frozen.KV(4, 9), frozen.KV(6, 13)),
		frozen.MapMap(m, func(k, v int) int { return k + v }),
	)
}

func TestMapUpdate(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(3, 4), frozen.KV(4, 5), frozen.KV(1, 2))
	n := frozen.NewMap(frozen.KV(3, 4), frozen.KV(4, 7), frozen.KV(6, 7))
	assertMapEqual(t, frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4), frozen.KV(4, 7), frozen.KV(6, 7)), m.Update(n))
	oom := 10
	if testing.Short() {
		oom = 5
	}
	lotsa := frozen.Powerset(frozen.Iota(oom))
	plus := func(n int) func(int) int {
		return func(key int) int { return n + key }
	}
	for i := lotsa.Range(); i.Next(); {
		s := i.Value()
		a := frozen.NewMapFromKeys(s, plus(0))
		for j := lotsa.Range(); j.Next(); {
			u := j.Value()
			b := frozen.NewMapFromKeys(u, plus(100))
			actual := a.Update(b)
			expected := frozen.NewMapFromKeys(s.Union(u), func(key int) int {
				if u.Has(key) {
					return 100 + key
				}
				return key
			})
			if !assertMapEqual(t, expected, actual) {
				// log.Print("a:        ", a)
				// log.Print("b:        ", b)
				// log.Print("expected: ", expected)
				// log.Print("actual:   ", actual)
				// frozen.NewMapFromKeys(s, plus(0))
				// frozen.NewMapFromKeys(u, plus(10))
				// a.Update(b)
				return
			}
		}
	}
}

func TestMapHashAndEqual(t *testing.T) {
	t.Parallel()

	maps := []mapIntInt{
		{},
		frozen.NewMap(frozen.KV(1, 2)),
		frozen.NewMap(frozen.KV(1, 3)),
		frozen.NewMap(frozen.KV(3, 4)),
		frozen.NewMap(frozen.KV(3, 5)),
		frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4)),
		frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 5)),
		frozen.NewMap(frozen.KV(1, 3), frozen.KV(3, 4)),
		frozen.NewMap(frozen.KV(1, 3), frozen.KV(3, 5)),
	}
	for i, a := range maps {
		for j, b := range maps {
			test.Equal(t, i == j, a.Equal(b), "i=%v j=%v a=%v b=%v", i, j, a, b)
			test.Equal(t, i == j, a.Hash(0) == b.Hash(0),
				"i=%v j=%v a=%v b=%v a.Hash()=%v b.Hash()=%v", i, j, a, b, a.Hash(0), b.Hash(0))
		}
	}
}

func TestMapString(t *testing.T) {
	t.Parallel()

	var m mapIntInt
	test.Equal(t, "()", m.String())
	m = m.With(1, 2)
	test.Equal(t, "(1: 2)", m.String())
	m = m.With(3, 4)
	test.True(t, m.String() == "(1: 2, 3: 4)" || m.String() == "(3: 4, 1: 2)")
}

func TestMapRange(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(1, 2), frozen.KV(3, 4), frozen.KV(4, 5), frozen.KV(6, 7))
	output := map[int]int{}
	for i := m.Range(); i.Next(); {
		k, v := i.Entry()
		output[k] = v
	}

	test.Equal(t, map[int]int{1: 2, 3: 4, 4: 5, 6: 7}, output)
}

func TestMapMergeSameValueType(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(1, 1), frozen.KV(2, 2), frozen.KV(3, 3))
	n := frozen.NewMap(frozen.KV(1, 2), frozen.KV(2, 3), frozen.KV(4, 4))
	resolve := func(key, a, b int) int {
		return key + a + b
	}
	expected := frozen.NewMap(frozen.KV(1, 4), frozen.KV(2, 7), frozen.KV(3, 3), frozen.KV(4, 4))
	result := m.Merge(n, resolve)
	test.True(t, expected.Equal(result))
}

func TestMapMergeEmptyMap(t *testing.T) {
	t.Parallel()

	empty := frozen.NewMap[int, int]()
	nonEmpty := frozen.NewMap(frozen.KV(1, 2), frozen.KV(2, 3))

	test.True(t, nonEmpty.Equal(empty.Merge(nonEmpty, func(key, a, b int) int { return a })))
	test.True(t, nonEmpty.Equal(nonEmpty.Merge(empty, func(key, a, b int) int { return a })))
}

func TestMapToGoMapEmpty(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap[int, int]()
	expected := map[int]int{}
	actual := frozen.MapToGoMap(m)
	test.Equal(t, expected, actual)
}

func TestMapToGoMapSmall(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(frozen.KV(1, 1), frozen.KV(2, 2), frozen.KV(3, 3))
	expected := map[int]int{1: 1, 2: 2, 3: 3}
	actual := frozen.MapToGoMap(m)
	test.Equal(t, expected, actual)
}

func TestMapToGoMapBig(t *testing.T) {
	t.Parallel()

	if !testing.Short() {
		const N = 1_000_000
		b := frozen.NewMapBuilder[int, int](N)
		expected := make(map[int]int, N)
		for i := 0; i < N; i++ {
			b.Put(i, i*i)
			expected[i] = i * i
		}
		actual := frozen.MapToGoMap(b.Finish())
		test.Equal(t, expected, actual)
	}
}

var prepopMapInt = memoizePrepop(func(n int) map[int]int {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInt(b *testing.B, n int) {
	b.Helper()

	m := prepopMapInt(n)
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

var prepopMapInterface = memoizePrepop(func(n int) map[int]int {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInterface(b *testing.B, n int) {
	b.Helper()

	m := prepopMapInterface(n)
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

var prepopFrozenMap = memoizePrepop(func(n int) mapIntInt {
	var m mapIntInt
	for i := 0; i < n; i++ {
		m = m.With(i, i*i)
	}
	return m
})

func benchmarkInsertFrozenMap(b *testing.B, n int) {
	b.Helper()

	m := prepopFrozenMap(n)
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

	plus := func(_, a, b int) int { return a + b }
	mapToTest := frozen.NewMapFromKeys(frozen.Iota(limit), func(key int) int {
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
