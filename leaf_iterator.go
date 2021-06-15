package frozen

type leafIterator struct {
	l     leaf
	index int
}

func newLeafIterator(l leaf) *leafIterator {
	return &leafIterator{
		l:     l,
		index: -1,
	}
}

func (i *leafIterator) Next() bool {
	i.index++
	return i.index < len(i.l)
}

func (i *leafIterator) Value() interface{} {
	return i.l[i.index]
}

// func (i *leafIterator) elem() *interface{} { //nolint:gocritic
// 	return &i.l[i.index]
// }
