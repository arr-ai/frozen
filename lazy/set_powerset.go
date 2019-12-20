package lazy

import (
	"math/bits"
	"unsafe"
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
			if count < limit {
				return count, true
			}
			return limit, true
		}
		panic("Count(): too many elements")
	}
	return 0, false
}

func (s *powerSet) Has(el interface{}) bool {
	ss, ok := el.(Set)
	return ok && ss.IsSubsetOf(s.set)
}

func (s *powerSet) Range() SetIterator {
	return &powerSetSetIterator{i: s.set.Range()}
}

type powerSetSetIterator struct {
	i     SetIterator
	end   uint64
	mask  uint64
	pool  []interface{}
	value Set
}

func (i *powerSetSetIterator) Next() bool {
	oldMask := i.mask
	i.mask++
	if i.mask >= i.end {
		if i.mask > i.end {
			panic("called Next() after reaching the end")
		}
		if !i.i.Next() {
			return false
		}
		i.pool = append(i.pool, i.i.Value())
		i.end <<= 1
	}
	for del := oldMask & ^i.mask; del != 0; del &= del - 1 {
		i.value = i.value.Without(i.pool[bits.TrailingZeros64(del)])
	}
	for add := i.mask & ^oldMask; add != 0; add &= add - 1 {
		i.value = i.value.With(i.pool[bits.TrailingZeros64(add)])
	}
	return true
}

func (i *powerSetSetIterator) Value() interface{} {
	return i.value
}
