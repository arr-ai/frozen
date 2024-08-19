package frozen_test

import (
	"log"
	"math/bits"
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/iterator"
	"github.com/arr-ai/frozen/internal/pkg/test"
	testset "github.com/arr-ai/frozen/internal/pkg/test/set"
	"github.com/arr-ai/frozen/internal/pkg/tree"
)

func largeIntSet() frozen.Set[int] {
	return intSet(0, 10_000)
}

func hugeIntSet() frozen.Set[int] {
	return intSet(0, hugeCollectionSize())
}

func TestSetEmpty(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
	test.True(t, s.IsEmpty())
	testset.AssertSetEqual(t, frozen.Set[int]{}, s)
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

// faster that test.ElementsMatch.
func assertSameElements[T comparable](t *testing.T, a, b []T, msgAndArgs ...any) bool {
	t.Helper()

	aOnly, bOnly := compareElements(a, b)
	aOK := test.Empty(t, aOnly, msgAndArgs...)
	bOK := test.Empty(t, bOnly, msgAndArgs...)
	return aOK && bOK
}

func requireSameElements[T comparable](t *testing.T, a, b []T, msgAndArgs ...any) {
	t.Helper()

	if !assertSameElements(t, a, b, msgAndArgs...) {
		t.FailNow()
	}
}

func TestSetNewSet(t *testing.T) {
	t.Parallel()

	N := 1_000
	if testing.Short() {
		N /= 10
	}
	arr := make([]int, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, i)
	}

	for i := N - 1; i >= 0; i-- {
		expected := arr[i:]
		actual := frozen.NewSet(arr[i:]...)
		test.RequireEqual(t, len(expected), actual.Count())
		assertSameElements(t, expected, actual.Elements())
	}
}

func TestSetNew2(t *testing.T) {
	t.Parallel()

	a := frozen.NewSet(1, 2)
	b := frozen.NewSet(2, 1)
	test.True(t, a.Equal(b))
}

func TestSetNewSetFromStrings(t *testing.T) {
	t.Parallel()

	const N = 256
	arr := make([]string, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, string(rune(i)))
	}

	for i := N - 1; i >= 0; i-- {
		assertSameElements(t, arr[i:], frozen.NewSet(arr[i:]...).Elements())
	}
}

func TestSetWith(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
	arr := []int{}
	n := 1_000
	if testing.Short() {
		n /= 10
	}
	for i := 0; i < n; i++ {
		expected := frozen.NewSet(arr...)
		if !testset.AssertSetEqual(t, expected, s, "i=%v", i) {
			// log.Print("expected: ", expected)
			// log.Print("actual:   ", s)
			// expected.Equal(s)
			break
		}
		if !test.Equal(t, i, s.Count(), "i=%v", i) {
			break
		}
		if !test.False(t, s.Has(i), "i=%v", i) {
			break
		}
		s = s.With(i)
		if !test.True(t, s.Has(i), "i=%v", i) {
			break
		}
		if !test.False(t, s.IsEmpty(), "i=%v", i) {
			break
		}
		arr = append(arr, i)
	}
}

func TestSetWithout(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
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
		u := frozen.NewSet(arr...)
		requireSameElements(t, arr, u.Elements(), i)
		if !testset.AssertSetEqual(t, u, s, "i=%v", i) {
			break
		}
		if !test.False(t, s.IsEmpty(), "i=%v", i) {
			break
		}
		if !test.True(t, s.Has(i), "i=%v", i) {
			break
		}
		s = s.Without(i)
		if !test.False(t, s.Has(i), "i=%v", i) {
			break
		}
		arr = arr[1:]
	}
	test.True(t, s.IsEmpty(), "%v %v", s.Count(), s)
}

func TestSetAny(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
	test.Panic(t, func() { s.Any() })
	s = s.With(1)
	test.Equal(t, 1, s.Any())
	s = s.With(2)
	i := s.Any()
	test.True(t, i == 1 || i == 2)
}

