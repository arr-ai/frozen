package frozen

import (
	"container/heap"
)

// Less dictates the order of two elements.
type Less func(a, b interface{}) bool

type packedIterator struct {
	stack []packed
	i     Iterator
}

func newPackedIterator(buf []packed, p packed) *packedIterator {
	buf = append(buf, p)
	return &packedIterator{stack: buf, i: emptyIterator{}}
}

func (i *packedIterator) Next() bool {
	if i.i.Next() {
		return true
	}
	p := &i.stack[0]
	if len(p.data) == 0 {
		return false
	}
	i.i = p.data[0].iterator(i.stack[1:])
	p.data = p.data[1:]
	return i.i.Next()
}

func (i *packedIterator) Value() interface{} {
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
	o.val = heap.Pop(o)
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

func (o *ordered) Push(x interface{}) {
	o.elements = append(o.elements, x)
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
	r.val = heap.Pop(r)
	return true
}

func (r *reverseOrdered) Value() interface{} {
	return r.val
}

// Less returns the opposite of the embedded implementation's Less method.
func (r *reverseOrdered) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}
