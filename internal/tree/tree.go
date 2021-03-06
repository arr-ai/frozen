package tree

import (
	"container/heap"
	"fmt"
	"math/bits"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
)

func packedIteratorBuf(count int) [][]node {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]node, 0, depth)
}

type Tree struct {
	root  node
	count int
}

func newTree(n node, count int) (out Tree) {
	return Tree{root: n, count: count}
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
		defer vet(func() { t.Combine(args, u) }, &t, &u)(&out)
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

func (t Tree) Difference(args *EqArgs, u Tree) (out Tree) {
	if vetting {
		defer vet(func() { t.Difference(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return t
	}
	root, matches := t.root.Difference(args, u.root, 0)
	return newTree(root, t.count-matches)
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
		defer vet(func() { t.Intersection(args, u) }, &t, &u)(&out)
	}
	if t.root == nil || u.root == nil {
		return Tree{}
	}
	if t.count > u.count {
		t, u = u, t
		args = args.Flip()
	}

	return newTree(t.root.Intersection(args, u.root, 0))
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
		defer vet(func() { t.Map(args, f) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	return newTree(t.root.Map(args, 0, f))
}

func (t Tree) Reduce(args NodeArgs, r func(values ...elementT) elementT) elementT {
	if t.root == nil {
		return zero
	}
	return t.root.Reduce(args, 0, r)
}

func (t Tree) Vet() {
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

func (t Tree) Where(args *WhereArgs) (out Tree) {
	if vetting {
		defer vet(func() { t.Where(args) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	return newTree(t.root.Where(args, 0))
}

func (t Tree) With(args *CombineArgs, v elementT) (out Tree) {
	if vetting {
		defer vet(func() { t.With(args, v) }, &t)(&out)
	}
	if t.root == nil {
		return Tree{root: newLeaf1(v), count: 1}
	}
	h := newHasher(v, 0)
	root, matches := t.root.With(args, v, 0, h)
	return newTree(root, t.count+1-matches)
}

func (t Tree) Without(args *EqArgs, v elementT) (out Tree) {
	if vetting {
		defer vet(func() { t.Without(args, v) }, &t)(&out)
	}
	if t.root == nil {
		return t
	}
	h := newHasher(v, 0)
	root, matches := t.root.Without(args, v, 0, h)
	return newTree(root, t.count-matches)
}

func (t Tree) clone() Tree {
	return Tree{root: t.root.clone(), count: t.count}
}
