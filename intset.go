package frozen

import (
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
	return i / blockBits, i % blockBits / cellBits, cellMask(1) << (i % cellBits)
}

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

func (s IntSet) IsEmpty() bool {
	return s.data.IsEmpty()
}

func (s IntSet) Count() int {
	return s.count
}

func (s IntSet) Range() IntIterator {
	return &intSetIterator{blockIter: s.data.Range()}
}

func (s IntSet) Elements() []int {
	result := make([]int, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// func (s IntSet) OrderedElements(less Less) []int {}

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
// func (s IntSet) String() string                       {}
// func (s IntSet) Format(state fmt.State, _ rune)       {}
// func (s IntSet) OrderedRange(less Less) Iterator      {}
// func (s IntSet) Hash(seed uintptr) uintptr            {}
// func (s IntSet) Equal(t int) bool                     {}
func (s IntSet) EqualSet(t IntSet) bool {
	return s.data.Equal(t.data)
}

func (s IntSet) IsSubsetOf(t IntSet) bool {
	for sBlock := s.data.Range(); sBlock.Next(); {
		if tBlock, exists := t.data.Get(sBlock.Key()); !exists || tBlock != sBlock.Value() {
			return false
		}
	}
	return true
}

func (s IntSet) Has(val int) bool {
	block, _, cellIndex, bitMask := s.locate(val)
	return block[cellIndex]&bitMask != 0
}

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

// func (s IntSet) Where(pred func(elem int) bool) IntSet    {}
// func (s IntSet) Map(f func(elem int) int) IntSet          {}
// func (s IntSet) Reduce(reduce func(elems ...int) int) int {}
// func (s IntSet) Reduce2(reduce func(a, b int) int) int    {}
func (s IntSet) Intersection(t IntSet) IntSet {
	var intersectMap MapBuilder
	count := 0
	for tBlock := t.data.Range(); tBlock.Next(); {
		if sBlock, exists := s.data.Get(tBlock.Key()); exists {
			intersectBlock := sBlock.(cellBlock).intersection(tBlock.Value().(cellBlock))
			intersectMap.Put(tBlock.Key(), intersectBlock)
			count += intersectBlock.count()
		}
	}
	return IntSet{data: intersectMap.Finish(), count: count}
}

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
