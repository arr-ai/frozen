package frozen

import (
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

func (n *node) prepareForUpdate(mutate bool, prepared **node) *node {
	if *prepared == nil {
		if mutate {
			*prepared = n
		}
		result := *n
		*prepared = &result
	}
	return *prepared
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

func (n *node) clearChildren(mask BitIterator) {
	n.mask &^= mask
	for ; mask != 0; mask = mask.Next() {
		i := mask.Index()
		n.children[i] = nil
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
		for i := o.leaf().iterator(); i.Next(); {
			v := *i.elem()
			if p := n.only(v, depth, newHasher(v, depth), count); p != nil {
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

func (n *node) only(v interface{}, depth int, h hasher, count *int) *node {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		return n.leaf().only(v, count)
	default:
		return n.children[h.hash()].only(v, depth+1, h.next(), count)
	}
}

func (n *node) union(o *node, mutate, useRHS bool, depth int, matches *int) *node {
	var prepared *node
	switch {
	case n == nil:
		return o
	case o == nil:
		return n
	case n.isLeaf():
		n, o, useRHS = o, n, !useRHS
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.with(v, mutate, useRHS, depth, newHasher(v, depth), matches, &prepared)
		}
		return n
	default:
		result := n.prepareForUpdate(mutate, &prepared)
		result.setChildren(o.mask&^n.mask, &o.children)
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			result.setChild(i, n.children[i].union(o.children[i], mutate, useRHS, depth+1, matches))
		}
		return result
	}
}

func (n *node) with(v interface{}, mutate, useRHS bool, depth int, h hasher, matches *int, prepared **node) *node {
	switch {
	case n == nil:
		return newLeaf(v).node()
	case n.isLeaf():
		return n.leaf().with(v, mutate, useRHS, depth, h, matches)
	default:
		offset := h.hash()
		var childPrepared *node
		child := n.children[offset].with(v, mutate, useRHS, depth+1, h.next(), matches, &childPrepared)
		if (n.mask|BitIterator(1)<<offset).Count() == 1 && child.isLeaf() {
			return child
		}
		return n.prepareForUpdate(mutate, prepared).setChild(offset, child)
	}
}

func (n *node) difference(o *node, mutate bool, depth int, matches *int) *node {
	var prepared *node
	switch {
	case n == nil || o == nil:
		return n
	case o.isLeaf():
		for i := o.leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.without(v, mutate, depth, newHasher(v, depth), matches, &prepared)
		}
		return n
	case n.isLeaf():
		return n.leaf().difference(o, depth, matches)
	default:
		result := n.prepareForUpdate(false, &prepared)
		result.clearChildren(o.mask &^ n.mask)
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			result.setChild(i, n.children[i].difference(o.children[i], mutate, depth+1, matches))
		}
		return result.canonical()
	}
}

func (n *node) without(v interface{}, mutate bool, depth int, h hasher, matches *int, prepared **node) *node {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		return n.leaf().without(v, mutate, matches)
	default:
		offset := h.hash()
		var childPrepared *node
		child := n.children[offset].without(v, mutate, depth+1, h.next(), matches, &childPrepared)
		return n.prepareForUpdate(mutate, prepared).setChild(offset, child).canonical()
	}
}

func (n *node) Format(f fmt.State, _ rune) {
	s := n.String()
	fmt.Fprint(f, s)
	padFormat(f, len(s))
}

func (n *node) String() string {
	if n == nil {
		return "∅"
	}
	if n.isLeaf() {
		return n.leaf().String()
	}
	var sb strings.Builder
	deep := false
	for mask := n.mask; mask != 0; mask = mask.Next() {
		if !n.children[mask.Index()].isLeaf() {
			deep = true
			break
		}
	}
	fmt.Fprintf(&sb, "⁅%v ", n.mask)
	if deep {
		sb.WriteString("\n")
	}
	for mask := n.mask; mask != 0; mask = mask.Next() {
		v := n.children[mask.Index()]
		if deep {
			sb.WriteString(indentBlock(v.String()))
			sb.WriteString("\n")
		} else {
			if mask != n.mask {
				sb.WriteString(" ")
			}
			fmt.Fprint(&sb, v)
		}
	}
	sb.WriteString("⁆")
	return sb.String()
}

func (n *node) iterator() Iterator {
	if n == nil {
		return exhaustedIterator{}
	}
	if n.isLeaf() {
		return newNodeIter([]*node{n})
	}
	return newNodeIter(n.children[:])
}

func (n *node) elements() []interface{} {
	elems := []interface{}{}
	for i := n.iterator(); i.Next(); {
		elems = append(elems, i.Value())
	}
	return elems
}
