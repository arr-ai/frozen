package frozen_test

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

func hugeCollectionSize() int {
	if testing.Short() {
		return 50_000
	}
	return 1_000_000
}

func TestIntSetEmpty(t *testing.T) {
	t.Parallel()

	var s frozen.IntSet[int]
	test.True(t, s.IsEmpty())
	test.False(t, s.Has(0))
}

func TestIntSetNew(t *testing.T) {
	t.Parallel()

	s := frozen.NewIntSet[int]()
	test.RequireTrue(t, s.IsEmpty())
	// test.False(t, s.Has(0))
	// test.False(t, s.Has(1))

	s = frozen.NewIntSet(1)
	test.RequireFalse(t, s.IsEmpty())
	// test.False(t, s.Has(0))
	test.True(t, s.Has(1), "%+v", s)

	s = frozen.NewIntSet(1, 2, 3, 4, 5)
	test.RequireFalse(t, s.IsEmpty())
	// test.False(t, s.Has(0))
	// test.True(t, s.Has(1))
	// test.True(t, s.Has(2))
	// test.True(t, s.Has(3))
	// test.True(t, s.Has(4))
	// test.True(t, s.Has(5))
}

func TestIntSetNewHuge(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		test.True(t, set.Has(i), i)
	}
}

func TestIntSetIter(t *testing.T) {
	t.Parallel()

	set := frozen.NewIntSet(1)

	container := []int{}
	for i := set.Range(); i.Next(); {
		container = append(container, i.Value())
	}

	// An extra pass to validate repeatability
	for i := set.Range(); i.Next(); { //nolint:revive
	}

	test.Equal(t, set.Count(), len(container))
}

func TestIntSetIterLarge(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())

	container := make([]int, 0, hugeCollectionSize())
	for i := set.Range(); i.Next(); {
		container = append(container, i.Value())
	}

	// An extra pass to validate repeatability
	for i := set.Range(); i.Next(); { //nolint:revive
	}

	_, set2 := generateIntArrayAndSet(hugeCollectionSize())
	test.True(t, set.EqualSet(set2), "%+v\n%+v", set, set2)
	distinct := getDistinctInts(arr)
	sort.Slice(distinct, func(i, j int) bool { return distinct[i] < distinct[j] })
	sort.Slice(container, func(i, j int) bool { return container[i] < container[j] })
	test.Equal(t, distinct, container)
}

func TestIntSetHas(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		test.True(t, set.Has(i))
	}
	test.False(t, set.Has(-1))
}

func TestIntSetWith(t *testing.T) {
	t.Parallel()

	arr, _ := generateIntArrayAndSet(hugeCollectionSize())
	set := frozen.NewIntSet(arr[:len(arr)/2]...)

	for _, i := range arr[len(arr)/2:] {
		if !test.False(t, set.Has(i), i) {
			break
		}
	}

	for _, i := range arr[len(arr)/2:] {
		set = set.With(i)
	}

	for _, i := range arr[len(arr)/2:] {
		if !test.True(t, set.Has(i), "%v %v %v", i, set, arr) {
			break
		}
	}
	test.Equal(t, len(getDistinctInts(arr)), set.Count())
}

func TestIntSetWithout(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	left, right := arr[:len(arr)/2], arr[len(arr)/2:]
	wo := set
	for _, i := range left {
		wo = wo.Without(i)
	}
	expectedCount := len(getDistinctInts(arr)) - len(getDistinctInts(left))
	test.Equal(t, expectedCount, wo.Count(), "%v\n%+v\n%v\n%v\n%v", arr, set.String(), left, right, wo)
	for _, i := range left {
		if !test.False(t, wo.Has(i), "%v\n%v\n%v\n%v\n%v\n%v", i, arr, set, left, right, wo) {
			break
		}
	}
	for _, i := range right {
		if !test.True(t, wo.Has(i),
			"i = %v\narr = %v\nset = %v\nleft = %v\nright = %v\nwo = %v",
			i, arr, set, left, right, wo,
		) {
			break
		}
	}
}

func TestIntSetIntersection(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	firstQuartile := frozen.NewIntSet(arr[:len(arr)/4]...)
	secondToFifthDecile := frozen.NewIntSet(arr[len(arr)/5 : len(arr)/2]...)
	thirdQuartile := frozen.NewIntSet(arr[len(arr)/2 : 3*len(arr)/4]...)

	intersect := fullSet.Intersection(firstQuartile)
	test.True(t, intersect.EqualSet(firstQuartile))
	test.Equal(t, firstQuartile.Count(), intersect.Count())

	intersect = firstQuartile.Intersection(thirdQuartile)
	test.True(t, intersect.IsEmpty())
	test.Equal(t, 0, intersect.Count())

	intersect = secondToFifthDecile.Intersection(firstQuartile)
	distinctNums := len(getDistinctInts(arr[len(arr)/5 : len(arr)/4]))
	for i := intersect.Range(); i.Next(); {
		test.True(t, secondToFifthDecile.Has(i.Value()) && firstQuartile.Has(i.Value()))
	}
	test.Equal(t, distinctNums, intersect.Count())
}

