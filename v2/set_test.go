package frozen_test

import (
	"log"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/arr-ai/frozen/v2"
	"github.com/arr-ai/frozen/v2/pkg/kv"
	"github.com/arr-ai/frozen/v2/internal/pkg/test"
	"github.com/arr-ai/frozen/v2/internal/pkg/tree"
)

func largeIntSet() Set[int] {
	return intSet(0, 10_000)
}

func hugeIntSet() Set[int] {
	return intSet(0, hugeCollectionSize())
}

func TestSetEmpty(t *testing.T) {
	t.Parallel()

	var s Set[int]
	assert.True(t, s.IsEmpty())
	test.AssertSetEqual(t, Set[int]{}, s)
}

func compareElements[T comparable](a, b []T) (aOnly, bOnly []T) {
	ma := map[T]struct{}{}
	for _, e := range a {
		ma[e] = struct{}{}
	}

	mb := map[T]struct{}{}
	for _, e := range b {
		mb[e] = struct{}{}
	}

	for e := range ma {
		if _, has := mb[e]; !has {
			aOnly = append(aOnly, e)
		}
	}

	for e := range mb {
		if _, has := ma[e]; !has {
			bOnly = append(bOnly, e)
		}
	}

	return
}

// faster that assert.ElementsMatch.
func assertSameElements[T comparable](t *testing.T, a, b []T, msgAndArgs ...interface{}) bool {
	t.Helper()

	aOnly, bOnly := compareElements(a, b)
	aOK := assert.Empty(t, aOnly, msgAndArgs...)
	bOK := assert.Empty(t, bOnly, msgAndArgs...)
	return aOK && bOK
}

func requireSameElements[T comparable](t *testing.T, a, b []T, msgAndArgs ...interface{}) {
	t.Helper()

	if !assertSameElements(t, a, b, msgAndArgs...) {
		t.FailNow()
	}
}

func fromStringArr(a []string) []interface{} {
	b := make([]interface{}, 0, len(a))
	for _, i := range a {
		b = append(b, i)
	}
	return b
}

func TestNewSet(t *testing.T) {
	t.Parallel()

	N := 1_000
	if testing.Short() {
		N /= 10
	}
	arr := make([]interface{}, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, i)
	}

	for i := N - 1; i >= 0; i-- {
		expected := arr[i:]
		actual := NewSet(arr[i:]...)
		require.Equal(t, len(expected), actual.Count())
		assertSameElements(t, expected, actual.Elements())
	}
}

func TestNewSetFromStrings(t *testing.T) {
	t.Parallel()

	const N = 256
	arr := make([]string, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, string(rune(i)))
	}

	for i := N - 1; i >= 0; i-- {
		assertSameElements(t, fromStringArr(arr[i:]), NewSet(fromStringArr(arr[i:])...).Elements())
	}
}

func TestSetWith(t *testing.T) {
	t.Parallel()

	var s Set[int]
	arr := []int{}
	n := 1_000
	if testing.Short() {
		n /= 10
	}
	for i := 0; i < n; i++ {
		expected := NewSet(arr...)
		if !test.AssertSetEqual(t, expected, s, "i=%v", i) {
			// log.Print("expected: ", expected)
			// log.Print("actual:   ", s)
			// expected.Equal(s)
			break
		}
		if !assert.Equal(t, i, s.Count(), "i=%v", i) {
			break
		}
		if !assert.False(t, s.Has(i), "i=%v", i) {
			break
		}
		s = s.With(i)
		if !assert.True(t, s.Has(i), "i=%v", i) {
			break
		}
		if !assert.False(t, s.IsEmpty(), "i=%v", i) {
			break
		}
		arr = append(arr, i)
	}
}

func TestSetWithout(t *testing.T) {
	t.Parallel()

	var s Set[int]
	arr := []int{}
	N := 1_000
	if testing.Short() {
		N /= 10
	}
	for i := 0; i < N; i++ {
		s = s.With(i)
		arr = append(arr, i)
	}
	for i := 0; i < N; i++ {
		u := NewSet(arr...)
		requireSameElements(t, arr, u.Elements(), i)
		if !test.AssertSetEqual(t, u, s, "i=%v", i) {
			break
		}
		if !assert.False(t, s.IsEmpty(), "i=%v", i) {
			break
		}
		if !assert.True(t, s.Has(i), "i=%v", i) {
			break
		}
		s = s.Without(i)
		if !assert.False(t, s.Has(i), "i=%v", i) {
			break
		}
		arr = arr[1:]
	}
	assert.True(t, s.IsEmpty(), "%v %v", s.Count(), s)
}

