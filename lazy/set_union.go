package lazy

type unionSet struct {
	baseSet
	a, b Set
}

func union(a, b Set) Set {
	s := &unionSet{a: a, b: b}
	s.baseSet.set = s
	return memo(s)
}

func (s *unionSet) FastCountUpTo(max int) (count int, ok bool) {
	if empty, ok := s.a.FastIsEmpty(); ok && empty {
		return s.b.CountUpTo(max), true
	}
	if empty, ok := s.b.FastIsEmpty(); ok && empty {
		return s.a.CountUpTo(max), true
	}
	if count, ok := s.a.FastCountUpTo(max); ok && count == max {
		return count, true
	}
	if count, ok := s.b.FastCountUpTo(max); ok && count == max {
		return count, true
	}
	return 0, false
}

func (s *unionSet) Has(el interface{}) bool {
	return s.a.Has(el) || s.b.Has(el)
}

func (s *unionSet) FastHas(el interface{}) (has, ok bool) {
	if has, ok = s.a.FastHas(el); ok {
		return has, ok
	}
	return s.a.FastHas(el)
}

func (s *unionSet) Range() SetIterator {
	return &unionSetIterator{i: s.a.Range(), b: s.b}
}

type unionSetIterator struct {
	i SetIterator
	b Set
}

func (i *unionSetIterator) Next() bool {
	for {
		if i.i.Next() {
			return true
		}
		if i.b == nil {
			return false
		}
		i.i = i.b.Range()
		i.b = nil
	}
}

func (i *unionSetIterator) Value() interface{} {
	return i.i.Value()
}
