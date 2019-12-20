package lazy

type symmetricDifferenceSet struct {
	baseSet
	a, b Set
}

func symmetricDifference(a, b Set) Set {
	if empty, ok := b.FastIsEmpty(); ok && empty {
		return a
	}
	if empty, ok := a.FastIsEmpty(); ok && empty {
		return b
	}
	s := &symmetricDifferenceSet{a: a, b: b}
	s.baseSet.set = s
	return memo(s)
}

func (s *symmetricDifferenceSet) FastCountUpTo(max int) (count int, ok bool) {
	return s.baseSet.FastCountUpTo(max)
}

func (s *symmetricDifferenceSet) Has(el interface{}) bool {
	return s.a.Has(el) != s.b.Has(el)
}

func (s *symmetricDifferenceSet) FastHas(el interface{}) (has, ok bool) {
	if aHas, ok := s.a.FastHas(el); ok {
		if bHas, ok := s.b.FastHas(el); ok {
			return aHas != bHas, true
		}
	}
	return false, false
}

func (s *symmetricDifferenceSet) Range() SetIterator {
	return &symmetricDifferenceSetIterator{
		i: s.a.Range(),
		j: s.b.Range(),
		s: s,
	}
}

type symmetricDifferenceSetIterator struct {
	i, j SetIterator
	s    *symmetricDifferenceSet
}

func (i *symmetricDifferenceSetIterator) Next() bool {
	if i.i != nil {
		for {
			if !i.i.Next() {
				i.i = nil
				break
			}
			if !i.s.b.Has(i.i.Value()) {
				return true
			}
		}
	}
	for {
		if !i.j.Next() {
			return false
		}
		if !i.s.a.Has(i.j.Value()) {
			return true
		}
	}
}

func (i *symmetricDifferenceSetIterator) Value() interface{} {
	if i.i != nil {
		return i.i.Value()
	}
	return i.j.Value()
}