func TestSetAny(t *testing.T) {
	t.Parallel()

	var s Set[int]
	assert.Panics(t, func() { s.Any() })
	s = s.With(1)
	assert.Equal(t, 1, s.Any())
	s = s.With(2)
	assert.Contains(t, []int{1, 2}, s.Any())
}

func TestSetAnyN(t *testing.T) {
	t.Parallel()

	var s Set[int]
	assert.Equal(t, 0, s.AnyN(0).Count())

	s = s.With(1)
	assert.Equal(t, 1, s.AnyN(1).Count())
	assert.Equal(t, 1, s.AnyN(2).Count())
	assert.Equal(t, 1, s.AnyN(4<<10-1).Count())

	s = Iota(4<<10 - 1)
	assert.Equal(t, 1, s.AnyN(1).Count())
	test.AssertSetEqual(t, Iota(4<<10-1), s.AnyN(4<<10-1))
}

func TestSetOrderedElements(t *testing.T) {
	t.Parallel()

	s := Iota(4<<10 - 1)
	less := tree.Less[int](func(a, b int) bool { return a < b })
	assert.Equal(t, generateSortedIntArray(0, 4<<10-1, 1), s.OrderedElements(less))

	less = tree.Less[int](func(a, b int) bool { return a > b })
	assert.Equal(t, generateSortedIntArray(4<<10-2, -1, -1), s.OrderedElements(less))
}

func TestSetHash(t *testing.T) {
	t.Parallel()

	maps := []Set[interface{}]{
		{},
		NewSet[interface{}](1, 2),
		NewSet[interface{}](1, 3),
		NewSet[interface{}](3, 4),
		NewSet[interface{}](3, 5),
		NewSet[interface{}](1, 3, 4),
		NewSet[interface{}](1, 3, 5),
		NewSet[interface{}](1, 2, 3, 4),
		NewSet[interface{}](1, 2, 3, 5),
		NewSet[interface{}](NewMap(kv.KV("cc", NewSet(NewMap(kv.KV("c", 1)))))),
		NewSet[interface{}](NewMap(kv.KV("cc", NewSet(NewMap(kv.KV("c", 2)))))),
	}
	for i, a := range maps {
		for j, b := range maps {
			assert.Equal(t, i == j, a.Equal(b), "i=%v j=%v a=%v b=%v", i, j, a, b)
			assert.Equal(t, i == j, a.Hash(0) == b.Hash(0),
				"i=%d j=%d a=%+v b=%+v a.Hash()=%v b.Hash()=%v", i, j, a, b, a.Hash(0), b.Hash(0))
		}
		assert.False(t, a.Equal(NewSet[interface{}](42)))
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	sets := []Set[int]{
		{},
		NewSet(1),
		NewSet(2),
		NewSet(1, 2),
		NewSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		Iota(7),
		Iota(7).AnyN(6),
	}
	for i, a := range sets {
		for j, b := range sets {
			assert.Equal(t, i == j, a.Equal(b),
				"i=%d, a=%+v\nj=%d, b=%+v", i, a, j, b)
		}
		assert.False(t, a.Equal(NewSet(42)), "i=%d", i)
	}
}

func TestSetEqualLarge(t *testing.T) {
	t.Parallel()

	n := 100_000
	if testing.Short() {
		n /= 10
	}
	a := intSet(0, n)
	b := intSet(0, n)
	// c := SetMap(intSet(n, n), func(e int) int { return e - n })

	require.Equal(t, n, a.Count())
	require.Equal(t, n, b.Count())
	// require.Equal(t, n, c.Count())
	// for i := 0; i < n; i++ {
	// 	require.True(t, c.Has(i), i)
	// }

	test.AssertSetEqual(t, a, b)
	// test.AssertSetEqual(t, a, c)
}

func TestSetFirst(t *testing.T) {
	t.Parallel()

	var s Set[int]
	less := tree.Less[int](func(a, b int) bool { return a < b })
	assert.Panics(t, func() { s.First(less) }, "empty set")

	s = Iota(4<<10 - 1)
	assert.Equal(t, 0, s.First(less))

	less = tree.Less[int](func(a, b int) bool { return a > b })
	assert.Equal(t, 4<<10-2, s.First(less))
}

func TestSetFirstN(t *testing.T) {
	t.Parallel()

	less := tree.Less[int](func(a, b int) bool { return a < b })

	s := NewSet[int]()
	assert.True(t, NewSet[int]().Equal(s.FirstN(0, less)))
	assert.True(t, NewSet[int]().Equal(s.FirstN(1, less)))

	s = Iota(4<<10 - 1)
	assert.True(t, NewSet(0).Equal(s.FirstN(1, less)))
	assert.True(t, s.Equal(s.FirstN(4<<10-1, less)))

	s = Iota(5)
	assert.True(t, NewSet(0, 1, 2, 3, 4).Equal(s.FirstN(10, less)))
}

func TestSetOrderedFirstN(t *testing.T) {
	t.Parallel()

	less := tree.Less[int](func(a, b int) bool { return a < b })

	s := NewSet[int]()
	assert.Equal(t, []int{}, s.OrderedFirstN(0, less))
	assert.Equal(t, []int{}, s.OrderedFirstN(1, less))

	s = Iota(4<<10 - 1)
	assert.Equal(t, generateSortedIntArray(0, 4<<10-1, 1), s.OrderedFirstN(4<<10-1, less))
	assert.Equal(t, []int{0, 1, 2, 3, 4}, s.OrderedFirstN(5, less))

	s = Iota(5)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, s.OrderedFirstN(10, less))
}

