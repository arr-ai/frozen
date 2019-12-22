package lazy

import (
	"sync/atomic"
	"unsafe"

	"github.com/marcelocantos/frozen"
)

type memoSet struct {
	set *Set
}

func memo(src Set) *memoSet {
	if m, ok := src.(*memoSet); ok {
		return m
	}
	result := &memoSet{}
	result.set = &src
	return result
}

func (s *memoSet) pointer() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(&s.set))
}

func (s *memoSet) getSet() Set {
	return *(*Set)(atomic.LoadPointer(s.pointer()))
}

func (s *memoSet) Range() SetIterator {
	if f, ok := s.getSet().(*frozenSet); ok {
		return f.set.Range()
	}
	return &cachingIterator{iter: s.getSet().Range(), s: s}
}

type cachingIterator struct {
	iter SetIterator
	s    *memoSet
	seen frozen.SetBuilder
}

func (i *cachingIterator) Next() bool {
	for i.iter.Next() {
		if val := i.iter.Value(); !i.seen.Has(val) {
			i.seen.Add(val)
			return true
		}
	}
	seen := Frozen(i.seen.Finish())
	atomic.StorePointer(i.s.pointer(), unsafe.Pointer(&seen))
	return false
}

func (i *cachingIterator) Value() interface{} {
	return i.iter.Value()
}

func (s *memoSet) IsEmpty() bool                         { return s.getSet().IsEmpty() }
func (s *memoSet) FastIsEmpty() (empty, ok bool)         { return s.getSet().FastIsEmpty() }
func (s *memoSet) Count() int                            { return s.getSet().Count() }
func (s *memoSet) FastCount() (count int, ok bool)       { return s.getSet().FastCount() }
func (s *memoSet) CountUpTo(limit int) int               { return s.getSet().CountUpTo(limit) }
func (s *memoSet) Freeze() Set                           { return s.getSet().Freeze() }
func (s *memoSet) Hash(seed uintptr) uintptr             { return s.getSet().Hash(seed) }
func (s *memoSet) Equal(set interface{}) bool            { return s.getSet().Equal(set) }
func (s *memoSet) EqualSet(set Set) bool                 { return s.getSet().EqualSet(set) }
func (s *memoSet) IsSubsetOf(set Set) bool               { return s.getSet().IsSubsetOf(set) }
func (s *memoSet) Has(el interface{}) bool               { return s.getSet().Has(el) }
func (s *memoSet) FastHas(el interface{}) (has, ok bool) { return s.getSet().FastHas(el) }
func (s *memoSet) With(els ...interface{}) Set           { return s.getSet().With(els...) }
func (s *memoSet) Without(els ...interface{}) Set        { return s.getSet().Without(els...) }
func (s *memoSet) Where(pred Predicate) Set              { return s.getSet().Where(pred) }
func (s *memoSet) Map(m Mapper) Set                      { return s.getSet().Map(m) }
func (s *memoSet) Union(set Set) Set                     { return s.getSet().Union(set) }
func (s *memoSet) Intersection(set Set) Set              { return s.getSet().Intersection(set) }
func (s *memoSet) Difference(set Set) Set                { return s.getSet().Difference(set) }
func (s *memoSet) SymmetricDifference(set Set) Set       { return s.getSet().SymmetricDifference(set) }
func (s *memoSet) Powerset() Set                         { return s.getSet().Powerset() }

func (s *memoSet) FastCountUpTo(limit int) (count int, ok bool) {
	return s.getSet().FastCountUpTo(limit)
}
