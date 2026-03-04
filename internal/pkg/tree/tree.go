package tree

import (
	"container/heap"
	"fmt"
	"math/bits"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
)

func packedIteratorBuf[T any](count int) [][]node[T] {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]node[T], 0, depth)
}

type Tree[T any] struct {
	root  node[T]
	count int
	built bool // true after Builder.Finish() or immutable ops; enables h0 vetting
}

func newTree[T any](n node[T], count int) (out Tree[T]) {
	return Tree[T]{root: n, count: count}
}

func (t Tree[T]) Count() int {
	return t.count
}

func (t Tree[T]) H0() H128 {
	if t.root == nil {
		return H128{}
	}
	return t.root.H0()
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
		return
	}
	t.root.Format(f, verb)
}

func (t Tree[T]) Combine(args *CombineArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Combine(args, u) }, &t, &u)(&out)
	}
	if t.root == nil {
		return Tree[T]{root: u.root, count: u.count, built: true}
	}
	if u.root == nil {
		return Tree[T]{root: t.root, count: t.count, built: true}
	}
	root, matches := t.root.Combine(args, u.root, 0)
	return Tree[T]{root: root, count: t.count + u.count - matches, built: true}
}

func (t Tree[T]) Difference(args *EqArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Difference(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return t
	}
	root, matches := t.root.Difference(args, u.root, 0)
	return Tree[T]{root: root, count: t.count - matches, built: true}
}

func (t Tree[T]) Equal(args *EqArgs[T], u Tree[T]) bool {
	switch {
	case t.count != u.count:
		return false
	case t.count == 0 && u.count == 0:
		return true
	case t.root.H0() != u.root.H0():
		return false
	case args.FullHash() && !t.root.H0().isZero():
		return true
	default:
		return t.root.Equal(args, u.root, 0)
	}
}

func (t Tree[T]) Get(v T) *T {
	if t.root == nil {
		return nil
	}
	args := DefaultNPEqArgs[T]()
	h := newHasherWith(v, 0, args.hf)
	return t.root.Get(args, v, h, 0)
}

func (t Tree[T]) Intersection(args *EqArgs[T], u Tree[T]) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Intersection(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return Tree[T]{built: true}
	}
	if t.count > u.count {
		t, u = u, t
	}

	root, matches := t.root.Intersection(args, u.root, 0)
	return Tree[T]{root: root, count: matches, built: true}
}

func (t Tree[T]) Iterator() Iterator[T] {
	if t.root == nil {
		return Empty[T]()
	}
	buf := packedIteratorBuf[T](t.count)
	return t.root.Iterator(buf)
}

func (t Tree[T]) OrderedIterator(less Less[T], n int) Iterator[T] {
	if n < 0 || n > t.count {
		n = t.count
	}
	if n == 0 {
		return Empty[T]()
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
	return r.(Iterator[T])
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

func Map[T, U any](t Tree[T], f func(v T) U) (out Tree[U]) {
	if vetting {
		defer vet[U](func() { Map(t, f) }, &t)(&out)
	}
	if t.root == nil {
		return
	}
	var b Builder[U]
	for i := t.Iterator(); i.Next(); {
		b.Add(f(i.Value()))
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
			panic(fmt.Errorf("empty root count > 0 (%d)", t.count))
		}
	} else {
		count := t.root.Vet()
		if count != t.count {
			panic(fmt.Errorf("count mismatch: measured (%d) != tracked (%d)", count, t.count))
		}
		if t.built {
			vetNodeH0(t.root, GetHashFunc[T]())
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
	root, matches := t.root.Where(args, 0)
	return Tree[T]{root: root, count: matches, built: true}
}

func (t Tree[T]) With(v T) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.With(v) }, &t)(&out)
	}
	hf := GetHashFunc[T]()
	if t.root == nil {
		return Tree[T]{root: newLeaf1WithHash(v, newElemH128(v, hf)), count: 1, built: true}
	}
	h := newHasherWith(v, 0, hf)
	if b, ok := t.root.(*branch[T]); ok {
		root, matches := withFastBatched(b, v, h, hf)
		return Tree[T]{root: root, count: t.count + 1 - matches, built: true}
	}
	root, matches := t.root.WithFast(v, 0, h)
	return Tree[T]{root: root, count: t.count + 1 - matches, built: true}
}

func (t Tree[T]) Without(v T) (out Tree[T]) {
	if vetting {
		defer vet[T](func() { t.Without(v) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	args := DefaultNPEqArgs[T]()
	h := newHasherWith(v, 0, args.hf)
	if b, ok := t.root.(*branch[T]); ok {
		root, matches := withoutBatched(b, args, v, h)
		return Tree[T]{root: root, count: t.count - matches, built: true}
	}
	root, matches := t.root.Without(args, v, 0, h)
	return Tree[T]{root: root, count: t.count - matches, built: true}
}

func vetNodeH0[T any](n node[T], hf func(T) H128) {
	switch n := n.(type) {
	case *leaf1[T]:
		n.vetH0(hf)
	case *leaf2[T]:
		n.vetH0(hf)
	case *leaf[T]:
		n.vetH0(hf)
	case *branch[T]:
		n.vetH0(hf)
	}
}

func (t Tree[T]) clone() Tree[T] {
	return Tree[T]{root: t.root.clone(), count: t.count, built: t.built}
}