func TestSetIsSubsetOf(t *testing.T) {
	t.Parallel()

	const N = 10
	for i := BitIterator(0); i < N; i++ {
		a := NewSetFromMask64(uint64(i))
		for j := BitIterator(0); j < N; j++ {
			b := NewSetFromMask64(uint64(j))
			if !assert.Equal(t, i&^j == 0, a.IsSubsetOf(b)) {
				log.Print("a: ", a)
				log.Print("b: ", b)
				a = NewSetFromMask64(uint64(i))
				b = NewSetFromMask64(uint64(j))
				a.IsSubsetOf(b)
				return
			}
		}
	}
}

func TestSetIsSubsetOfLarge(t *testing.T) {
	t.Parallel()

	n := 100_000
	if testing.Short() {
		n /= 10
	}
	a := intSet(0, n)
	b := intSet(0, n+1)
	c := intSet(1, n)
	assert.True(t, a.IsSubsetOf(a))
	assert.True(t, a.IsSubsetOf(b))
	assert.False(t, b.IsSubsetOf(a))
	assert.False(t, a.IsSubsetOf(c))
	assert.False(t, c.IsSubsetOf(a))
}

func TestSetString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "{}", Set[int]{}.String())
	assert.Equal(t, "{1}", NewSet(1).String())
	assert.Contains(t, []string{"{1, 2}", "{2, 1}"}, NewSet(1, 2).String())
	assert.Contains(t, []string{"{1, 2}", "{2, 1}"}, NewSet(2, 1).String())
}

func TestSetWhereEmpty(t *testing.T) {
	t.Parallel()

	test.AssertSetEqual(t, NewSet[int](), NewSet[int]().Where(func(int) bool { return false }))
	test.AssertSetEqual(t, NewSet[int](), NewSet[int]().Where(func(int) bool { return true }))
}

func TestSetWhere(t *testing.T) {
	t.Parallel()

	s := Iota2(1, 10)
	multipleOf := func(n int) func(e int) bool {
		return func(e int) bool { return e%n == 0 }
	}
	not := func(f func(e int) bool) func(e int) bool {
		return func(e int) bool { return !f(e) }
	}
	test.AssertSetEqual(t, Iota3(2, 10, 2), s.Where(multipleOf(2)))
	test.AssertSetEqual(t, Iota3(1, 10, 2), s.Where(not(multipleOf(2))))
	test.AssertSetEqual(t, Iota3(3, 10, 3), s.Where(multipleOf(3)))
	test.AssertSetEqual(t, NewSet(1, 2, 4, 5, 7, 8), s.Where(not(multipleOf(3))))
	test.AssertSetEqual(t, NewSet(6), s.Where(multipleOf(2)).Where(multipleOf(3)))
	test.AssertSetEqual(t, NewSet(6), s.Where(multipleOf(3)).Where(multipleOf(2)))
	test.AssertSetEqual(t, NewSet(3, 9), s.Where(not(multipleOf(2))).Where(multipleOf(3)))
	test.AssertSetEqual(t, NewSet(3, 9), s.Where(multipleOf(3)).Where(not(multipleOf(2))))
	test.AssertSetEqual(t, NewSet(2, 4, 8), s.Where(multipleOf(2)).Where(not(multipleOf(3))))
	test.AssertSetEqual(t, NewSet(2, 4, 8), s.Where(not(multipleOf(3))).Where(multipleOf(2)))
	test.AssertSetEqual(t, NewSet(1, 5, 7), s.Where(not(multipleOf(2))).Where(not(multipleOf(3))))
	test.AssertSetEqual(t, NewSet(1, 5, 7), s.Where(not(multipleOf(3))).Where(not(multipleOf(2))))
}

