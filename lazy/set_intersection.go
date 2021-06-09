package lazy

type intersectionSet struct {
	baseSet
	a, b Set
}

func intersection(a, b Set) Set {
	if isEmpty, ok := a.FastIsEmpty(); ok && isEmpty {
		return EmptySet{}
	}
	if isEmpty, ok := b.FastIsEmpty(); ok && isEmpty {
		return EmptySet{}
	}
	s := &intersectionSet{a: a, b: b}
	s.baseSet.set = s
	return memo(s)
}

func (s *intersectionSet) Has(el interface{}) bool {
	return s.a.Has(el) && s.b.Has(el)
}

func (s *intersectionSet) FastHas(el interface{}) (has, ok bool) {
	aHas, aOk := s.a.FastHas(el)
	if aOk && !aHas {
		return false, true
	}
	bHas, bOk := s.b.FastHas(el)
	if bOk && !bHas {
		return false, true
	}
	return aHas && bHas, aOk && bOk
}

func (s *intersectionSet) Range() SetIterator {
	return &intersectionSetIterator{i: s.a.Range(), b: s.b}
}

type intersectionSetIterator struct {
	i SetIterator
	b Set
}

func (i *intersectionSetIterator) Next() bool {
	for {
		if !i.i.Next() {
			return false
		}
		if i.b.Has(i.i.Value()) {
			return true
		}
	}
}

func (i *intersectionSetIterator) Value() interface{} {
	return i.i.Value()
}