func TestSetAnyN(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
	test.Equal(t, 0, s.AnyN(0).Count())

	s = s.With(1)
	test.Equal(t, 1, s.AnyN(1).Count())
	test.Equal(t, 1, s.AnyN(2).Count())
	test.Equal(t, 1, s.AnyN(4<<10-1).Count())

	s = frozen.Iota(4<<10 - 1)
	test.Equal(t, 1, s.AnyN(1).Count())
	testset.AssertSetEqual(t, frozen.Iota(4<<10-1), s.AnyN(4<<10-1))
}

func TestSetOrderedElements(t *testing.T) {
	t.Parallel()

	s := frozen.Iota(4<<10 - 1)
	less := tree.Less[int](func(a, b int) bool { return a < b })
	test.Equal(t, generateSortedIntArray(0, 4<<10-1, 1), s.OrderedElements(less))

	less = tree.Less[int](func(a, b int) bool { return a > b })
	test.Equal(t, generateSortedIntArray(4<<10-2, -1, -1), s.OrderedElements(less))
}

func TestSetHash(t *testing.T) {
	t.Parallel()

	maps := []frozen.Set[any]{
		{},
		frozen.NewSet[any](1, 2),
		frozen.NewSet[any](1, 3),
		frozen.NewSet[any](3, 4),
		frozen.NewSet[any](3, 5),
		frozen.NewSet[any](1, 3, 4),
		frozen.NewSet[any](1, 3, 5),
		frozen.NewSet[any](1, 2, 3, 4),
		frozen.NewSet[any](1, 2, 3, 5),
		frozen.NewSet[any](frozen.NewMap(frozen.KV("cc", frozen.NewSet(frozen.NewMap(frozen.KV("c", 1)))))),
		frozen.NewSet[any](frozen.NewMap(frozen.KV("cc", frozen.NewSet(frozen.NewMap(frozen.KV("c", 2)))))),
	}
	for i, a := range maps {
		for j, b := range maps {
			test.Equal(t, i == j, a.Equal(b), "i=%v j=%v a=%v b=%v", i, j, a, b)
			test.Equal(t, i == j, a.Hash(0) == b.Hash(0),
				"i=%d j=%d a=%+v b=%+v a.Hash()=%v b.Hash()=%v", i, j, a, b, a.Hash(0), b.Hash(0))
		}
		test.False(t, a.Equal(frozen.NewSet[any](42)))
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	sets := []frozen.Set[int]{
		{},
		frozen.NewSet(1),
		frozen.NewSet(2),
		frozen.NewSet(1, 2),
		frozen.NewSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		frozen.Iota(7),
		frozen.Iota(7).AnyN(6),
	}
	for i, a := range sets {
		for j, b := range sets {
			test.Equal(t, i == j, a.Equal(b),
				"i=%d, a=%+v\nj=%d, b=%+v", i, a, j, b)
		}
		test.False(t, a.Equal(frozen.NewSet(42)), "i=%d", i)
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
	// c := frozen.SetMap(intSet(n, n), func(e int) int { return e - n })

	test.RequireEqual(t, n, a.Count())
	test.RequireEqual(t, n, b.Count())
	// test.RequireEqual(t, n, c.Count())
	// for i := 0; i < n; i++ {
	// 	test.RequireTrue(t, c.Has(i), i)
	// }

	testset.AssertSetEqual(t, a, b)
	// testset.AssertSetEqual(t, a, c)
}

func TestSetFirst(t *testing.T) {
	t.Parallel()

	var s frozen.Set[int]
	less := tree.Less[int](func(a, b int) bool { return a < b })
	test.Panic(t, func() { s.First(less) }, "empty set")

	s = frozen.Iota(4<<10 - 1)
	test.Equal(t, 0, s.First(less))

	less = tree.Less[int](func(a, b int) bool { return a > b })
	test.Equal(t, 4<<10-2, s.First(less))
}

func TestSetFirstN(t *testing.T) {
	t.Parallel()

	less := tree.Less[int](func(a, b int) bool { return a < b })

	s := frozen.NewSet[int]()
	test.True(t, frozen.NewSet[int]().Equal(s.FirstN(0, less)))
	test.True(t, frozen.NewSet[int]().Equal(s.FirstN(1, less)))

	s = frozen.Iota(4<<10 - 1)
	test.True(t, frozen.NewSet(0).Equal(s.FirstN(1, less)))
	test.True(t, s.Equal(s.FirstN(4<<10-1, less)))

	s = frozen.Iota(5)
	test.True(t, frozen.NewSet(0, 1, 2, 3, 4).Equal(s.FirstN(10, less)))
}

func TestSetOrderedFirstN(t *testing.T) {
	t.Parallel()

	less := tree.Less[int](func(a, b int) bool { return a < b })

	s := frozen.NewSet[int]()
	test.Equal(t, []int{}, s.OrderedFirstN(0, less))
	test.Equal(t, []int{}, s.OrderedFirstN(1, less))

	s = frozen.Iota(4<<10 - 1)
	test.Equal(t, generateSortedIntArray(0, 4<<10-1, 1), s.OrderedFirstN(4<<10-1, less))
	test.Equal(t, []int{0, 1, 2, 3, 4}, s.OrderedFirstN(5, less))

	s = frozen.Iota(5)
	test.Equal(t, []int{0, 1, 2, 3, 4}, s.OrderedFirstN(10, less))
}

func TestSetIsSubsetOf(t *testing.T) {
	t.Parallel()

	const N = 10
	for i := iterator.BitIterator(0); i < N; i++ {
		a := frozen.NewSetFromMask64(uint64(i))
		for j := iterator.BitIterator(0); j < N; j++ {
			b := frozen.NewSetFromMask64(uint64(j))
			if !test.Equal(t, i&^j == 0, a.IsSubsetOf(b)) {
				log.Print("a: ", a)
				log.Print("b: ", b)
				a = frozen.NewSetFromMask64(uint64(i))
				b = frozen.NewSetFromMask64(uint64(j))
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
	test.True(t, a.IsSubsetOf(a))
	test.True(t, a.IsSubsetOf(b))
	test.False(t, b.IsSubsetOf(a))
	test.False(t, a.IsSubsetOf(c))
	test.False(t, c.IsSubsetOf(a))
}

func TestSetString(t *testing.T) {
	t.Parallel()

	test.Equal(t, "{}", frozen.Set[int]{}.String())
	test.Equal(t, "{1}", frozen.NewSet(1).String())
	s := frozen.NewSet(1, 2).String()
	test.True(t, s == "{1, 2}" || s == "{2, 1}")
	s = frozen.NewSet(2, 1).String()
	test.True(t, s == "{1, 2}" || s == "{2, 1}")
}

func TestSetWhereEmpty(t *testing.T) {
	t.Parallel()

	testset.AssertSetEqual(t, frozen.NewSet[int](), frozen.NewSet[int]().Where(func(int) bool { return false }))
	testset.AssertSetEqual(t, frozen.NewSet[int](), frozen.NewSet[int]().Where(func(int) bool { return true }))
}

func TestSetWhere(t *testing.T) {
	t.Parallel()

	s := frozen.Iota2(1, 10)
	multipleOf := func(n int) func(e int) bool {
		return func(e int) bool { return e%n == 0 }
	}
	not := func(f func(e int) bool) func(e int) bool {
		return func(e int) bool { return !f(e) }
	}
	testset.AssertSetEqual(t, frozen.Iota3(2, 10, 2), s.Where(multipleOf(2)))
	testset.AssertSetEqual(t, frozen.Iota3(1, 10, 2), s.Where(not(multipleOf(2))))
	testset.AssertSetEqual(t, frozen.Iota3(3, 10, 3), s.Where(multipleOf(3)))
	testset.AssertSetEqual(t, frozen.NewSet(1, 2, 4, 5, 7, 8), s.Where(not(multipleOf(3))))
	testset.AssertSetEqual(t, frozen.NewSet(6), s.Where(multipleOf(2)).Where(multipleOf(3)))
	testset.AssertSetEqual(t, frozen.NewSet(6), s.Where(multipleOf(3)).Where(multipleOf(2)))
	testset.AssertSetEqual(t, frozen.NewSet(3, 9), s.Where(not(multipleOf(2))).Where(multipleOf(3)))
	testset.AssertSetEqual(t, frozen.NewSet(3, 9), s.Where(multipleOf(3)).Where(not(multipleOf(2))))
	testset.AssertSetEqual(t, frozen.NewSet(2, 4, 8), s.Where(multipleOf(2)).Where(not(multipleOf(3))))
	testset.AssertSetEqual(t, frozen.NewSet(2, 4, 8), s.Where(not(multipleOf(3))).Where(multipleOf(2)))
	testset.AssertSetEqual(t, frozen.NewSet(1, 5, 7), s.Where(not(multipleOf(2))).Where(not(multipleOf(3))))
	testset.AssertSetEqual(t, frozen.NewSet(1, 5, 7), s.Where(not(multipleOf(3))).Where(not(multipleOf(2))))
}

func TestSetMap(t *testing.T) {
	t.Parallel()

	square := func(e int) int { return e * e }
	div2 := func(e int) int { return e / 2 }
	testset.AssertSetEqual(t, frozen.Set[int]{}, frozen.SetMap(frozen.Set[int]{}, square))
	testset.AssertSetEqual(t, frozen.NewSet(1, 4, 9, 16, 25), frozen.SetMap(frozen.NewSet(1, 2, 3, 4, 5), square))
	testset.AssertSetEqual(t, frozen.NewSet(0, 1, 3), frozen.SetMap(frozen.NewSet(1, 2, 3, 6), div2))
}

func TestSetMapShrunk(t *testing.T) {
	t.Parallel()

	div2 := func(e int) int { return e / 2 }
	s := frozen.NewSet(1, 2, 3, 6)
	mapped := frozen.SetMap(s, div2)
	test.Equal(t, 3, mapped.Count(), "%v", mapped)
}

func TestSetMapLarge(t *testing.T) {
	t.Parallel()

	s := intSet(0, 50)
	// assertSetEqual(t, frozen.NewSet(42), s.Map(func(e any) any { return 42 }))
	if !testset.AssertSetEqual(t, frozen.Iota3(0, 2*s.Count(), 2), frozen.SetMap(s, func(e int) int { return 2 * e })) {
		expected := frozen.Iota3(0, 2*s.Count(), 2)
		actual := frozen.SetMap(s, func(e int) int { return 2 * e })
		log.Print(expected)
		log.Print(actual)
		for {
			frozen.SetMap(s, func(e int) int { return 2 * e })
		}
	}
	testset.AssertSetEqual(t, frozen.Iota(s.Count()/10), frozen.SetMap(s, func(e int) int { return e / 10 }))
}

func TestSetReduce(t *testing.T) {
	t.Parallel()

	sum := func(acc, b int) int { return acc + b }
	product := func(acc, b int) int { return acc * b }

	if !test.NoPanic(t, func() { frozen.Iota2(1, 11).Reduce2(sum) }) {
		i := frozen.Iota2(1, 11)
		i.Reduce2(sum)
	}

	assertReduce2 := func(expected int, s frozen.Set[int], f func(acc, b int) int) bool {
		actual, ok := s.Reduce2(f)
		return test.True(t, ok) && test.Equal(t, expected, actual)
	}
	assertReduce2(42, frozen.NewSet(42), sum)
	assertReduce2(42, frozen.NewSet(42), product)
	assertReduce2(12, frozen.NewSet(5, 7), sum)
	assertReduce2(35, frozen.NewSet(5, 7), product)
	assertReduce2(55, frozen.Iota2(1, 11), sum)
	assertReduce2(720, frozen.Iota2(2, 7), product)
}

func TestSetReduceHuge(t *testing.T) {
	t.Parallel()

	sum := func(a, b int) int { return a + b }

	n := hugeCollectionSize()
	actual, ok := frozen.Iota(n).Reduce2(sum)
	_ = test.True(t, ok) && test.Equal(t, (n-1)*n/2, actual)
}

func testSetBinaryOperatorPair(
	t *testing.T,
	bitop func(a, b uint64) uint64,
	setop func(a, b frozen.Set[int]) frozen.Set[int],
	x, y uint64,
) {
	t.Helper()

	sx := frozen.NewSetFromMask64(x)
	sy := frozen.NewSetFromMask64(y)
	bxy := bitop(x, y)
	sxy := frozen.NewSetFromMask64(bxy)
	sxsy := setop(sx, sy)
	if !test.Equal(t, bits.OnesCount64(bxy), sxsy.Count()) ||
		!testset.AssertSetEqual(t, sxy, sxsy, "sx=%v sy=%v", sx, sy) {
		log.Print("sx:   ", sx)
		log.Print("sy:   ", sy)
		log.Print("bxy:  ", bxy)
		log.Print("sxy:  ", sxy)
		log.Print("sxsy: ", sxsy)
		log.Print("==:   ", sxy.Equal(sxsy))
		if bits.OnesCount64(x) < 4 && bits.OnesCount64(y) < 4 {
			frozen.NewSetFromMask64(x)
			frozen.NewSetFromMask64(y)
			setop(sx, sy)
			sxy.Equal(sxsy)
		}
		return
	}
}

func testSetBinaryOperator(
	t *testing.T,
	bitop func(a, b uint64) uint64,
	setop func(a, b frozen.Set[int]) frozen.Set[int],
) {
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
			testSetBinaryOperatorPair(t, bitop, setop, x, y)
		}
	}
}

func TestSetIntersection(t *testing.T) {
	t.Parallel()

	expected := frozen.NewSet(3, 5, 6, 10, 11)
	a := frozen.NewSet(3, 5, 6, 7, 10, 11, 12)
	b := frozen.NewSet(3, 5, 6, 8, 9, 10, 11, 15, 19)
	testset.AssertSetEqual(t, expected, a.Intersection(b))

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a & b },
		func(a, b frozen.Set[int]) frozen.Set[int] { return a.Intersection(b) },
	)
}

func TestSetIntersectionLarge(t *testing.T) {
	t.Parallel()

	bits := 13
	if testing.Short() {
		bits -= 3
	}
	for i := 0; i <= bits; i++ {
		a := frozen.Iota2(1<<uint(i), 9<<uint(i))
		b := frozen.Iota(9 << uint(i)).Intersection(frozen.Iota2(1<<uint(i), 10<<uint(i)))
		if !testset.AssertSetEqual(t, a, b, "%d", i) {
			test.ElementsMatch(t, a.Elements(), b.Elements())
			t.FailNow()
		}
	}
}

func TestSetUnion(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a | b },
		func(a, b frozen.Set[int]) frozen.Set[int] { return a.Union(b) },
	)
}