// func TestSetMap(t *testing.T) {
// 	t.Parallel()

// 	square := func(e interface{}) interface{} { return e.(int) * e.(int) }
// 	div2 := func(e interface{}) interface{} { return e.(int) / 2 }
// 	test.AssertSetEqual(t, Set[int]{}, Set[int]{}.Map(square))
// 	test.AssertSetEqual(t, NewSet(1, 4, 9, 16, 25), NewSet(1, 2, 3, 4, 5).Map(square))
// 	test.AssertSetEqual(t, NewSet(0, 1, 3), NewSet(1, 2, 3, 6).Map(div2))
// }

// func TestSetMapShrunk(t *testing.T) {
// 	t.Parallel()

// 	div2 := func(e interface{}) interface{} { return e.(int) / 2 }
// 	s := NewSet(1, 2, 3, 6)
// 	mapped := s.Map(div2)
// 	assert.Equal(t, 3, mapped.Count(), "%v", mapped)
// }

// func TestSetMapLarge(t *testing.T) {
// 	t.Parallel()

// 	s := intSet(0, 50)
// 	// assertSetEqual(t, NewSet(42), s.Map(func(e interface{}) interface{} { return 42 }))
// 	if !test.AssertSetEqual(t, Iota3(0, 2*s.Count(), 2), s.Map(func(e interface{}) interface{} { return 2 * e.(int) })) {
// 		expected := Iota3(0, 2*s.Count(), 2)
// 		actual := s.Map(func(e interface{}) interface{} { return 2 * e.(int) })
// 		log.Print(expected)
// 		log.Print(actual)
// 		for {
// 			s.Map(func(e interface{}) interface{} { return 2 * e.(int) })
// 		}
// 	}
// 	test.AssertSetEqual(t, Iota(s.Count()/10), s.Map(func(e interface{}) interface{} { return e.(int) / 10 }))
// }

// func TestSetReduce(t *testing.T) {
// 	t.Parallel()

// 	sum := func(acc, b interface{}) interface{} { return acc.(int) + b.(int) }
// 	product := func(acc, b interface{}) interface{} { return acc.(int) * b.(int) }

// 	if !assert.NotPanics(t, func() { Iota2(1, 11).Reduce2(sum) }) {
// 		i := Iota2(1, 11)
// 		i.Reduce2(sum)
// 	}

// 	assert.Equal(t, 42, NewSet(42).Reduce2(sum))
// 	assert.Equal(t, 42, NewSet(42).Reduce2(product))
// 	assert.Equal(t, 12, NewSet(5, 7).Reduce2(sum))
// 	assert.Equal(t, 35, NewSet(5, 7).Reduce2(product))
// 	assert.Equal(t, 55, Iota2(1, 11).Reduce2(sum))
// 	assert.Equal(t, 720, Iota2(2, 7).Reduce2(product))
// }

// func TestSetReduceHuge(t *testing.T) {
// 	t.Parallel()

// 	sum := func(a, b interface{}) interface{} { return a.(int) + b.(int) }

// 	n := hugeCollectionSize()
// 	assert.Equal(t, (n-1)*n/2, Iota(n).Reduce2(sum))
// }

