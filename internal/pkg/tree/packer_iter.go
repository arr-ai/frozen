package tree

import (
	"container/heap"

	"github.com/arr-ai/frozen/internal/pkg/masker"
)

// Less dictates the order of two elements.
type Less[T any] func(a, b T) bool

type packerIterator[T any] struct {
	mask  masker.Masker
	data  *[fanout]node[T]
	stack [][]node[T]
	i     Iterator[T]
}

func newPackerIterator[T any](buf [][]node[T], p *packer[T]) *packerIterator[T] {
	return &packerIterator[T]{mask: p.mask, data: &p.data, stack: buf, i: Empty[T]()}
}

func (i *packerIterator[T]) Next() bool {
	if i.i.Next() {
		return true
	}
	for i.mask != 0 {
		idx := i.mask.FirstIndex()
		i.mask = i.mask.Next()
		c := i.data[idx]
		if !c.Empty() {
			i.i = c.Iterator(i.stack)
			return i.i.Next()
		}
	}
	return false
}

func (i *packerIterator[T]) Value() T {
	return i.i.Value()
}

type ordered[T any] struct {
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

func (o *ordered[T]) Push(x any) {
	o.elements = append(o.elements, x.(T))
}

func (o *ordered[T]) Pop() any {
	result := o.elements[len(o.elements)-1]
	o.elements = o.elements[:len(o.elements)-1]
	return result
}

type reverseOrdered[T any] struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	heap.Interface
	val T
}

// Reverse returns the reverse order for data.
func reverseO[T any](data heap.Interface) heap.Interface {
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
