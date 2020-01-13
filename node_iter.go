package frozen

import "container/heap"

// Less dictates the order of two elements.
type Less func(a, b interface{}) bool

type nodeIter struct {
	stk [][]*node
	li  Iterator
}

func newNodeIter(base []*node) *nodeIter {
	return &nodeIter{
		stk: [][]*node{base},
		li:  exhaustedIterator{},
	}
}

func (i *nodeIter) Next() bool {
	if i.li.Next() {
		return true
	}
	for {
		if nodes := &i.stk[len(i.stk)-1]; len(*nodes) > 0 {
			b := (*nodes)[0]
			*nodes = (*nodes)[1:]
			switch {
			case b == nil:
			case b.isLeaf():
				i.li = b.leaf().iterator()
				return i.li.Next()
			default:
				i.stk = append(i.stk, b.children[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}

func (i *nodeIter) Value() interface{} {
	return i.li.Value()
}

func (n *node) orderedIterator(less Less, capacity int) *ordered {
	o := &ordered{less: less, elems: make([]interface{}, 0, capacity)}
	for i := n.iterator(); i.Next(); {
		heap.Push(o, i.Value())
	}
	return o
}

type ordered struct {
	less  Less
	elems []interface{}
	val   interface{}
}

func (o *ordered) Next() bool {
	if len(o.elems) == 0 {
		return false
	}
	o.val = heap.Pop(o)
	return true
}

func (o *ordered) Value() interface{} {
	return o.val
}

func (o *ordered) Len() int {
	return len(o.elems)
}

func (o *ordered) Less(i, j int) bool {
	return o.less(o.elems[i], o.elems[j])
}

func (o *ordered) Swap(i, j int) {
	o.elems[i], o.elems[j] = o.elems[j], o.elems[i]
}

func (o *ordered) Push(x interface{}) {
	o.elems = append(o.elems, x)
}

func (o *ordered) Pop() interface{} {
	result := o.elems[len(o.elems)-1]
	o.elems = o.elems[:len(o.elems)-1]
	return result
}
