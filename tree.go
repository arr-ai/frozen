package frozen

import (
	"container/heap"
	"math/bits"
)

type tree struct {
	root  node
	count int
}

func newTree(n node, count *int) tree {
	return tree{root: n, count: *count}
}

func newTreeNeg(n node, count *int) tree {
	return tree{root: n, count: -*count}
}

func (t tree) Root() node {
	if t.count == 0 {
		return emptyNode{}
	}
	return t.root
}

func (t tree) Gauge() parallelDepthGauge {
	return newParallelDepthGauge(t.count)
}

func (t tree) String() string {
	return t.Root().String()
}

func (t tree) Builder() *nodeBuilder {
	return &nodeBuilder{t: unTree{root: unDefroster{n: t.root}}}
}

func (t tree) Combine(args *combineArgs, u tree) tree {
	count := -(t.count + u.count)
	return newTreeNeg(t.Root().Combine(args, u.Root(), 0, &count), &count)
}

func (t tree) Difference(args *eqArgs, u tree) tree {
	count := -t.count
	a := t.Root()
	b := u.Root()
	return newTreeNeg(a.Difference(args, b, 0, &count), &count)
}

func (t tree) Equal(args *eqArgs, n tree) bool {
	return t.Root().Equal(args, n.Root(), 0)
}

func (t tree) Get(args *eqArgs, v interface{}) *interface{} {
	return t.Root().Get(args, v, newHasher(v, 0))
}

func (t tree) SubsetOf(args *eqArgs, u tree) bool {
	return t.Root().SubsetOf(args, u.Root(), 0)
}

func (t tree) Intersection(args *eqArgs, u tree) tree {
	if t.count > u.count {
		t, u = u, t
		args = args.flip
	}
	count := 0
	return newTree(t.Root().Intersection(args, u.Root(), 0, &count), &count)
}

func (t tree) Iterator() Iterator {
	return t.Root().Iterator(packedIteratorBuf(t.count))
}

func (t tree) OrderedIterator(less Less, n int) Iterator {
	if n == -1 {
		n = t.count
	}
	o := &ordered{less: less, elements: make([]interface{}, 0, n)}
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

func (t tree) Transform(args *combineArgs, f func(v interface{}) interface{}) tree {
	count := 0
	return newTree(t.Root().Transform(args, 0, &count, f), &count)
}

func (t tree) Reduce(args nodeArgs, r func(values ...interface{}) interface{}) interface{} {
	return t.Root().Reduce(args, 0, r)
}

func (t tree) Where(args *whereArgs) tree {
	count := 0
	return newTree(t.Root().Where(args, 0, &count), &count)
}

func (t tree) With(args *combineArgs, v interface{}) tree {
	count := -(t.count + 1)
	return newTreeNeg(t.Root().With(args, v, 0, newHasher(v, 0), &count), &count)
}

// func (t tree) without(args *eqArgs, v interface{}) nodeRoot {
// 	count := -t.count
// 	return newTreeNeg(t.x().without(args, v, 0, newHasher(v, 0), &count), &count)
// }

func packedIteratorBuf(count int) []packed {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([]packed, 0, depth)
}
