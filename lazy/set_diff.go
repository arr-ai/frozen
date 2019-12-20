package lazy

type differenceSet struct {
	baseSet
	a, b Set
}

func difference(a, b Set) Set {
	if empty, ok := a.FastIsEmpty(); ok && empty {
		return a
	}
	if empty, ok := b.FastIsEmpty(); ok && empty {
		return a
	}
	s := &differenceSet{a: a, b: b}
	s.baseSet.set = s
	return memo(s)
}

func (s *differenceSet) Has(el interface{}) bool {
	return s.a.Has(el) || s.b.Has(el)
}

func (s *differenceSet) FastHas(el interface{}) (has, ok bool) {
	if has, ok = s.a.FastHas(el); ok {
		return
	}
	return s.a.FastHas(el)
}

func (s *differenceSet) Range() SetIterator {
	return &differenceSetIterator{i: s.a.Range(), b: s.b}
}

type differenceSetIterator struct {
	i SetIterator
	b Set
}

func (i *differenceSetIterator) Next() bool {
	for {
		if !i.i.Next() {
			return false
		}
		if !i.b.Has(i.i.Value()) {
			return true
		}
	}
}

func (i *differenceSetIterator) Value() interface{} {
	return i.i.Value()
}