func TestIntSetUnion(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	set := frozen.NewIntSet(arr[:len(arr)/2]...)
	distinct := getDistinctInts(arr)
	for _, i := range arr[len(arr)/2:] {
		if !test.False(t, set.Has(i), i) {
			break
		}
	}

	union := set.Union(fullSet)
	for _, i := range distinct {
		if !test.True(t, union.Has(i), i) {
			break
		}
	}
	test.Equal(t, len(distinct), union.Count())

	union = fullSet.Union(set)
	for _, i := range distinct {
		if !test.True(t, union.Has(i), i) {
			break
		}
	}
	test.Equal(t, len(distinct), union.Count())
}

func TestIntSetIsSubsetOf(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	test.True(t, frozen.NewIntSet[int]().IsSubsetOf(fullSet))
	test.True(t, fullSet.IsSubsetOf(fullSet))
	test.True(t, frozen.NewIntSet(arr[:len(arr)/2]...).IsSubsetOf(fullSet))
	test.False(t, frozen.NewIntSet(arr[:len(arr)/2]...).IsSubsetOf(frozen.NewIntSet(arr[len(arr)/3:]...)))
}

func TestIntSetWhere(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	var evens, odds []int
	evenPred := func(e int) bool { return e%2 == 0 }
	oddPred := func(e int) bool { return e%2 != 0 }

	for _, i := range arr {
		if evenPred(i) {
			evens = append(evens, i)
		} else {
			odds = append(odds, i)
		}
	}

	evenSet := frozen.NewIntSet(evens...)
	oddSet := frozen.NewIntSet(odds...)

	evenPredWhere := fullSet.Where(evenPred)
	oddPredWhere := fullSet.Where(oddPred)
	test.True(t, evenPredWhere.EqualSet(evenSet))
	test.Equal(t, evenSet.Count(), evenPredWhere.Count())
	test.True(t, oddPredWhere.EqualSet(oddSet))
	test.Equal(t, oddSet.Count(), oddPredWhere.Count())
	test.True(t, frozen.NewIntSet[int]().Where(evenPred).EqualSet(frozen.NewIntSet[int]()))
}

func TestIntSetMap(t *testing.T) {
	t.Parallel()

	subtract := func(e int) int { return e - 1 }
	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	mappedArr := make([]int, 0, len(arr))

	for _, i := range arr {
		mappedArr = append(mappedArr, subtract(i))
	}

	mappedSet := frozen.NewIntSet(mappedArr...)

	test.True(t, mappedSet.EqualSet(fullSet.Map(subtract)))
	test.True(t, frozen.NewIntSet[int]().EqualSet(frozen.NewIntSet[int]().Map(subtract)))
}

func TestSetOfIntSet(t *testing.T) {
	t.Parallel()

	// this fails
	test.NoPanic(t, func() {
		frozen.NewSet(frozen.NewIntSet(1, 2, 3), frozen.NewIntSet(4, 5, 6))
	})

	// this succeeds
	test.NoPanic(t, func() {
		frozen.NewSet(frozen.NewSet(1, 2, 3), frozen.NewSet(4, 5, 6))
	})
}

func generateIntArrayAndSet(maxLen int) ([]int, frozen.IntSet[int]) {
	arr := make([]int, 0, maxLen)
	curr := float64(1.0)
	multiplier := math.Pow(2, 64/1e6)
	for i := 0; i < maxLen; i++ {
		arr = append(arr, int(curr))
		curr *= multiplier
	}

	seen := make(map[int]bool, len(arr))
	out := arr[:0]
	for _, e := range arr {
		if !seen[e] {
			out = append(out, e)
			seen[e] = true
		}
	}
	set := frozen.NewIntSet(out...)
	rand.Shuffle(len(out), func(i, j int) {
		a := out
		a[i], a[j] = a[j], a[i]
	})
	return out, set
}

func getDistinctInts(x []int) []int {
	m := make(map[int]byte)
	for _, i := range x {
		m[i] = 'a'
	}

	distinct := make([]int, 0, len(m))
	for k := range m {
		distinct = append(distinct, k)
	}
	return distinct
}
