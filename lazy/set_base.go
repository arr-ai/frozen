package lazy

import (
	"github.com/arr-ai/frozen"
	"github.com/marcelocantos/hash"
)

const (
	maxInt   = int(^uint(0) >> 1)
	hashSeed = uintptr(uint64(624409645898692063) & uint64(^uintptr(0)))
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

func (s *baseSet) Freeze() Set {
	if m, ok := memo(s.set).(*memoSet); ok {
		for i := m.Range(); i.Next(); {
		}
		return m.getSet()
	}
	return s.set
}

func (s *baseSet) Equal(set interface{}) bool {
	if set, ok := set.(Set); ok {
		return s.set.EqualSet(set)
	}
	return false
}

func (s *baseSet) EqualSet(t Set) bool {
	return s.Freeze().Equal(t.Freeze())
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

func (s *baseSet) IsSubsetOf(t Set) bool {
	return s.Freeze().IsSubsetOf(t.Freeze())
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

func (s *baseSet) Union(t Set) Set {
	return union(s.set, t)
}

func (s *baseSet) Intersection(t Set) Set {
	return intersection(s.set, t)
}

func (s *baseSet) Difference(t Set) Set {
	return difference(s.set, t)
}

func (s *baseSet) SymmetricDifference(t Set) Set {
	return symmetricDifference(s.set, t)
}

func (s *baseSet) Powerset() Set {
	return powerset(s.set)
}
