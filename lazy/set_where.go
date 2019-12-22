package lazy

type whereSet struct {
	baseSet
	src  Set
	pred Predicate
}

func where(set Set, pred Predicate) Set {
	s := &whereSet{src: set, pred: pred}
	s.baseSet.set = s
	return memo(s)
}

func (s *whereSet) FastIsEmpty() (empty, ok bool) {
	if empty, ok = s.src.FastIsEmpty(); ok && empty {
		return
	}
	return false, false
}

func (s *whereSet) Has(el interface{}) bool {
	return s.pred(el) && s.src.Has(el)
}

func (s *whereSet) Range() SetIterator {
	return &whereSetIterator{i: s.src.Range(), pred: s.pred}
}

type whereSetIterator struct {
	i    SetIterator
	pred Predicate
}

func (s *whereSetIterator) Next() bool {
	for s.i.Next() {
		if s.pred(s.i.Value()) {
			return true
		}
	}
	return false
}

func (s *whereSetIterator) Value() interface{} {
	return s.i.Value()
}
