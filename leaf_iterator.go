package frozen

type exhaustedIterator struct{}

func (exhaustedIterator) Next() bool {
	return false
}

func (exhaustedIterator) Value() interface{} {
	panic("empty")
}

type leafIterator struct {
	l      *leaf
	index  int
	extras extraLeafElems
}

func newLeafIterator(l *leaf) leafIterator {
	return leafIterator{
		l:     l,
		index: -1,
	}
}

func (i *leafIterator) Next() bool {
	if i.index == int(i.l.lastIndex) {
		return false
	}
	if i.index == leafElems-2 && i.l.lastIndex > leafElems {
		i.index++
		i.extras = (*i.l.last()).(extraLeafElems)
	}
	i.index++
	return true
}

func (i *leafIterator) Value() interface{} {
	return *i.elem()
}

func (i *leafIterator) elem() *interface{} { //nolint:gocritic
	if i.index < leafElems {
		return &i.l.elems[i.index]
	}
	return &i.extras[i.index-leafElems]
}
