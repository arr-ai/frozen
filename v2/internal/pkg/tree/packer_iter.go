package tree

import (
	"container/heap"

	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
)

// Less dictates the order of two elements.
type Less[T comparable] func(a, b T) bool

type packerIterator[T comparable] struct {
	stack [][]node[T]
	i     iterator.Iterator[T]
}

func newPackerIterator[T comparable](buf [][]node[T], p *packer[T]) *packerIterator[T] {
	buf = append(buf, p.data[:])
	// TODO: Speed up with mask.
	return &packerIterator[T]{stack: buf, i: iterator.Empty[T]()}
}

func (i *packerIterator[T]) Next() bool {
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

func (i *packerIterator[T]) Value() T {
	return i.i.Value()
}

type ordered[T comparable] struct {
	less     Less[T]
	elements []T
	val      T
}

func (o *ordered[T]) Next() bool {
	if o.Len() == 0 {
		return false
	}
	o.val = heap.Pop(o).(T)
	return true
}

func (o *ordered[T]) Value() T {
	return o.val
}

func (o *ordered[T]) Len() int {
	return len(o.elements)
}

func (o *ordered[T]) Less(i, j int) bool {
	return o.less(o.elements[j], o.elements[i])
}

func (o *ordered[T]) Swap(i, j int) {
	o.elements[i], o.elements[j] = o.elements[j], o.elements[i]
}

func (o *ordered[T]) Push(x interface{}) {
	o.elements = append(o.elements, x.(T))
}

func (o *ordered[T]) Pop() interface{} {
	result := o.elements[len(o.elements)-1]
	o.elements = o.elements[:len(o.elements)-1]
	return result
}

type reverseOrdered[T comparable] struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	heap.Interface
	val T
}

// Reverse returns the reverse order for data.
func reverseO[T comparable](data heap.Interface) heap.Interface {
	return &reverseOrdered[T]{Interface: data}
}

func (r *reverseOrdered[T]) Next() bool {
	if r.Len() == 0 {
		return false
	}
	r.val = heap.Pop(r).(T)
	return true
}

func (r *reverseOrdered[T]) Value() T {
	return r.val
}

// Less returns the opposite of the embedded implementation's Less method.
func (r *reverseOrdered[T]) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}
