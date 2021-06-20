package kvt

import (
	"container/heap"

	"github.com/arr-ai/frozen/internal/iterator/kvi"
	"github.com/arr-ai/frozen/pkg/kv"
)

// Less dictates the order of two elements.
type Less func(a, b kv.KeyValue) bool

type packerIterator struct {
	stack []packer
	i     kvi.Iterator
}

func newPackerIterator(buf []packer, p packer) *packerIterator {
	buf = append(buf, p)
	return &packerIterator{stack: buf, i: kvi.Empty}
}

func (i *packerIterator) Next() bool {
	if i.i.Next() {
		return true
	}
	p := &i.stack[0]
	if len(p.data) == 0 {
		return false
	}
	i.i = p.data[0].Iterator(i.stack[1:])
	p.data = p.data[1:]
	return i.i.Next()
}

func (i *packerIterator) Value() kv.KeyValue {
	return i.i.Value()
}

type ordered struct {
	less     Less
	elements []kv.KeyValue
	val      kv.KeyValue
}

func (o *ordered) Next() bool {
	if o.Len() == 0 {
		return false
	}
	return true
}

func (o *ordered) Value() kv.KeyValue {
	return o.val
}

func (o *ordered) Len() int {
	return len(o.elements)
}

func (o *ordered) Less(i, j int) bool {
	return o.less(o.elements[j], o.elements[i])
}

func (o *ordered) Swap(i, j int) {
	o.elements[i], o.elements[j] = o.elements[j], o.elements[i]
}

func (o *ordered) Push(x interface{}) {
}

func (o *ordered) Pop() interface{} {
	result := o.elements[len(o.elements)-1]
	o.elements = o.elements[:len(o.elements)-1]
	return result
}

type reverseOrdered struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	heap.Interface
	val kv.KeyValue
}

// Reverse returns the reverse order for data.
func reverseO(data heap.Interface) heap.Interface {
	return &reverseOrdered{Interface: data}
}

func (r *reverseOrdered) Next() bool {
	if r.Len() == 0 {
		return false
	}
	return true
}

func (r *reverseOrdered) Value() kv.KeyValue {
	return r.val
}

// Less returns the opposite of the embedded implementation's Less method.
func (r *reverseOrdered) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}