func TestSetDifference(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a &^ b },
		func(a, b frozen.Set[int]) frozen.Set[int] { return a.Difference(b) },
	)
}

func TestSetSymmetricDifference(t *testing.T) {
	t.Parallel()

	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a ^ b },
		func(a, b frozen.Set[int]) frozen.Set[int] { return a.SymmetricDifference(b) },
	)
}

func TestSetPowerset(t *testing.T) {
	t.Parallel()

	expected := frozen.NewSet(
		frozen.NewSet[int](),
		frozen.NewSet(3),
		frozen.NewSet(2),
		frozen.NewSet(2, 3),
		frozen.NewSet(1),
		frozen.NewSet(1, 3),
		frozen.NewSet(1, 2),
		frozen.NewSet(1, 2, 3),
	)
	actual := frozen.Powerset(frozen.NewSet(1, 2, 3))
	if !test.True(t, expected.Equal(actual),
		"%v", map[string]frozen.Set[frozen.Set[int]]{"expected": expected, "actual": actual},
	) {
		log.Print("expected: ", expected)
		log.Print("actual:   ", actual)
		expected.Equal(actual)
	}
}

func TestSetPowersetLarge(t *testing.T) {
	t.Parallel()

	expected := frozen.NewSet[frozen.Set[int]]()
	var b frozen.SetBuilder[frozen.Set[int]]
	bits := 15
	if testing.Short() {
		bits -= 3
	}
	for i := iterator.BitIterator(0); i <= 1<<bits; i++ {
		if i.Count() == 1 {
			expected = expected.Union(b.Finish())
			test.True(t, expected.Equal(frozen.Powerset(frozen.NewSetFromMask64(uint64(i-1)))), "i=%v", i)
		}
		b.Add(frozen.NewSetFromMask64(uint64(i)))
	}
}

