// Generated by gen-kv.pl kvt kv.KeyValue. DO NOT EDIT.
package kvt

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

func (i *leafIterator) Value() elementT {
	return i.l[i.index]
}
