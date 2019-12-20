package lazy

import (
	"github.com/marcelocantos/frozen"
	"github.com/marcelocantos/hash"
)

const (
	maxInt   = int(^uint(0) >> 1)
	hashSeed = uintptr(624409645898692063)
)

type baseSet struct {
	set Set
}

func (s *baseSet) IsEmpty() bool {
	if empty, ok := s.set.FastIsEmpty(); ok {
		return empty
	}
	return !s.set.Range().Next()
}

func (s *baseSet) FastIsEmpty() (has, ok bool) {
	return false, false
}

func (s *baseSet) Count() int {
	if count, ok := s.set.FastCount(); ok {
		return count
	}
	return s.set.CountUpTo(maxInt)
}

func (s *baseSet) FastCount() (count int, ok bool) {
	return 0, false
}

func (s *baseSet) CountUpTo(limit int) int {
	if count, ok := s.set.FastCountUpTo(limit); ok {
		return count
	}
	n := 0
	for i := s.set.Range(); n < limit && i.Next(); n++ {
	}
	return n
}

func (s *baseSet) FastCountUpTo(limit int) (count int, ok bool) {
	return 0, false
}

func (s *baseSet) Freeze() frozen.Set {
	m := memo(s.set)
	for i := m.Range(); i.Next(); {
	}
	return m.Set.(*frozenSet).set
}

func (s *baseSet) Equal(set interface{}) bool {
	if set, ok := set.(Set); ok {
		return s.EqualSet(set)
	}
	return false
}

func (s *baseSet) EqualSet(set Set) bool {
	return s.set.EqualSet(set)
}

func (s *baseSet) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(hashSeed, seed)
	for i := s.set.Range(); i.Next(); {
		h = hash.Interface(i.Value(), h)
	}
	return h
}

func (s *baseSet) Has(el interface{}) bool {
	if has, ok := s.set.FastHas(el); ok {
		return has
	}
	for i := s.set.Range(); i.Next(); {
		if frozen.Equal(el, i.Value()) {
			return true
		}
	}
	return false
}

func (s *baseSet) FastHas(el interface{}) (has, ok bool) {
	return false, false
}

func (s *baseSet) IsSubsetOf(set Set) bool {
	return s.set.IsSubsetOf(set)
}

func (s *baseSet) Where(pred Predicate) Set {
	return where(s.set, pred)
}

func (s *baseSet) With(els ...interface{}) Set {
	return union(s.set, Frozen(frozen.NewSet(els...)))
}

func (s *baseSet) Without(els ...interface{}) Set {
	return difference(s.set, Frozen(frozen.NewSet(els...)))
}

func (s *baseSet) Map(m Mapper) Set {
	return mapper(s.set, m)
}

func (s *baseSet) Union(set Set) Set {
	return union(s.set, set)
}

func (s *baseSet) Intersection(set Set) Set {
	return intersection(s.set, set)
}

func (s *baseSet) Difference(set Set) Set {
	return difference(s.set, set)
}

func (s *baseSet) SymmetricDifference(set Set) Set {
	return symmetricDifference(s.set, set)
}

func (s *baseSet) PowerSet() Set {
	return powerset(s.set)
}
