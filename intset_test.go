package frozen_test

import (
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

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

	var s IntSet
	assert.True(t, s.IsEmpty())
	assert.False(t, s.Has(0))
}

func TestNewIntSet(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		assert.True(t, set.Has(i), i)
	}
}

func TestIntSetIter(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	container := make([]int, 0, hugeCollectionSize())
	for i := set.Range(); i.Next(); {
		container = append(container, i.Value())
	}
	distinct := getDistinctInts(arr)
	sort.Ints(distinct)
	sort.Ints(container)
	assert.Equal(t, distinct, container)
}

func TestIntSetHas(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	for _, i := range arr {
		assert.True(t, set.Has(i))
	}
	assert.False(t, set.Has(arr[len(arr)-1]-1))
}

func TestIntSetWith(t *testing.T) {
	t.Parallel()

	arr, _ := generateIntArrayAndSet(hugeCollectionSize())
	set := NewIntSet(arr[:len(arr)/2]...)

	for _, i := range arr[len(arr)/2:] {
		assert.False(t, set.Has(i))
	}

	set = set.With(arr[len(arr)/2:]...)

	for _, i := range arr[len(arr)/2:] {
		assert.True(t, set.Has(i))
	}
	assert.Equal(t, len(getDistinctInts(arr)), set.Count())
}

func TestIntSetWithout(t *testing.T) {
	t.Parallel()

	arr, set := generateIntArrayAndSet(hugeCollectionSize())
	left, right := arr[:len(arr)/2], arr[len(arr)/2:]
	set = set.Without(left...)
	expectedCount := len(getDistinctInts(arr)) - len(getDistinctInts(left))
	assert.Equal(t, expectedCount, set.Count())
	for _, i := range left {
		assert.False(t, set.Has(i))
	}
	for _, i := range right {
		assert.True(t, set.Has(i))
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
		assert.False(t, set.Has(i))
	}

	union := set.Union(fullSet)
	for _, i := range distinct {
		assert.True(t, union.Has(i))
	}
	assert.Equal(t, len(distinct), union.Count())

	union = fullSet.Union(set)
	for _, i := range distinct {
		assert.True(t, union.Has(i))
	}
	assert.Equal(t, len(distinct), union.Count())
}

func TestIntSetIsSubsetOf(t *testing.T) {
	t.Parallel()

	arr, fullSet := generateIntArrayAndSet(hugeCollectionSize())
	assert.True(t, NewIntSet().IsSubsetOf(fullSet))
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
	assert.True(t, NewIntSet().Where(evenPred).EqualSet(NewIntSet()))
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
	assert.True(t, NewIntSet().EqualSet(NewIntSet().Map(subtract)))
}

func generateIntArrayAndSet(maxLen int) ([]int, IntSet) {
	arr := make([]int, 0, maxLen)
	curr := float64(1.0)
	multiplier := math.Pow(2, 64/1e6)
	for i := 0; i < maxLen; i++ {
		arr = append(arr, int(curr))
		curr *= multiplier
	}
	return arr, NewIntSet(arr...)
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