func TestSetGroupBy(t *testing.T) {
	t.Parallel()

	const N = 100
	const D = 7
	group := frozen.SetGroupBy(frozen.Iota(N), func(el int) int {
		return el % D
	})
	var b frozen.MapBuilder[int, frozen.Set[int]]
	for i := 0; i < D; i++ {
		b.Put(i, frozen.Iota3(i, N, D))
	}
	assertMapEqual(t, b.Finish(), group)
}

func TestSetRange(t *testing.T) {
	t.Parallel()

	mask := uint64(0)
	for i := frozen.Iota(64).Range(); i.Next(); {
		mask |= uint64(1) << uint(i.Value())
	}
	test.Equal(t, ^uint64(0), mask)
}

func TestSetOrderedRange(t *testing.T) {
	t.Parallel()

	output := []int{}
	less := tree.Less[int](func(a, b int) bool { return a < b })
	for i := frozen.Iota(10).OrderedRange(less); i.Next(); {
		output = append(output, i.Value())
	}
	test.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, output)

	output = output[:0]
	less = tree.Less[int](func(a, b int) bool { return a > b })
	for i := frozen.Iota(10).OrderedRange(less); i.Next(); {
		output = append(output, i.Value())
	}
	test.Equal(t, []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, output)
}

func TestSetWhere_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Where(func(int) bool { return true })
	testset.AssertSetEqual(t, s, s2)

	s = hugeIntSet()
	s2 = s.Where(func(int) bool { return true })
	testset.AssertSetEqual(t, s, s2)

	s = largeIntSet()
	s2 = s.Where(func(int) bool { return false })
	testset.AssertSetEqual(t, frozen.NewSet[int](), s2)
}

