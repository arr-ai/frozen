package lazy

import "github.com/marcelocantos/frozen"

type frozenSet struct {
	baseSet
	set frozen.Set
}

func Frozen(set frozen.Set) Set {
	s := &frozenSet{set: set}
	s.baseSet.set = s
	return s
}

func (s *frozenSet) FastIsEmpty() (empty, ok bool) {
	return s.set.IsEmpty(), true
}

func (s *frozenSet) FastCount() (count int, ok bool) {
	return s.set.Count(), true
}

func (s *frozenSet) FastCountUpTo(limit int) (count int, ok bool) {
	if n := s.set.Count(); n < limit {
		return n, true
	}
	return limit, true
}

func (s *frozenSet) FastHas(el interface{}) (has, ok bool) {
	return s.set.Has(el), true
}

func (s *frozenSet) EqualSet(set Set) bool {
	if f, ok := set.(*frozenSet); ok {
		return s.set.EqualSet(f.set)
	}
	n := s.set.Count()
	i := set.Range()
	for ; n > 0 && i.Next(); n-- {
		if !s.set.Has(i.Value()) {
			return false
		}
	}
	return n == 0 && !i.Next()
}

func (s *frozenSet) IsSubsetOf(set Set) bool {
	if m, ok := set.(*memoSet); ok {
		if f, ok := m.Set.(*frozenSet); ok {
			return s.IsSubsetOf(f)
		}
	}

	n := s.set.Count()
	for i := set.Range(); n > 0 && i.Next(); {
		if s.set.Has(i.Value()) {
			n--
		}
	}
	return n == 0
}

func (s *frozenSet) Range() SetIterator {
	return s.set.Range()
}

func (s *frozenSet) With(els ...interface{}) Set {
	return Frozen(s.set.With(els...))
}

func (s *frozenSet) Without(els ...interface{}) Set {
	return Frozen(s.set.Without(els...))
}

func (s *frozenSet) Map(m Mapper) Set {
	return Frozen(s.set.Map(m))
}
