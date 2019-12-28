package frozen

import (
	"container/heap"
	"fmt"
	"strings"
	"unsafe"
)

const nodeCount = 1 << nodeBits

type node struct {
	mask     BitIterator
	children [nodeCount]*node
}

var empty *node = nil

func (n *node) isLeaf() bool {
	return n.mask&(1<<nodeCount-1) == 0
}

func (n *node) leaf() *leaf {
	return (*leaf)(unsafe.Pointer(n))
}

func (n *node) canonical() *node {
	if n.mask == 0 {
		return nil
	}
	if n.mask.Count() == 1 {
		if child := n.children[n.mask.Index()]; child.isLeaf() {
			return child
		}
	}
	return n
}

func (n *node) equal(o *node, eq func(a, b interface{}) bool) bool {
	switch {
	case n == o:
		return true
	case n == nil || o == nil || n.mask != o.mask:
		return false
	case n.isLeaf():
		return n.leaf().equal(o.leaf(), eq)
	default:
		for mask := n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if !n.children[i].equal(o.children[i], eq) {
				return false
			}
		}
		return true
	}
}

func (n *node) apply(c *composer, elem interface{}) *node {
	return n.applyImpl(elem, c, 0, newHasher(elem, 0))
}

//nolint:funlen,gocognit
func (n *node) applyImpl(elem interface{}, c *composer, depth int, h hasher) *node {
	switch {
	case n == nil:
		if c.keep&rightSideOnly == 0 {
			return n
		}
		return newLeaf(elem).node()
	case n.isLeaf():
		return n.leaf().applyImpl(elem, c, depth, h)
	default:
		offset := h.hash()
		child := n.children[offset].applyImpl(elem, c, depth+1, h.next())
		mask := BitIterator(1) << offset
		if (n.mask == mask || c.keep&leftSideOnly == 0) && (child == nil || child.isLeaf()) {
			return child
		}
		var result *node
		if c.keep&leftSideOnly == 0 {
			result = &node{}
		} else {
			if child != nil {
				mask = n.mask | mask
			} else {
				mask = n.mask &^ mask
			}
			if mask.Count() == 1 {
				if onlyChild := n.children[mask.Index()]; onlyChild.isLeaf() {
					return onlyChild
				}
			}
			if c.mutate {
				result = n
			} else {
				result = &node{}
				*result = *n
			}
		}
		result.mask = mask
		result.children[offset] = child
		return result
	}
}

func (n *node) isolate(v interface{}, delta *matchDelta, depth int, h hasher) (_ *node, count int) {
	switch {
	case n == nil:
		return nil, 0
	case n.isLeaf():
		return n.leaf().isolate(v, delta)
	default:
		offset := h.hash()
		return n.children[offset].isolate(v, delta, depth+1, h.next())
	}
}

func (n *node) get(elem interface{}) interface{} {
	return n.getImpl(elem, newHasher(elem, 0))
}

func (n *node) getImpl(v interface{}, h hasher) interface{} {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		if elem, _ := n.leaf().get(v, Equal); elem != nil {
			return elem
		}
		return nil
	default:
		return n.children[h.hash()].getImpl(v, h.next())
	}
}

func (n *node) merge(o *node, c *composer) *node {
	return n.mergeImpl(o, c, 0)
}

func (n *node) mergeImpl(o *node, c *composer, depth int) *node { //nolint:funlen
	switch {
	case n == nil:
		if c.keep&rightSideOnly != 0 {
			return o
		}
		return nil
	case o == nil:
		if c.keep&leftSideOnly != 0 {
			return n
		}
		return nil
	case n.isLeaf():
		n, o, c = o, n, c.flip()
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.next(); {
			v := *i.elem()
			n = n.applyImpl(v, c, depth, newHasher(v, depth))
		}
		return n
	default:
		var result node
		if c.keep&leftSideOnly != 0 {
			for mask := n.mask &^ o.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				result.children[i] = n.children[i]
			}
			result.mask |= n.mask &^ o.mask
		}
		if c.keep&rightSideOnly != 0 {
			for mask := o.mask &^ n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				result.children[i] = o.children[i]
			}
			result.mask |= o.mask &^ n.mask
		}
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if child := n.children[i].mergeImpl(o.children[i], c, depth+1); child != nil {
				result.children[i] = child
				result.mask |= BitIterator(1) << i
			}
		}
		return result.canonical()
	}
}

func (n *node) intersection(o *node, delta *matchDelta, depth int) (_ *node, count int) { //nolint:funlen
	switch {
	case n == nil || o == nil:
		return nil, 0
	case n.isLeaf():
		n, o = o, n
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.next(); {
			v := *i.elem()
			if p, count := n.isolate(v, delta, depth, newHasher(v, depth)); p != nil {
				return p, count
			}
		}
		return nil, 0
	default:
		var result node
		total := 0
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if child, count := n.children[i].intersection(o.children[i], delta, depth+1); child != nil {
				result.children[i] = child
				result.mask |= BitIterator(1) << i
				total += count
			}
		}
		return result.canonical(), total
	}
}

func (n *node) String() string {
	if n == nil {
		return "∅"
	}
	if n.isLeaf() {
		return n.leaf().String()
	}
	var b strings.Builder
	b.WriteString("[")
	for i, v := range n.children {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprint(&b, v)
	}
	b.WriteString("]")
	return b.String()
}

func (n *node) iterator() *nodeIter {
	if n == nil {
		return newNodeIter(nil)
	}
	if n.isLeaf() {
		return newNodeIter([]*node{n})
	}
	return newNodeIter(n.children[:])
}

type nodeIter struct {
	stk  [][]*node
	elem interface{}
}

func newNodeIter(base []*node) *nodeIter {
	return &nodeIter{stk: [][]*node{base}}
}

func (i *nodeIter) next() bool {
	for {
		if nodesp := &i.stk[len(i.stk)-1]; len(*nodesp) > 0 {
			b := (*nodesp)[0]
			*nodesp = (*nodesp)[1:]
			switch {
			case b == nil:
			case b.isLeaf():
				i.elem = b.leaf().elems[0]
				return true
			default:
				i.stk = append(i.stk, b.children[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			i.elem = nil
			return false
		}
	}
}

func (n *node) orderedIterator(less func(a, b interface{}) bool, capacity int) *ordered {
	o := &ordered{less: less, elems: make([]interface{}, 0, capacity)}
	for i := n.iterator(); i.next(); {
		heap.Push(o, i.elem)
	}
	return o
}

type ordered struct {
	less  func(a, b interface{}) bool
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
