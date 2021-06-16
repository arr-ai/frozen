package frozen

import (
	"container/heap"
	"math/bits"
)

type tree struct {
	n     node
	count int
}

func newTree(n node, count *int) tree {
	return tree{n: n, count: *count}
}

func newTreeNeg(n node, count *int) tree {
	return tree{n: n, count: -*count}
}

func (x tree) empty() bool {
	return x.count == 0
}

func (x tree) x() node {
	if x.empty() {
		return emptyNode{}
	}
	return x.n
}

func (x tree) gauge() parallelDepthGauge {
	return newParallelDepthGauge(x.count)
}

func (x tree) String() string {
	return x.x().String()
}

func (x tree) combine(args *combineArgs, y tree) tree {
	count := -(x.count + y.count)
	return newTreeNeg(x.x().Combine(args, y.x(), 0, &count), &count)
}

func (x tree) difference(args *eqArgs, y tree) tree {
	count := -x.count
	a := x.x()
	b := y.x()
	return newTreeNeg(a.Difference(args, b, 0, &count), &count)
}

func (x tree) equal(args *eqArgs, n tree) bool {
	return x.x().Equal(args, n.x(), 0)
}

func (x tree) get(args *eqArgs, v interface{}) *interface{} {
	return x.x().Get(args, v, newHasher(v, 0))
}

func (x tree) isSubsetOf(args *eqArgs, y tree) bool {
	return x.x().SubsetOf(args, y.x(), 0)
}

func (x tree) intersection(args *eqArgs, y tree) tree {
	if x.count > y.count {
		x, y = y, x
		args = args.flip
	}
	count := 0
	return newTree(x.x().Intersection(args, y.x(), 0, &count), &count)
}

func (x tree) iterator() Iterator {
	return x.x().Iterator(packedIteratorBuf(x.count))
}

func (x tree) orderedIterator(less Less, n int) Iterator {
	if n == -1 {
		n = x.count
	}
	o := &ordered{less: less, elements: make([]interface{}, 0, n)}
	for i := x.x().Iterator(packedIteratorBuf(x.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO(o)
	heap.Init(r)
	return r.(Iterator)
}

func (x tree) transform(args *combineArgs, f func(v interface{}) interface{}) tree {
	count := 0
	return newTree(x.x().Transform(args, 0, &count, f), &count)
}

func (x tree) reduce(args nodeArgs, r func(values ...interface{}) interface{}) interface{} {
	return x.x().Reduce(args, 0, r)
}

func (x tree) where(args *whereArgs) tree {
	count := 0
	return newTree(x.x().Where(args, 0, &count), &count)
}

func (x tree) with(args *combineArgs, v interface{}) tree {
	count := -(x.count + 1)
	return newTreeNeg(x.x().With(args, v, 0, newHasher(v, 0), &count), &count)
}

// func (x nodeRoot) without(args *eqArgs, v interface{}) nodeRoot {
// 	count := -x.count
// 	return newTreeNeg(x.x().without(args, v, 0, newHasher(v, 0), &count), &count)
// }

func packedIteratorBuf(count int) []packed {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([]packed, 0, depth)
}