func testSetBinaryOperator(t *testing.T, bitop func(a, b uint64) uint64, setop func(a, b Set[int]) Set[int]) { //nolint:cyclop
	t.Helper()

	m := map[uint64]struct{}{
		0x0000: {}, // 000000000000000
		0x0001: {}, // 000000000000001
		0x0002: {}, // 000000000000010
		0x0210: {}, // 000001000010000
		0x7fff: {}, // 111111111111111
		0x2aaa: {}, // 010101010101010
		0x4924: {}, // 100100100100100
		0x0888: {}, // 000100010001000
		0x4210: {}, // 100001000010000
	}
	f := 10
	if testing.Short() {
		f = 1
	}

	for i := 0; i < f*10; i++ {
		m[uint64(i)] = struct{}{}
	}
	for i := 100; i < f*1_000; i += 100 {
		m[uint64(i)] = struct{}{}
	}
	for i := 1_000; i < f*100_000; i += 10_000 {
		m[uint64(i)] = struct{}{}
	}
	for i := 1_000_000; i < f*10_000_000; i += 1_000_000 {
		m[uint64(i)] = struct{}{}
	}
	sets := make([]uint64, 0, len(m))
	for set := range m {
		sets = append(sets, set)
	}

	for _, x := range sets {
		for _, y := range sets {
			sx := NewSetFromMask64(x)
			sy := NewSetFromMask64(y)
			bxy := bitop(x, y)
			sxy := NewSetFromMask64(bxy)
			sxsy := setop(sx, sy)
			if !assert.Equal(t, bits.OnesCount64(bxy), sxsy.Count()) ||
				!test.AssertSetEqual(t, sxy, sxsy, "sx=%v sy=%v", sx, sy) {
				log.Print("sx:   ", sx)
				log.Print("sy:   ", sy)
				log.Print("bxy:  ", bxy)
				log.Print("sxy:  ", sxy)
				log.Print("sxsy: ", sxsy)
				log.Print("==:   ", sxy.Equal(sxsy))
				if bits.OnesCount64(x) < 4 && bits.OnesCount64(y) < 4 {
					NewSetFromMask64(x)
					NewSetFromMask64(y)
					setop(sx, sy)
					sxy.Equal(sxsy)
				}
				return
			}
		}
	}
}

func TestSetIntersection(t *testing.T) {
	t.Parallel()

	expected := NewSet(3, 5, 6, 10, 11)
	a := NewSet(3, 5, 6, 7, 10, 11, 12)
	b := NewSet(3, 5, 6, 8, 9, 10, 11, 15, 19)
	test.AssertSetEqual(t, expected, a.Intersection(b))

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a & b },
		func(a, b Set[int]) Set[int] { return a.Intersection(b) },
	)
}

func TestSetIntersectionLarge(t *testing.T) {
	t.Parallel()

	bits := 13
	if testing.Short() {
		bits -= 3
	}
	for i := 0; i <= bits; i++ {
		a := Iota2(1<<uint(i), 9<<uint(i))
		b := Iota(9 << uint(i)).Intersection(Iota2(1<<uint(i), 10<<uint(i)))
		if !test.AssertSetEqual(t, a, b, "%d", i) {
			assert.ElementsMatch(t, a.Elements(), b.Elements())
			t.FailNow()
		}
	}
}

func TestSetUnion(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a | b },
		func(a, b Set[int]) Set[int] { return a.Union(b) },
	)
}

func TestSetDifference(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a &^ b },
		func(a, b Set[int]) Set[int] { return a.Difference(b) },
	)
}

func TestSetSymmetricDifference(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a ^ b },
		func(a, b Set[int]) Set[int] { return a.SymmetricDifference(b) },
	)
}

// func TestSetPowerset(t *testing.T) {
// 	t.Parallel()

// 	expected := NewSet(
// 		NewSet(),
// 		NewSet(3),
// 		NewSet(2),
// 		NewSet(2, 3),
// 		NewSet(1),
// 		NewSet(1, 3),
// 		NewSet(1, 2),
// 		NewSet(1, 2, 3),
// 	)
// 	actual := NewSet(1, 2, 3).Powerset()
// 	if !test.AssertSetEqual(t, expected, actual, "%v", mapOfSet{"expected": expected, "actual": actual}) {
// 		log.Print("expected: ", expected)
// 		log.Print("actual:   ", actual)
// 		expected.Equal(actual)
// 	}
// }

// func TestSetPowersetLarge(t *testing.T) {
// 	t.Parallel()

// 	expected := NewSet()
// 	var b SetBuilder
// 	bits := 15
// 	if testing.Short() {
// 		bits -= 3
// 	}
// 	for i := BitIterator(0); i <= 1<<bits; i++ {
// 		if i.Count() == 1 {
// 			expected = expected.Union(b.Finish())
// 			test.AssertSetEqual(t, expected, NewSetFromMask64(uint64(i-1)).Powerset(), "i=%v", i)
// 		}
// 		b.Add(NewSetFromMask64(uint64(i)))
// 	}
// }

