package tree

import (
	"container/heap"
	"fmt"

	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
)

type Tree struct {
	root  node
	count int
}

func newTree(n node, count *int) (out Tree) {
	return Tree{root: n, count: *count}
}

func newTreeNeg(n node, count *int) (out Tree) {
	return Tree{root: n, count: -*count}
}

func (t Tree) Count() int {
	return t.count
}

func (t Tree) Gauge() depth.Gauge {
	return depth.NewGauge(t.count)
}

func (t Tree) String() string {
	if t.root == nil {
		return "∅"
	}
	return t.root.String()
}

func (t Tree) Format(f fmt.State, verb rune) {
	if t.root == nil {
		fu.WriteString(f, "∅")
	}
	t.root.Format(f, verb)
}

func (t Tree) Combine(args *CombineArgs, u Tree) (out Tree) {
	if vetting {
		defer vet(func() { t.Combine(args, u) }, t.root, u.root)(&out.root)
	}
	if t.root == nil {
		return u
	}
	if u.root == nil {
		return t
	}
	count := -(t.count + u.count)
	return newTreeNeg(t.root.Combine(args, u.root, 0, &count), &count)
}

func (t Tree) Difference(args *EqArgs, u Tree) (out Tree) {
	if vetting {
		defer vet(func() { t.Difference(args, u) }, t.root, u.root)(&out.root)
	}
	if t.root == nil || u.root == nil {
		return t
	}
	count := -t.count
	return newTreeNeg(t.root.Difference(args, u.root, 0, &count), &count)
}

func (t Tree) Equal(args *EqArgs, u Tree) bool {
	if t.count != u.count {
		return false
	}
	if t.root == nil {
		return u.root == nil
	}
	if u.root == nil {
		return false
	}
	return t.root.Equal(args, u.root, 0)
}

func (t Tree) Get(args *EqArgs, v elementT) *elementT {
	if t.root == nil {
		return nil
	}
	h := newHasher(v, 0)
	return t.root.Get(args, v, h)
}

func (t Tree) Intersection(args *EqArgs, u Tree) (out Tree) {
	if vetting {
		defer vet(func() { t.Intersection(args, u) }, t.root, u.root)(&out.root)
	}
	if t.root == nil || u.root == nil {
		return Tree{}
	}
	if t.count > u.count {
		t, u = u, t
		args = args.flip
	}
	count := 0
	return newTree(t.root.Intersection(args, u.root, 0, &count), &count)
}

func (t Tree) Iterator() Iterator {
	if t.root == nil {
		return emptyIterator
	}
	buf := packedIteratorBuf(t.count)
	return t.root.Iterator(buf)
}

func (t Tree) OrderedIterator(less Less, n int) Iterator {
	if n < 0 || n > t.count {
		n = t.count
	}
	if n == 0 {
		return emptyIterator
	}
	o := &ordered{less: less, elements: make([]elementT, 0, n)}
	for i := t.root.Iterator(packedIteratorBuf(t.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO(o)
	heap.Init(r)
	return r.(Iterator)
}

func (t Tree) SubsetOf(args *EqArgs, u Tree) bool {
	if t.root == nil {
		return true
	}
	if u.root == nil {
		return false
	}
	return t.root.SubsetOf(args, u.root, 0)
}

func (t Tree) Map(args *CombineArgs, f func(v elementT) elementT) (out Tree) {
	if vetting {
		defer vet(func() { t.Map(args, f) }, t.root)(&out.root)
	}
	if t.root == nil {
		return t
	}
	count := 0
	return newTree(t.root.Map(args, 0, &count, f), &count)
}

func (t Tree) Reduce(args NodeArgs, r func(values ...elementT) elementT) elementT {
	if t.root == nil {
		return zero
	}
	return t.root.Reduce(args, 0, r)
}

func (t Tree) Where(args *WhereArgs) (out Tree) {
	if vetting {
		defer vet(func() { t.Where(args) }, t.root)(&out.root)
	}
	if t.root == nil {
		return t
	}
	count := 0
	return newTree(t.root.Where(args, 0, &count), &count)
}

func (t Tree) With(args *CombineArgs, v elementT) (out Tree) {
	if vetting {
		defer vet(func() { t.With(args, v) }, t.root)(&out.root)
	}
	if t.root == nil {
		return Tree{root: newLeaf1(v), count: 1}
	}
	count := -(t.count + 1)
	h := newHasher(v, 0)
	return newTreeNeg(t.root.With(args, v, 0, h, &count), &count)
}

func (t Tree) Without(args *EqArgs, v elementT) (out Tree) {
	if vetting {
		defer vet(func() { t.Without(args, v) }, t.root)(&out.root)
	}
	if t.root == nil {
		return t
	}
	count := -t.count
	h := newHasher(v, 0)
	return newTreeNeg(t.root.Without(args, v, 0, h, &count), &count)
}
