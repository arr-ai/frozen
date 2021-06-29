package tree

import (
	"container/heap"
)

// Less dictates the order of two elements.
type Less func(a, b elementT) bool

type packerIterator struct {
	stack [][]node
	i     Iterator
}

func newPackerIterator(buf [][]node, p *packer) *packerIterator {
	buf = append(buf, p.data[:])
	// TODO: Speed up with mask.
	return &packerIterator{stack: buf, i: emptyIterator}
}

func (i *packerIterator) Next() bool {
	if i.i.Next() {
		return true
	}
	p := &i.stack[0]
	for len(*p) > 0 {
		c := (*p)[0]
		*p = (*p)[1:]
		if c != nil && !c.Empty() {
			i.i = c.Iterator(i.stack[1:])
			return i.i.Next()
		}
	}
	return false
}

func (i *packerIterator) Value() elementT {
	return i.i.Value()
}

type ordered struct {
	less     Less
	elements []elementT
	val      elementT
}

func (o *ordered) Next() bool {
	if o.Len() == 0 {
		return false
	}
	o.val = interfaceAsElement(heap.Pop(o))
	return true
}

func (o *ordered) Value() elementT {
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
	o.elements = append(o.elements, interfaceAsElement(x))
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
	val elementT
}

// Reverse returns the reverse order for data.
func reverseO(data heap.Interface) heap.Interface {
	return &reverseOrdered{Interface: data}
}

func (r *reverseOrdered) Next() bool {
	if r.Len() == 0 {
		return false
	}
	r.val = interfaceAsElement(heap.Pop(r))
	return true
}

func (r *reverseOrdered) Value() elementT {
	return r.val
}

// Less returns the opposite of the embedded implementation's Less method.
func (r *reverseOrdered) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}
