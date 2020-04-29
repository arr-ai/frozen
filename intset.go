package frozen

import (
	"strconv"
	"strings"
	"unsafe"
)

type IntSet struct {
	data  Map
	count int
}

type cellMask uintptr

const (
	blockCells = 8 // Tune.
	cellBits   = int(8 * unsafe.Sizeof(cellMask(0)))
	blockBits  = cellBits * blockCells
)

type cellBlock [blockCells]cellMask

var emptyBlock cellBlock

func locateBlock(i int) (blockIndex, cellIndex int, bitMask cellMask) {
	return i / blockBits, i % blockBits / cellBits, cellMask(1) << uint(i%cellBits)
}

// NewIntSet returns an IntSet with the values provided.
func NewIntSet(is ...int) IntSet {
	var b MapBuilder
	count := 0
	prevBlockIndex := int(^uint(0) >> 1) // maxint
	var block cellBlock
	blockIsFilled := false
	for _, i := range is {
		blockIndex, cellIndex, bitMask := locateBlock(i)
		if blockIndex != prevBlockIndex {
			if blockIsFilled {
				b.Put(prevBlockIndex, block)
			}
			prevBlockIndex = blockIndex
			var v interface{}
			if v, blockIsFilled = b.Get(blockIndex); blockIsFilled {
				block = v.(cellBlock)
			} else {
				block = emptyBlock
			}
		}
		if block[cellIndex]&bitMask == 0 {
			block[cellIndex] |= bitMask
			count++
			blockIsFilled = true
		}
	}
	if blockIsFilled {
		b.Put(prevBlockIndex, block)
	}
	return IntSet{data: b.Finish(), count: count}
}

// IsEmpty returns true if there is no values in s and false otherwise.
func (s IntSet) IsEmpty() bool {
	return s.data.IsEmpty()
}

// Count returns the number of elements in IntSet.
func (s IntSet) Count() int {
	return s.count
}

// Range returns the iterator for IntSet.
func (s IntSet) Range() IntIterator {
	return &intSetIterator{blockIter: s.data.Range()}
}

