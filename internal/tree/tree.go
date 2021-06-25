package tree

import (
	"container/heap"
	"math/bits"

	"github.com/arr-ai/frozen/internal/depth"
)

type Tree struct {
	root  *node
	count int
}

func newTree(n *node, count *int) Tree {
	return Tree{root: n, count: *count}
}

func newTreeNeg(n *node, count *int) Tree {
	return Tree{root: n, count: -*count}
}

func (t Tree) Root() *node {
	if t.root == nil {
		t.root = theEmptyNode
	}
	return t.root
}

func (t Tree) MutableRoot() *node {
	if t.root == nil {
		t.root = newMutableLeaf().Node()
	}
	return t.root
}

func (t *Tree) Add(args *CombineArgs, v elementT) {
	count := -(t.count + 1)
	t.root = t.MutableRoot().Add(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t Tree) Count() int {
	return t.count
}

func (t Tree) Gauge() depth.Gauge {
	return depth.NewGauge(t.count)
}

func (t Tree) String() string {
	return t.Root().String()
}

func (t Tree) Combine(args *CombineArgs, u Tree) Tree {
	count := -(t.count + u.count)
	return newTreeNeg(t.Root().Combine(args, u.Root(), 0, &count), &count)
}

func (t Tree) Difference(args *EqArgs, u Tree) Tree {
	count := -t.count
	a := t.Root()
	b := u.Root()
	return newTreeNeg(a.Difference(args, b, 0, &count), &count)
}

func (t Tree) Equal(args *EqArgs, n Tree) bool {
	return t.Root().Equal(args, n.Root(), 0)
}

func (t Tree) Get(args *EqArgs, v elementT) *elementT {
	return t.Root().Get(args, v, newHasher(v, 0))
}

func (t Tree) Intersection(args *EqArgs, u Tree) Tree {
	if t.count > u.count {
		t, u = u, t
		args = args.flip
	}
	count := 0
	return newTree(t.Root().Intersection(args, u.Root(), 0, &count), &count)
}

func (t Tree) Iterator() Iterator {
	return t.Root().Iterator(packedIteratorBuf(t.count))
}

func (t Tree) OrderedIterator(less Less, n int) Iterator {
	if n == -1 {
		n = t.count
	}
	o := &ordered{less: less, elements: make([]elementT, 0, n)}
	for i := t.Root().Iterator(packedIteratorBuf(t.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO(o)
	heap.Init(r)
	return r.(Iterator)
}

func (t *Tree) Remove(args *EqArgs, v elementT) {
	count := -t.count
	t.root = t.MutableRoot().Remove(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t Tree) SubsetOf(args *EqArgs, u Tree) bool {
	return t.Root().SubsetOf(args, u.Root(), 0)
}

func (t Tree) Transform(args *CombineArgs, f func(v elementT) elementT) Tree {
	count := 0
	return newTree(t.Root().Transform(args, 0, &count, f), &count)
}

func (t Tree) Reduce(args NodeArgs, r func(values ...elementT) elementT) elementT {
	return t.Root().Reduce(args, 0, r)
}

func (t Tree) Where(args *WhereArgs) Tree {
	count := 0
	return newTree(t.Root().Where(args, 0, &count), &count)
}

func (t Tree) With(args *CombineArgs, v elementT) Tree {
	count := -(t.count + 1)
	return newTreeNeg(t.Root().With(args, v, 0, newHasher(v, 0), &count), &count)
}

func (t Tree) Without(args *EqArgs, v elementT) Tree {
	count := -t.count
	return newTreeNeg(t.Root().Without(args, v, 0, newHasher(v, 0), &count), &count)
}

func packedIteratorBuf(count int) [][]*node {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]*node, 0, depth)
}
