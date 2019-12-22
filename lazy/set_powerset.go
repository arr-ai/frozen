package lazy

import (
	"math/bits"
	"unsafe"

	"github.com/marcelocantos/frozen"
)

type powerSet struct {
	baseSet
	set Set
}

func powerset(set Set) Set {
	s := &powerSet{set: set}
	s.baseSet.set = s
	return memo(s)
}

func (s *powerSet) IsEmpty() bool {
	return s.set.IsEmpty()
}

func (s *powerSet) FastIsEmpty() (empty, ok bool) {
	return s.set.FastIsEmpty()
}

func (s *powerSet) Count() int {
	if count := s.set.Count(); count < int(8*unsafe.Sizeof(0)) {
		return 1 << count
	}
	panic("Count(): too many elements")
}

func (s *powerSet) FastCount() (count int, ok bool) {
	if count, ok := s.set.FastCount(); ok {
		if count < int(8*unsafe.Sizeof(0)) {
			return 1 << count, true
		}
		panic("Count(): too many elements")
	}
	return 0, false
}

func (s *powerSet) FastCountUpTo(limit int) (count int, ok bool) {
	if count, ok := s.set.FastCount(); ok {
		if count < int(8*unsafe.Sizeof(0)) {
			n := 1 << count
			if n < limit {
				return n, true
			}
			return limit, true
		}
		panic("Count(): too many elements")
	}
	return 0, false
}

func (s *powerSet) Has(el interface{}) bool {
	ss, ok := el.(frozen.Set)
	return ok && Frozen(ss).IsSubsetOf(s.set)
}

func (s *powerSet) Range() SetIterator {
	return &powerSetSetIterator{
		i:    s.set.Range(),
		end:  1,
		mask: ^uint64(0),
	}
}

type powerSetSetIterator struct {
	i     SetIterator
	end   uint64
	mask  uint64
	elems []interface{}
	value frozen.Set
}

func (i *powerSetSetIterator) Next() bool {
	i.mask++
	if i.mask >= i.end {
		if i.mask > i.end {
			panic("called Next() after reaching the end")
		}
		if !i.i.Next() {
			return false
		}
		i.elems = append(i.elems, i.i.Value())
		i.end <<= 1
	}
	// Use a special counting order that flips one bit at at time. See
	// (frozen.Set).Powerset() for a detailed explanation.
	if flip := bits.TrailingZeros64(i.mask); flip < len(i.elems) {
		if i.mask>>(flip+1)&1 == 0 {
			i.value = i.value.With(i.elems[flip])
		} else {
			i.value = i.value.Without(i.elems[flip])
		}
	}
	return true
}

func (i *powerSetSetIterator) Value() interface{} {
	return Frozen(i.value)
}
