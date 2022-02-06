package frozen

import (
	"fmt"
	"math/bits"

	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
)

type integer interface{
	~uint16 | ~int | ~int8
}

type IntSet[I integer] struct {
	data  Map[I, *cellBlock]
	count int
}

type cellMask uint64

const (
	blockCells = 8 // Tune.
	cellBits   = 64
	blockBits  = 512 // cellBits * blockCells
	blockShift = 9
)

func (b *cellBlock) isSubsetOf(c *cellBlock) bool {
	for i, x := range b {
		y := c[i]
		if x&^y != 0 {
			return false
		}
	}
	return true
}

func locateBlock[I integer](i I) (blockIndex, cellIndex I, bitMask cellMask) {
	return i >> blockShift, (i - i >> blockShift << blockShift) / cellBits, cellMask(1) << uint(i%cellBits)
}

// NewIntSet returns an IntSet with the values provided.
func NewIntSet[I integer](is ...I) IntSet[I] {
	m := map[I]*cellBlock{}
	for _, i := range is {
		blockIndex, cellIndex, bitMask := locateBlock(i)
		block, has := m[blockIndex]
		if !has {
			block = &cellBlock{}
			m[blockIndex] = block
		}
		block[cellIndex] |= bitMask
	}
	b := NewMapBuilder[I, *cellBlock](len(m))
	count := 0
	for blockIndex, block := range m {
		b.Put(blockIndex, block)
		count += block.count()
	}
	return IntSet[I]{data: b.Finish(), count: count}
}

// IsEmpty returns true if there is no values in s and false otherwise.
func (s IntSet[I]) IsEmpty() bool {
	return s.data.IsEmpty()
}

// Count returns the number of elements in IntSet.
func (s IntSet[I]) Count() int {
	return s.count
}

// Range returns the iterator for IntSet.
func (s IntSet[I]) Range() IntIterator[I] {
	return &intSetIterator[I]{blockIter: s.data.Range()}
}

// Elements returns all the values of IntSet.
func (s IntSet[I]) Elements() []I {
	result := make([]I, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// func (s IntSet[I]) OrderedElements(less Less) []I {}

// Any returns a random value from s.
func (s IntSet[I]) Any() I {
	blockIndex, block := s.data.Any()
	for cellIndex, cell := range block {
		if cell != 0 {
			bit := iterator.BitIterator(cell).Index()
			return blockIndex<<blockShift + I(cellBits*cellIndex) + I(bit)
		}
	}
	panic("empty block")
}

// func (s IntSet[I]) AnyN(n I) IntSet                  {}
// func (s IntSet[I]) OrderedFirstN(n I, less Less) []I {}
// func (s IntSet[I]) First(less Less) I                {}
// func (s IntSet[I]) FirstN(n I, less Less) IntSet     {}

// String returns a string representation of IntSet.
func (s IntSet[I]) String() string {
	return fu.String(s)
}

// Format formats IntSet.
func (s IntSet[I]) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fu.Fprint(f, s.data)
		return
	}

	fu.WriteString(f, "[")
	for i, r := 0, s.Range(); r.Next(); i++ {
		fu.Comma(f, i)
		fu.Format(r.Value(), f, verb)
	}
	fu.WriteString(f, "]")
}

// func (s IntSet[I]) OrderedRange(less Less) Iterator      {}
// func (s IntSet[I]) Hash(seed uint64) uint64            {}
// func (s IntSet[I]) Equal(t int) bool                     {}

// EqualSet returns true if both IntSets are equal.
func (s IntSet[I]) EqualSet(t IntSet[I]) bool {
	if s.data.Count() != t.data.Count() {
		return false
	}
	for r := s.data.Range(); r.Next(); {
		blockIndex, sBlock := r.Entry()
		tBlock, has := t.data.Get(blockIndex)
		if !has || *sBlock != *tBlock {
			return false
		}
	}
	return true
}

// IsSubsetOf returns true if s is a subset of t and false otherwise.
func (s IntSet[I]) IsSubsetOf(t IntSet[I]) bool {
	for r := s.data.Range(); r.Next(); {
		sBlock := r.Value()
		if tBlock, has := t.data.Get(r.Key()); !has || !sBlock.isSubsetOf(tBlock) {
			return false
		}
	}
	return true
}

// Has returns true if value exists in the IntSet and false otherwise.
func (s IntSet[I]) Has(val I) bool {
	block, _, cellIndex, bitMask := s.locate(val)
	return block[cellIndex]&bitMask != 0
}

// With returns a new IntSet with the values of s and the provided values.
func (s IntSet[I]) With(is ...I) IntSet[I] {
	for _, i := range is {
		block, blockIndex, cellIndex, bitMask := s.locate(i)
		if block[cellIndex]&bitMask == 0 {
			newBlock := *block
			newBlock[cellIndex] |= bitMask
			s.data = s.data.With(blockIndex, &newBlock)
			s.count++
		}
	}
	return s
}

