package frozen

import (
	"container/heap"
	"math/bits"
)

type nodeRoot struct {
	n     node
	count int
}

func newNodex(n node, count *int) nodeRoot {
	return nodeRoot{n: n, count: *count}
}

func newNodexNeg(n node, count *int) nodeRoot {
	return nodeRoot{n: n, count: -*count}
}

func (x nodeRoot) empty() bool {
	return x.count == 0
}

func (x nodeRoot) x() node {
	if x.empty() {
		return emptyNode{}
	}
	return x.n
}

func (x nodeRoot) gauge() parallelDepthGauge {
	return newParallelDepthGauge(x.count)
}

func (x nodeRoot) String() string {
	return x.x().String()
}

func (x nodeRoot) combine(args *combineArgs, y nodeRoot) nodeRoot {
	count := -(x.count + y.count)
	return newNodexNeg(x.x().combine(args, y.x(), 0, &count), &count)
}

func (x nodeRoot) difference(args *eqArgs, y nodeRoot) nodeRoot {
	count := -x.count
	a := x.x()
	b := y.x()
	return newNodexNeg(a.difference(args, b, 0, &count), &count)
}

func (x nodeRoot) equal(args *eqArgs, n nodeRoot) bool {
	return x.x().equal(args, n.x(), 0)
}

func (x nodeRoot) get(args *eqArgs, v interface{}) *interface{} {
	return x.x().get(args, v, newHasher(v, 0))
}

func (x nodeRoot) isSubsetOf(args *eqArgs, y nodeRoot) bool {
	return x.x().isSubsetOf(args, y.x(), 0)
}

func (x nodeRoot) intersection(args *eqArgs, y nodeRoot) nodeRoot {
	if x.count > y.count {
		x, y = y, x
		args = args.flip
	}
	count := 0
	return newNodex(x.x().intersection(args, y.x(), 0, &count), &count)
}

func (x nodeRoot) iterator() Iterator {
	return x.x().iterator(packedIteratorBuf(x.count))
}

func (x nodeRoot) orderedIterator(less Less, n int) Iterator {
	if n == -1 {
		n = x.count
	}
	o := &ordered{less: less, elements: make([]interface{}, 0, n)}
	for i := x.x().iterator(packedIteratorBuf(x.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO(o)
	heap.Init(r)
	return r.(Iterator)
}

func (x nodeRoot) transform(args *combineArgs, f func(v interface{}) interface{}) nodeRoot {
	count := 0
	return newNodex(x.x().transform(args, 0, &count, f), &count)
}

func (x nodeRoot) reduce(args nodeArgs, r func(values ...interface{}) interface{}) interface{} {
	return x.x().reduce(args, 0, r)
}

func (x nodeRoot) where(args *whereArgs) nodeRoot {
	count := 0
	return newNodex(x.x().where(args, 0, &count), &count)
}

func (x nodeRoot) with(args *combineArgs, v interface{}) nodeRoot {
	count := -(x.count + 1)
	return newNodexNeg(x.x().with(args, v, 0, newHasher(v, 0), &count), &count)
}

// func (x nodeRoot) without(args *eqArgs, v interface{}) nodeRoot {
// 	count := -x.count
// 	return newNodexNeg(x.x().without(args, v, 0, newHasher(v, 0), &count), &count)
// }

func packedIteratorBuf(count int) []packed {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([]packed, 0, depth)
}
