package frozen

import (
	"fmt"
	"math/bits"

	"github.com/arr-ai/hash"
	"golang.org/x/exp/constraints"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	internalIterator "github.com/arr-ai/frozen/internal/pkg/iterator"
)

type IntSet[I constraints.Integer] struct {
	data  Map[I, cellMask]
	count int
}

type cellMask uint64

const (
	cellShift = 6
	cellBits  = 1 << cellShift
)

func locateCell[I constraints.Integer](i I) (cell I, bitMask cellMask) {
	return i >> cellShift, cellMask(1) << uint(i%cellBits)
}

// NewIntSet returns an IntSet with the values provided.
func NewIntSet[I constraints.Integer](is ...I) IntSet[I] {
	m := map[I]cellMask{}
	for _, i := range is {
		cellIndex, bitMask := locateCell(i)
		m[cellIndex] = m[cellIndex] | bitMask
	}
	b := NewMapBuilder[I, cellMask](len(m))
	count := 0
	for cellIndex, cell := range m {
		b.Put(cellIndex, cell)
		count += cell.count()
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
func (s IntSet[I]) Range() Iterator[I] {
	return &intSetIterator[I]{
		cellIter: s.data.Range(),
		cell:     1,
	}
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
	cellIndex, cell := s.data.Any()
	bit := internalIterator.BitIterator(cell).Index()
	return cellIndex<<cellShift + I(bit)
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

func (s IntSet[I]) Hash(seed uintptr) uintptr {
	for i := s.Range(); i.Next(); {
		seed ^= hash.Interface(i.Value(), 0)
	}
	return seed
}

// EqualSet is deprecated. Use Equal instead.
func (s IntSet[I]) EqualSet(t IntSet[I]) bool {
	return s.Equal(t)
}

// Equal returns true if both IntSets are equal.
func (s IntSet[I]) Equal(t IntSet[I]) bool {
	return s.data.Equal(t.data)
}

// Equal returns true if t is an IntSet and a and b are equal.
func (s IntSet[I]) Same(t any) bool {
	if t, is := t.(IntSet[I]); is {
		return s.Equal(t)
	}
	return false
}

// IsSubsetOf returns true if s is a subset of t and false otherwise.
func (s IntSet[I]) IsSubsetOf(t IntSet[I]) bool {
	for r := s.data.Range(); r.Next(); {
		sCell := r.Value()
		if tCell, has := t.data.Get(r.Key()); !has || sCell&^tCell != 0 {
			return false
		}
	}
	return true
}

// Has returns true if value exists in the IntSet and false otherwise.
func (s IntSet[I]) Has(val I) bool {
	cell, _, bitMask := s.locate(val)
	return cell&bitMask != 0
}

// With returns a new IntSet with the values of s and the provided values.
func (s IntSet[I]) With(i I) IntSet[I] {
	if cell, cellIndex, bitMask := s.locate(i); cell&bitMask == 0 {
		s.data = s.data.With(cellIndex, cell|bitMask)
		s.count++
	}
	return s
}

// Without returns an IntSet without the provided values.
func (s IntSet[I]) Without(i I) IntSet[I] {
	if cell, cellIndex, bitMask := s.locate(i); cell&bitMask != 0 {
		if cell &^= bitMask; cell != 0 {
			// TODO: optimize this so it doesn't do With many times
			s.data = s.data.With(cellIndex, cell)
		} else {
			s.data = s.data.Without(cellIndex)
		}
		s.count--
	}
	return s
}

// Where returns an IntSet whose values fulfill the provided condition.
func (s IntSet[I]) Where(pred func(elem I) bool) IntSet[I] {
	// TODO: find a way that works more on block level or maybe make IntSetBuilder?
	var b MapBuilder[I, cellMask]
	count := 0
	for r := s.data.Range(); r.Next(); {
		cellIndex, cell := r.Entry()
		cellOffset := cellIndex * cellBits
		var newCell cellMask
		for cell != 0 {
			maskIndex := I(cell.index())
			if pred(cellOffset + maskIndex) {
				newCell |= cellMask(1) << maskIndex
			}
			cell &= (cell - 1)
		}
		if newCell != 0 {
			b.Put(cellIndex, newCell)
			count += newCell.count()
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
	return NewIntSet(arr...)
}

// Intersection returns an IntSet whose values exists in s and t.
func (s IntSet[I]) Intersection(t IntSet[I]) IntSet[I] {
	var intersectMap MapBuilder[I, cellMask]
	count := 0
	for tCell := t.data.Range(); tCell.Next(); {
		if sCell, has := s.data.Get(tCell.Key()); has {
			if iCell := sCell & tCell.Value(); iCell != 0 {
				intersectMap.Put(tCell.Key(), iCell)
				count += iCell.count()
			}
		}
	}
	return IntSet[I]{data: intersectMap.Finish(), count: count}
}

// Union returns an integer set that is a union of s and t.
func (s IntSet[I]) Union(t IntSet[I]) IntSet[I] {
	unionMap := s.data
	count := s.count
	var unionCell cellMask
	for tCell := t.data.Range(); tCell.Next(); {
		if sCell, has := s.data.Get(tCell.Key()); has {
			unionCell = sCell | tCell.Value()
			count += unionCell.diffCount(sCell)
		} else {
			unionCell = tCell.Value()
			count += unionCell.count()
		}
		unionMap = unionMap.With(tCell.Key(), unionCell)
	}
	return IntSet[I]{data: unionMap, count: count}
}

func (s IntSet[I]) locate(i I) (cell cellMask, cellIndex I, bitMask cellMask) {
	cellIndex, bitMask = locateCell(i)
	cell, _ = s.data.Get(cellIndex)
	return
}

func (c cellMask) diffCount(c2 cellMask) int {
	return (c ^ c2).count()
}

func (c cellMask) count() int {
	return bits.OnesCount64(uint64(c))
}

func (c cellMask) index() int {
	return bits.TrailingZeros64(uint64(c))
}

func (c cellMask) next() cellMask {
	return c & (c - 1)
}