// Elements returns all the values of IntSet.
func (s IntSet) Elements() []int {
	result := make([]int, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// func (s IntSet) OrderedElements(less Less) []int {}

// Any returns a random value from s.
func (s IntSet) Any() int {
	k, v := s.data.Any()
	blockIndex := k.(int)
	block := v.(cellBlock)
	for cellIndex, cell := range block {
		if cell != 0 {
			bit := BitIterator(cell).Index()
			return blockBits*blockIndex + cellBits*cellIndex + bit
		}
	}
	panic("empty block")
}

// func (s IntSet) AnyN(n int) IntSet                    {}
// func (s IntSet) OrderedFirstN(n int, less Less) []int {}
// func (s IntSet) First(less Less) int                  {}
// func (s IntSet) FirstN(n int, less Less) IntSet       {}

// String returns a string representation of IntSet.
func (s IntSet) String() string {
	stringed := make([]string, 0, s.count)
	for i := s.Range(); i.Next(); {
		stringed = append(stringed, strconv.Itoa(i.Value()))
	}
	return "[" + strings.Join(stringed, ", ") + "]"
}

// func (s IntSet) Format(state fmt.State, _ rune)       {}
// func (s IntSet) OrderedRange(less Less) Iterator      {}
// func (s IntSet) Hash(seed uintptr) uintptr            {}
// func (s IntSet) Equal(t int) bool                     {}

// EqualSet returns true if both IntSets are equal.
func (s IntSet) EqualSet(t IntSet) bool {
	return s.data.Equal(t.data)
}

// IsSubsetOf returns true if s is a subset of t and false otherwise.
func (s IntSet) IsSubsetOf(t IntSet) bool {
	for sBlock := s.data.Range(); sBlock.Next(); {
		if tBlock, exists := t.data.Get(sBlock.Key()); !exists || tBlock != sBlock.Value() {
			return false
		}
	}
	return true
}

// Has returns true if value exists in the IntSet and false otherwise.
func (s IntSet) Has(val int) bool {
	block, _, cellIndex, bitMask := s.locate(val)
	return block[cellIndex]&bitMask != 0
}

// With returns a new IntSet with the values of s and the provided values.
func (s IntSet) With(is ...int) IntSet {
	for _, i := range is {
		block, blockIndex, cellIndex, bitMask := s.locate(i)
		if block[cellIndex]&bitMask == 0 {
			block[cellIndex] |= bitMask
			s.data = s.data.With(blockIndex, block)
			s.count++
		}
	}
	return s
}

// Without returns an IntSet without the provided values.
func (s IntSet) Without(is ...int) IntSet {
	indexToRemove := NewSetBuilder(0)
	for _, i := range is {
		block, blockIndex, cellIndex, bitMask := s.locate(i)
		if block[cellIndex]&bitMask != 0 {
			block[cellIndex] &^= bitMask
			//TODO: optimize this so it doesn't do With many times
			s.data = s.data.With(blockIndex, block)
			if block == emptyBlock {
				indexToRemove.Add(blockIndex)
			}
			s.count--
		}
	}
	s.data = s.data.Without(indexToRemove.Finish())
	return s
}

// Where returns an IntSet whose values fulfill the provided condition.
func (s IntSet) Where(pred func(elem int) bool) IntSet {
	//TODO: find a way that works more on block level or maybe make IntSetBuilder?
	arr := make([]int, 0, s.count)
	for i := s.Range(); i.Next(); {
		if pred(i.Value()) {
			arr = append(arr, i.Value())
		}
	}
	return NewIntSet(arr...)
}

// Map returns an IntSet with whose values are mapped from s.
func (s IntSet) Map(f func(elem int) int) IntSet {
	arr := make([]int, 0, s.count)
	for i := s.Range(); i.Next(); {
		arr = append(arr, f(i.Value()))
	}
	return NewIntSet(arr...)
}

// func (s IntSet) Reduce(reduce func(elems ...int) int) int {}
// func (s IntSet) Reduce2(reduce func(a, b int) int) int    {}

// Intersection returns an IntSet whose values exists in s and t.
func (s IntSet) Intersection(t IntSet) IntSet {
	var intersectMap MapBuilder
	count := 0
	for tBlock := t.data.Range(); tBlock.Next(); {
		if sBlock, exists := s.data.Get(tBlock.Key()); exists {
			intersectBlock := sBlock.(cellBlock).intersection(tBlock.Value().(cellBlock))
			if intersectBlock != emptyBlock {
				intersectMap.Put(tBlock.Key(), intersectBlock)
				count += intersectBlock.count()
			}
		}
	}
	return IntSet{data: intersectMap.Finish(), count: count}
}

// Union returns an integer set that is a union of s and t.
func (s IntSet) Union(t IntSet) IntSet {
	unionMap := s.data
	count := s.count
	var unionBlock cellBlock
	for tBlock := t.data.Range(); tBlock.Next(); {
		if sBlock, exists := s.data.Get(tBlock.Key()); exists {
			unionBlock = sBlock.(cellBlock).union(tBlock.Value().(cellBlock))
			count += unionBlock.diffCount(sBlock.(cellBlock))
		} else {
			unionBlock = tBlock.Value().(cellBlock)
			count += unionBlock.count()
		}
		unionMap = unionMap.With(tBlock.Key(), unionBlock)
	}
	return IntSet{data: unionMap, count: count}
}

// func (s IntSet) Difference(t IntSet) IntSet               {}
// func (s IntSet) SymmetricDifference(t IntSet) IntSet      {}
// func (s IntSet) Powerset() IntSet                         {}
// func (s IntSet) GroupBy(key func(el int) int) Map         {}

func (s IntSet) locate(i int) (block cellBlock, blockIndex, cellIndex int, bitMask cellMask) {
	blockIndex, cellIndex, bitMask = locateBlock(i)
	if v, has := s.data.Get(blockIndex); has {
		block = v.(cellBlock)
	}
	return
}

func (b cellBlock) intersection(b2 cellBlock) (intersect cellBlock) {
	for i := range b {
		intersect[i] = b[i] & b2[i]
	}
	return
}

func (b cellBlock) union(b2 cellBlock) (u cellBlock) {
	for i := range b {
		u[i] = b[i] | b2[i]
	}
	return
}

func (b cellBlock) diffCount(b2 cellBlock) (diffCount int) {
	for i := range b {
		diffCount += b[i].diffCount(b2[i])
	}
	return
}

func (b cellBlock) count() (count int) {
	for _, c := range b {
		count += c.onesCount()
	}
	return
}

func (c cellMask) diffCount(c2 cellMask) int { return (c ^ c2).onesCount() }

func (c cellMask) onesCount() (count int) {
	for ; c != 0; c &= c - 1 {
		count++
	}
	return
}