// Without returns an IntSet without the provided values.
func (s IntSet[I]) Without(is ...I) IntSet[I] {
	indexToRemove := NewSetBuilder[I](0)
	for _, i := range is {
		block, blockIndex, cellIndex, bitMask := s.locate(i)
		if block[cellIndex]&bitMask != 0 {
			newBlock := *block
			newBlock[cellIndex] &^= bitMask
			// TODO: optimize this so it doesn't do With many times
			s.data = s.data.With(blockIndex, &newBlock)
			if newBlock == (cellBlock{}) {
				indexToRemove.Add(blockIndex)
			}
			s.count--
		}
	}
	s.data = s.data.Without(indexToRemove.Finish())
	return s
}

// Where returns an IntSet whose values fulfill the provided condition.
func (s IntSet[I]) Where(pred func(elem I) bool) IntSet[I] {
	// TODO: find a way that works more on block level or maybe make IntSetBuilder?
	var b MapBuilder[I, *cellBlock]
	count := 0
	for r := s.data.Range(); r.Next(); {
		blockIndex, block := r.Entry()
		blockOffset := blockIndex << blockShift
		newBlock := &cellBlock{}
		for cellIndex, bitMask := range block {
			cellOffset := blockOffset + I(cellIndex*cellBits)
			for bitMask != 0 {
				maskIndex := I(bitMask.index())
				if pred(cellOffset + maskIndex) {
					newBlock[cellIndex] |= cellMask(1) << maskIndex
				}
				bitMask &= (bitMask - 1)
			}
		}
		if *newBlock != (cellBlock{}) {
			b.Put(blockIndex, newBlock)
			count += newBlock.count()
		}
	}
	return IntSet[I]{data: b.Finish(), count: count}
}

// Map returns an IntSet with whose values are mapped from s.
func (s IntSet[I]) Map(f func(elem I) I) IntSet[I] {
	arr := make([]I, 0, s.count)
	for i := s.Range(); i.Next(); {
		arr = append(arr, f(i.Value()))
	}
	return NewIntSet[I](arr...)
}

// func (s IntSet[I]) Reduce(reduce func(elems ...I) I) I {}
// func (s IntSet[I]) Reduce2(reduce func(a, b I) I) I    {}

// Intersection returns an IntSet whose values exists in s and t.
func (s IntSet[I]) Intersection(t IntSet[I]) IntSet[I] {
	var intersectMap MapBuilder[I, *cellBlock]
	count := 0
	for tBlock := t.data.Range(); tBlock.Next(); {
		if sBlock, exists := s.data.Get(tBlock.Key()); exists {
			intersectBlock := sBlock.intersection(tBlock.Value())
			if intersectBlock != nil {
				intersectMap.Put(tBlock.Key(), intersectBlock)
				count += intersectBlock.count()
			}
		}
	}
	return IntSet[I]{data: intersectMap.Finish(), count: count}
}

// Union returns an integer set that is a union of s and t.
func (s IntSet[I]) Union(t IntSet[I]) IntSet[I] {
	unionMap := s.data
	count := s.count
	var unionBlock *cellBlock
	for tBlock := t.data.Range(); tBlock.Next(); {
		if sBlock, exists := s.data.Get(tBlock.Key()); exists {
			unionBlock = sBlock.union(tBlock.Value())
			count += unionBlock.diffCount(sBlock)
		} else {
			unionBlock = tBlock.Value()
			count += unionBlock.count()
		}
		unionMap = unionMap.With(tBlock.Key(), unionBlock)
	}
	return IntSet[I]{data: unionMap, count: count}
}

// func (s IntSet[I]) Difference(t IntSet) IntSet               {}
// func (s IntSet[I]) SymmetricDifference(t IntSet) IntSet      {}
// func (s IntSet[I]) Powerset() IntSet                         {}
// func (s IntSet[I]) GroupBy(key func(el int) int) Map         {}

func (s IntSet[I]) locate(i I) (block *cellBlock, blockIndex, cellIndex I, bitMask cellMask) {
	blockIndex, cellIndex, bitMask = locateBlock(i)
	if v, has := s.data.Get(blockIndex); has {
		block = v
	} else {
		block = &cellBlock{}
	}
	return
}

type cellBlock [blockCells]cellMask

func (b *cellBlock) intersection(b2 *cellBlock) *cellBlock {
	ret := *b
	for i := range b {
		ret[i] &= b2[i]
	}
	if ret == (cellBlock{}) {
		return nil
	}
	return &ret
}

func (b *cellBlock) union(b2 *cellBlock) *cellBlock {
	ret := *b
	for i := range b {
		ret[i] |= b2[i]
	}
	return &ret
}

func (b *cellBlock) diffCount(b2 *cellBlock) (diffCount int) {
	for i := range b {
		diffCount += b[i].diffCount(b2[i])
	}
	return
}

func (b *cellBlock) count() (count int) {
	for _, c := range b {
		count += c.onesCount()
	}
	return
}

func (c cellMask) diffCount(c2 cellMask) int {
	return (c ^ c2).onesCount()
}

func (c cellMask) onesCount() int {
	return bits.OnesCount64(uint64(c))
}

func (c cellMask) index() int {
	return bits.TrailingZeros64(uint64(c))
}