func TestSetIntersection_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Intersection(s)
	testset.AssertSetEqual(t, s, s2)

	if !testing.Short() {
		s = hugeIntSet()
		s2 = s.Intersection(s)
		testset.AssertSetEqual(t, s, s2)
	}
}

func TestSetDifference_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Difference(s)
	testset.AssertSetEqual(t, frozen.NewSet[int](), s2)

	s = hugeIntSet()
	s2 = s.Difference(s)
	testset.AssertSetEqual(t, frozen.NewSet[int](), s2)
}

func TestSetUnion_Big(t *testing.T) {
	t.Parallel()

	s := largeIntSet()
	s2 := s.Union(s)
	testset.AssertSetEqual(t, s, s2)

	n := 100_000
	if testing.Short() {
		n /= 10
	}
	s = intSet(0, n)
	s2 = intSet(n/2, n)
	testset.AssertSetEqual(t, intSet(0, n*3/2), s.Union(s2))

	s = hugeIntSet()
	s2 = s.Union(s)
	testset.AssertSetEqual(t, s, s2)
}

func intSet(offset, size int) frozen.Set[int] {
	sb := frozen.NewSetBuilder[int](size)
	for i := offset; i < offset+size; i++ {
		sb.Add(i)
	}
	return sb.Finish()
}

var prepopSetInt = memoizePrepop(func(n int) any {
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

var prepopSetInterface = memoizePrepop(func(n int) any {
	m := make(map[any]struct{}, n)
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInterface(b *testing.B, n int) {
	b.Helper()

	m := prepopSetInterface(n).(map[any]struct{})
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

var prepopFrozenSet = memoizePrepop(func(n int) any {
	var s frozen.Set[int]
	for i := 0; i < n; i++ {
		s = s.With(i)
	}
	return s
})

func benchmarkInsertFrozenSet(b *testing.B, n int) {
	b.Helper()

	s := prepopFrozenSet(n).(frozen.Set[int])
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
