package frozen_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	. "github.com/arr-ai/frozen"
)

func hugeCollectionSize() int {
	if testing.Short() {
		return 50_000
	}
	return 1_000_000
}

func TestIntSetEmpty(t *testing.T) {
	t.Parallel()

	var s IntSet[int]
	assert.True(t, s.IsEmpty())
	assert.False(t, s.Has(0))
}

func TestIntSetNew(t *testing.T) {
	t.Parallel()

	s := NewIntSet[int]()
	require.True(t, s.IsEmpty())
	// assert.False(t, s.Has(0))
	// assert.False(t, s.Has(1))

	s = NewIntSet(1)
	require.False(t, s.IsEmpty())
	// assert.False(t, s.Has(0))
	assert.True(t, s.Has(1), "%+v", s)

	s = NewIntSet(1, 2, 3, 4, 5)
	require.False(t, s.IsEmpty())
	// assert.False(t, s.Has(0))
	// assert.True(t, s.Has(1))
	// assert.True(t, s.Has(2))
	// assert.True(t, s.Has(3))
	// assert.True(t, s.Has(4))
	// assert.True(t, s.Has(5))
}

func TestIntSetNewHuge(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		assert.True(t, set.Has(i), i)
	}
}

func TestIntSetIter(t *testing.T) {
	t.Parallel()

	set := NewIntSet(1)

	container := []int{}
	for i := set.Range(); i.Next(); {
		container = append(container, i.Value())
	}

	// An extra pass to validate repeatability
	for i := set.Range(); i.Next(); { //nolint:revive
	}

	assert.Equal(t, set.Count(), len(container))
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
	assert.True(t, set.EqualSet(set2), "%+v\n%+v", set, set2)
	distinct := getDistinctInts(arr)
	slices.Sort(distinct)
	slices.Sort(container)
	assert.Equal(t, distinct, container)
}

func TestIntSetHas(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		assert.True(t, set.Has(i))
	}
	assert.False(t, set.Has(-1))
}

func TestIntSetWith(t *testing.T) {
	t.Parallel()

	arr, _ := generateIntArrayAndSet(hugeCollectionSize())
	set := NewIntSet(arr[:len(arr)/2]...)

	for _, i := range arr[len(arr)/2:] {
		if !assert.False(t, set.Has(i), i) {
			break
		}
	}

	for _, i := range arr[len(arr)/2:] {
		set = set.With(i)
	}

	for _, i := range arr[len(arr)/2:] {
		if !assert.True(t, set.Has(i), "%v %v %v", i, set, arr) {
			break
		}
	}
	assert.Equal(t, len(getDistinctInts(arr)), set.Count())
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
	assert.Equal(t, expectedCount, wo.Count(), "%v\n%+v\n%v\n%v\n%v", arr, set.String(), left, right, wo)
	for _, i := range left {
		if !assert.False(t, wo.Has(i), "%v\n%v\n%v\n%v\n%v\n%v", i, arr, set, left, right, wo) {
			break
		}
	}
	for _, i := range right {
		if !assert.True(t, wo.Has(i),
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
	firstQuartile := NewIntSet(arr[:len(arr)/4]...)
	secondToFifthDecile := NewIntSet(arr[len(arr)/5 : len(arr)/2]...)
	thirdQuartile := NewIntSet(arr[len(arr)/2 : 3*len(arr)/4]...)

	intersect := fullSet.Intersection(firstQuartile)
	assert.True(t, intersect.EqualSet(firstQuartile))
	assert.Equal(t, firstQuartile.Count(), intersect.Count())

	intersect = firstQuartile.Intersection(thirdQuartile)
	assert.True(t, intersect.IsEmpty())
	assert.Equal(t, 0, intersect.Count())

	intersect = secondToFifthDecile.Intersection(firstQuartile)
	distinctNums := len(getDistinctInts(arr[len(arr)/5 : len(arr)/4]))
	for i := intersect.Range(); i.Next(); {
		assert.True(t, secondToFifthDecile.Has(i.Value()) && firstQuartile.Has(i.Value()))
	}
	assert.Equal(t, distinctNums, intersect.Count())
}

func TestIntSetUnion(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	set := NewIntSet(arr[:len(arr)/2]...)
	distinct := getDistinctInts(arr)
	for _, i := range arr[len(arr)/2:] {
		if !assert.False(t, set.Has(i), i) {
			break
		}
	}

	union := set.Union(fullSet)
	for _, i := range distinct {
		if !assert.True(t, union.Has(i), i) {
			break
		}
	}
	assert.Equal(t, len(distinct), union.Count())

	union = fullSet.Union(set)
	for _, i := range distinct {
		if !assert.True(t, union.Has(i), i) {
			break
		}
	}
	assert.Equal(t, len(distinct), union.Count())
}

func TestIntSetIsSubsetOf(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	assert.True(t, NewIntSet[int]().IsSubsetOf(fullSet))
	assert.True(t, fullSet.IsSubsetOf(fullSet))
	assert.True(t, NewIntSet(arr[:len(arr)/2]...).IsSubsetOf(fullSet))
	assert.False(t, NewIntSet(arr[:len(arr)/2]...).IsSubsetOf(NewIntSet(arr[len(arr)/3:]...)))
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

	evenSet := NewIntSet(evens...)
	oddSet := NewIntSet(odds...)

	evenPredWhere := fullSet.Where(evenPred)
	oddPredWhere := fullSet.Where(oddPred)
	assert.True(t, evenPredWhere.EqualSet(evenSet))
	assert.Equal(t, evenSet.Count(), evenPredWhere.Count())
	assert.True(t, oddPredWhere.EqualSet(oddSet))
	assert.Equal(t, oddSet.Count(), oddPredWhere.Count())
	assert.True(t, NewIntSet[int]().Where(evenPred).EqualSet(NewIntSet[int]()))
}

func TestIntSetMap(t *testing.T) {
	t.Parallel()

	subtract := func(e int) int { return e - 1 }
	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	mappedArr := make([]int, 0, len(arr))

	for _, i := range arr {
		mappedArr = append(mappedArr, subtract(i))
	}

	mappedSet := NewIntSet(mappedArr...)

	assert.True(t, mappedSet.EqualSet(fullSet.Map(subtract)))
	assert.True(t, NewIntSet[int]().EqualSet(NewIntSet[int]().Map(subtract)))
}

func TestSetOfIntSet(t *testing.T) {
	t.Parallel()

	// this fails
	assert.NotPanics(t, func() {
		NewSet(NewIntSet(1, 2, 3), NewIntSet(4, 5, 6))
	})

	// this succeeds
	assert.NotPanics(t, func() {
		NewSet(NewSet(1, 2, 3), NewSet(4, 5, 6))
	})
}

func generateIntArrayAndSet(maxLen int) ([]int, IntSet[int]) {
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
	set := NewIntSet(out...)
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
