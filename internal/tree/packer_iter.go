package tree

import (
	"container/heap"

	"github.com/arr-ai/frozen/internal/iterator"
)

// Less dictates the order of two elements.
type Less func(a, b interface{}) bool

type packerIterator struct {
	stack []packer
	i     iterator.Iterator
}

func newPackerIterator(buf []packer, p packer) *packerIterator {
	buf = append(buf, p)
	return &packerIterator{stack: buf, i: iterator.Empty}
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

func (i *packerIterator) Value() interface{} {
	return i.i.Value()
}

type ordered struct {
	less     Less
	elements []interface{}
	val      interface{}
}

func (o *ordered) Next() bool {
	if o.Len() == 0 {
		return false
	}
	o.val = heap.Pop(o) // SUBST %: \(o\) => (o).(%)
	return true
}

func (o *ordered) Value() interface{} {
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

func (o *ordered) Push(x actualInterface) {
	o.elements = append(o.elements, x) // SUBST %: x => x.(%)
}

func (o *ordered) Pop() actualInterface {
	result := o.elements[len(o.elements)-1]
	o.elements = o.elements[:len(o.elements)-1]
	return result
}

type reverseOrdered struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	heap.Interface
	val interface{}
}

// Reverse returns the reverse order for data.
func reverseO(data heap.Interface) heap.Interface {
	return &reverseOrdered{Interface: data}
}

func (r *reverseOrdered) Next() bool {
	if r.Len() == 0 {
		return false
	}
	r.val = heap.Pop(r) // SUBST %: \(r\) => (r).(%)
	return true
}

func (r *reverseOrdered) Value() interface{} {
	return r.val
}

// Less returns the opposite of the embedded implementation's Less method.
func (r *reverseOrdered) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}
