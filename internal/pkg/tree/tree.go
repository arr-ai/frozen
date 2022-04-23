package tree

import (
	"container/heap"
	"fmt"
	"math/bits"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/iterator"
	"github.com/arr-ai/frozen/pkg/errors"
)

func packedIteratorBuf[T any](count int) [][]node[T] {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]node[T], 0, depth)
}

type Tree[T any] struct {
	root  node[T]
	count int
}

func newTree[T any](n node[T], count int) (out Tree[T]) {
	return Tree[T]{root: n, count: count}
}

func (t Tree[T]) Count() int {
	return t.count
}

func (t Tree[T]) Gauge() depth.Gauge {
	return depth.NewGauge(t.count)
}

func (t Tree[T]) String() string {
	if t.root == nil {
		return "∅"
	}
	return t.root.String()
}

func (t Tree[T]) Format(f fmt.State, verb rune) {
	if t.root == nil {
		fu.WriteString(f, "∅")
	}
	t.root.Format(f, verb)
}

func (t Tree[T]) Combine(args *CombineArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Combine(args, u) }, &t, &u)(&out)
	}
	if t.root == nil {
		return u
	}
	if u.root == nil {
		return t
	}
	root, matches := t.root.Combine(args, u.root, 0)
	return newTree(root, t.count+u.count-matches)
}

func (t Tree[T]) Difference(args *EqArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Difference(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return t
	}
	root, matches := t.root.Difference(args, u.root, 0)
	return newTree(root, t.count-matches)
}

func (t Tree[T]) Equal(args *EqArgs[T], u Tree[T]) bool {
	switch {
	case t.count != u.count:
		return false
	case t.count == 0 && u.count == 0:
		return true
	default:
		return t.root.Equal(args, u.root, 0)
	}
}

func (t Tree[T]) Get(args *EqArgs[T], v T) *T {
	if t.root == nil {
		return nil
	}
	h := newHasher(v, 0)
	return t.root.Get(args, v, h)
}

func (t Tree[T]) Intersection(args *EqArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Intersection(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return Tree[T]{}
	}
	if t.count > u.count {
		t, u = u, t
	}

	return newTree(t.root.Intersection(args, u.root, 0))
}

func (t Tree[T]) Iterator() iterator.Iterator[T] {
	if t.root == nil {
		return iterator.Empty[T]()
	}
	buf := packedIteratorBuf[T](t.count)
	return t.root.Iterator(buf)
}

func (t Tree[T]) OrderedIterator(less Less[T], n int) iterator.Iterator[T] {
	if n < 0 || n > t.count {
		n = t.count
	}
	if n == 0 {
		return iterator.Empty[T]()
	}
	o := &ordered[T]{less: less, elements: make([]T, 0, n)}
	for i := t.root.Iterator(packedIteratorBuf[T](t.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO[T](o)
	heap.Init(r)
	return r.(iterator.Iterator[T])
}

func (t Tree[T]) SubsetOf(args *EqArgs[T], u Tree[T]) bool {
	if t.root == nil {
		return true
	}
	if u.root == nil {
		return false
	}
	return t.root.SubsetOf(args, u.root, 0)
}

func TreeMap[T, U any](t Tree[T], args *CombineArgs[U], f func(v T) U) (out Tree[U]) {
	if vetting {
		defer vet[U](func() { TreeMap(t, args, f) }, &t)(&out)
	}
	if t.root == nil {
		return
	}
	var b Builder[U]
	for i := t.Iterator(); i.Next(); {
		b.add(args, f(i.Value()))
	}
	return b.Finish()
}

func (t Tree[T]) Reduce(args NodeArgs, r func(values ...T) T) (_ T, _ bool) {
	if t.root != nil {
		return t.root.Reduce(args, 0, r), true
	}
	return
}

func (t Tree[T]) Vet() {
	if t.root == nil {
		if t.count != 0 {
			panic(errors.Errorf("empty root count > 0 (%d)", t.count))
		}
	} else {
		count := t.root.Vet()
		if count != t.count {
			panic(errors.Errorf("count mismatch: measured (%d) != tracked (%d)", count, t.count))
		}
	}
}

func (t Tree[T]) Where(args *WhereArgs[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Where(args) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	return newTree(t.root.Where(args, 0))
}

func (t Tree[T]) With(v T) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.With(v) }, &t)(&out)
	}
	if t.root == nil {
		return Tree[T]{root: newLeaf1(v), count: 1}
	}
	h := newHasher(v, 0)
	root, matches := t.root.FastWith(v, 0, h)
	return newTree(root, t.count+1-matches)
}

func (t Tree[T]) Without(args *EqArgs[T], v T) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Without(args, v) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	h := newHasher(v, 0)
	root, matches := t.root.Without(args, v, 0, h)
	return newTree(root, t.count-matches)
}

func (t Tree[T]) clone() Tree[T] {
	return Tree[T]{root: t.root.clone(), count: t.count}
}
