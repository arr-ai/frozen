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

func (n *node) prepareForUpdate(mutate bool) *node {
	if mutate {
		return n
	}
	result := *n
	return &result
}

func (n *node) setChild(i int, child *node) *node {
	mask := BitIterator(1) << i
	if child != nil {
		n.mask |= mask
	} else {
		n.mask &^= mask
	}
	n.children[i] = child
	return n
}

func (n *node) setChildren(mask BitIterator, children *[nodeCount]*node) {
	n.mask |= mask
	for ; mask != 0; mask = mask.Next() {
		i := mask.Index()
		n.children[i] = children[i]
	}
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

func (n *node) intersection(o *node, depth int, count *int) *node { //nolint:funlen
	switch {
	case n == nil || o == nil:
		return nil
	case n.isLeaf():
		n, o = o, n
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.next(); {
			v := *i.elem()
			if p := n.valueIntersection(v, depth, newHasher(v, depth), count); p != nil {
				return p
			}
		}
		return nil
	default:
		var result node
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if child := n.children[i].intersection(o.children[i], depth+1, count); child != nil {
				result.setChild(i, child)
			}
		}
		return result.canonical()
	}
}

func (n *node) valueIntersection(v interface{}, depth int, h hasher, count *int) *node {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		return n.leaf().valueIntersection(v, count)
	default:
		return n.children[h.hash()].valueIntersection(v, depth+1, h.next(), count)
	}
}

func (n *node) union(o *node, mutate, useRHS bool, depth int, matches *int) *node {
	switch {
	case n == nil:
		return o
	case o == nil:
		return n
	case n.isLeaf():
		n, o, useRHS = o, n, !useRHS
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.next(); {
			v := *i.elem()
			n = n.valueUnion(v, mutate, useRHS, depth, newHasher(v, depth), matches)
		}
		return n
	default:
		var result node
		result.setChildren(n.mask&^o.mask, &n.children)
		result.setChildren(o.mask&^n.mask, &o.children)
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if child := n.children[i].union(o.children[i], mutate, useRHS, depth+1, matches); child != nil {
				result.setChild(i, child)
			}
		}
		return &result
	}
}

func (n *node) valueUnion(v interface{}, mutate, useRHS bool, depth int, h hasher, matches *int) *node {
	switch {
	case n == nil:
		return newLeaf(v).node()
	case n.isLeaf():
		return n.leaf().valueUnion(v, mutate, useRHS, depth, h, matches)
	default:
		offset := h.hash()
		child := n.children[offset].valueUnion(v, mutate, useRHS, depth+1, h.next(), matches)
		if (n.mask|BitIterator(1)<<offset).Count() == 1 && child.isLeaf() {
			return child
		}
		return n.prepareForUpdate(mutate).setChild(offset, child)
	}
}

func (n *node) difference(o *node, mutate bool, depth int, matches *int) *node {
	switch {
	case n == nil || o == nil:
		return n
	case o.isLeaf():
		for i := o.leaf().iterator(); i.next(); {
			v := *i.elem()
			n = n.without(v, mutate, depth, newHasher(v, depth), matches)
		}
		return n
	case n.isLeaf():
		return n.leaf().difference(o, depth, matches)
	default:
		var result node
		result.setChildren(n.mask&^o.mask, &n.children)
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if child := n.children[i].difference(o.children[i], mutate, depth+1, matches); child != nil {
				result.children[i] = child
				result.mask |= BitIterator(1) << i
			}
		}
		return result.canonical()
	}
}

func (n *node) without(v interface{}, mutate bool, depth int, h hasher, matches *int) *node {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		return n.leaf().without(v, mutate, matches)
	default:
		offset := h.hash()
		child := n.children[offset].without(v, mutate, depth+1, h.next(), matches)
		mask := BitIterator(1) << offset
		if n.mask == mask && (child == nil || child.isLeaf()) {
			return child
		}
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
		return n.prepareForUpdate(mutate).setChild(offset, child).canonical()
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
