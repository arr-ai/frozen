// Generated by gen-kv.pl. DO NOT EDIT.
package kvt

import (
	"github.com/arr-ai/frozen/pkg/kv"
)

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

func (i *leafIterator) Value() kv.KeyValue {
	return i.l[i.index]
}