// func TestSetGroupBy(t *testing.T) {
// 	t.Parallel()

// 	const N = 100
// 	const D = 7
// 	group := Iota(N).GroupBy(func(el interface{}) interface{} {
// 		return el % D
// 	})
// 	var b MapBuilder
// 	for i := 0; i < D; i++ {
// 		b.Put(i, Iota3(i, N, D))
// 	}
// 	assertMapEqual(t, b.Finish(), group)
// }

func TestSetRange(t *testing.T) {
	t.Parallel()

	mask := uint64(0)
	for i := Iota(64).Range(); i.Next(); {
		mask |= uint64(1) << uint(i.Value())
	}
	assert.Equal(t, ^uint64(0), mask)
}

func TestSetOrderedRange(t *testing.T) {
	t.Parallel()

	output := []int{}
	less := tree.Less[int](func(a, b int) bool { return a < b })
	for i := Iota(10).OrderedRange(less); i.Next(); {
		output = append(output, i.Value())
	}
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, output)

	output = output[:0]
	less = tree.Less[int](func(a, b int) bool { return a > b })
	for i := Iota(10).OrderedRange(less); i.Next(); {
		output = append(output, i.Value())
	}
	assert.Equal(t, []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, output)
}

func TestSetWhere_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Where(func(e int) bool { return true })
	test.AssertSetEqual(t, s, s2)

	s = hugeIntSet()
	s2 = s.Where(func(e int) bool { return true })
	test.AssertSetEqual(t, s, s2)

	s = largeIntSet()
	s2 = s.Where(func(e int) bool { return false })
	test.AssertSetEqual(t, NewSet[int](), s2)
}

func TestSetIntersection_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Intersection(s)
	test.AssertSetEqual(t, s, s2)

	if !testing.Short() {
		s = hugeIntSet()
		s2 = s.Intersection(s)
		test.AssertSetEqual(t, s, s2)
	}
}

func TestSetDifference_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Difference(s)
	test.AssertSetEqual(t, NewSet[int](), s2)

	s = hugeIntSet()
	s2 = s.Difference(s)
	test.AssertSetEqual(t, NewSet[int](), s2)
}

func TestSetUnion_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Union(s)
	test.AssertSetEqual(t, s, s2)

	n := 100_000
	if testing.Short() {
		n /= 10
	}
	s = intSet(0, n)
	s2 = intSet(n/2, n)
	test.AssertSetEqual(t, intSet(0, n*3/2), s.Union(s2))

	s = hugeIntSet()
	s2 = s.Union(s)
	test.AssertSetEqual(t, s, s2)
}

func intSet(offset, size int) Set[int] {
	sb := NewSetBuilder[int](size)
	for i := offset; i < offset+size; i++ {
		sb.Add(i)
	}
	return sb.Finish()
}

var prepopSetInt = memoizePrepop(func(n int) interface{} {
	m := make(map[int]struct{}, n)
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInt(b *testing.B, n int) {
	b.Helper()

	m := prepopSetInt(n).(map[int]struct{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInt0(b *testing.B) {
	benchmarkInsertSetInt(b, 0)
}

func BenchmarkInsertSetInt1k(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<10)
}

func BenchmarkInsertSetInt1M(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<20)
}

var prepopSetInterface = memoizePrepop(func(n int) interface{} {
	m := make(map[interface{}]struct{}, n)
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInterface(b *testing.B, n int) {
	b.Helper()

	m := prepopSetInterface(n).(map[interface{}]struct{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInterface0(b *testing.B) {
	benchmarkInsertSetInterface(b, 0)
}

func BenchmarkInsertSetInterface1k(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<10)
}

func BenchmarkInsertSetInterface1M(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<20)
}

var prepopFrozenSet = memoizePrepop(func(n int) interface{} {
	var s Set[int]
	for i := 0; i < n; i++ {
		s = s.With(i)
	}
	return s
})

func benchmarkInsertFrozenSet(b *testing.B, n int) {
	b.Helper()

	s := prepopFrozenSet(n).(Set[int])
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		s.With(i)
	}
}

func BenchmarkInsertFrozenSet0(b *testing.B) {
	benchmarkInsertFrozenSet(b, 0)
}

func BenchmarkInsertFrozenSet1k(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<10)
}

func BenchmarkInsertFrozenSet1M(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<20)
}
