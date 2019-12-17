package frozen

import (
	"container/heap"
	"fmt"
	"math/bits"
	"strings"
)

const (
	nodeBits = 3
	nodeSize = 1 << nodeBits
)

type node struct {
	mask     uintptr
	children [nodeSize]*node
	elem     interface{}
}

var empty *node = nil

func (n *node) isLeaf() bool {
	return n.mask == 0
}

func (n *node) equal(o *node, eq func(a, b interface{}) bool) bool {
	switch {
	case n == o:
		return true
	case n == nil || o == nil || n.mask != o.mask:
		return false
	case n.isLeaf():
		return eq(n.elem, o.elem)
	default:
		for mask := n.mask; mask != 0; mask &= mask - 1 {
			i := bits.TrailingZeros64(uint64(mask))
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
		return &node{elem: elem}
	case n.isLeaf():
		if Equal(n.elem, elem) {
			*c.middleIn++
			if composed := c.compose(n.elem, elem); composed != nil {
				*c.middleOut++
				if c.mutate {
					n.elem = composed
					return n
				}
				return &node{elem: composed}
			}
			return nil
		}
		if c.keep&rightSideOnly == 0 {
			if c.keep&leftSideOnly == 0 {
				return nil
			}
			return n
		}
		if c.keep&leftSideOnly == 0 {
			return &node{elem: elem}
		}
		nh := newHasher(n.elem, depth)
		result := &node{}
		last := result
		noffset, offset := nh.hash(), h.hash()
		for insane := 0; noffset == offset; insane++ {
			if insane > 100 {
				msg := fmt.Sprintf("%#v <=> %#v", n.elem.(fmt.Stringer).String(), elem.(fmt.Stringer).String())
				fmt.Println(msg)
			}
			last.mask = uintptr(1) << offset
			newLast := &node{}
			last.children[offset] = newLast
			last = newLast
			nh, h = nh.next(n.elem), h.next(elem)
			noffset, offset = nh.hash(), h.hash()
		}
		last.mask = uintptr(1)<<noffset | uintptr(1)<<offset
		last.children[noffset] = n
		last.children[offset] = &node{elem: elem}
		return result
	default:
		offset := h.hash()
		child := n.children[offset].applyImpl(elem, c, depth+1, h.next(elem))
		mask := uintptr(1) << offset
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
			if mask&(mask-1) == 0 {
				if onlyChild := n.children[bits.TrailingZeros64(uint64(mask))]; onlyChild.isLeaf() {
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

func (n *node) get(elem interface{}) interface{} {
	return n.getImpl(elem, newHasher(elem, 0))
}

func (n *node) getImpl(elem interface{}, h hasher) interface{} {
	switch {
	case n == nil:
		return nil
	case n.isLeaf():
		if Equal(elem, n.elem) {
			return n.elem
		}
		return nil
	default:
		return n.children[h.hash()].getImpl(elem, h.next(elem))
	}
}

func (n *node) merge(o *node, c *composer) *node {
	return n.mergeImpl(o, c, 0)
}

func (n *node) mergeImpl(o *node, c *composer, depth int) *node {
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
	case o.isLeaf():
		return n.applyImpl(o.elem, c, depth, newHasher(o.elem, depth))
	case n.isLeaf():
		return o.applyImpl(n.elem, c.flip(), depth, newHasher(n.elem, depth))
	default:
		var result node
		if c.keep&leftSideOnly != 0 {
			for mask := n.mask &^ o.mask; mask != 0; mask &= mask - 1 {
				i := bits.TrailingZeros64(uint64(mask))
				result.children[i] = n.children[i]
			}
			result.mask |= n.mask &^ o.mask
		}
		if c.keep&rightSideOnly != 0 {
			for mask := o.mask &^ n.mask; mask != 0; mask &= mask - 1 {
				i := bits.TrailingZeros64(uint64(mask))
				result.children[i] = o.children[i]
			}
			result.mask |= o.mask &^ n.mask
		}
		for mask := o.mask & n.mask; mask != 0; mask &= mask - 1 {
			i := bits.TrailingZeros64(uint64(mask))
			if child := n.children[i].mergeImpl(o.children[i], c, depth+1); child != nil {
				result.children[i] = child
				result.mask |= uintptr(1) << i
			}
		}
		if result.mask == 0 {
			return nil
		}
		if result.mask&(result.mask-1) == 0 {
			i := bits.TrailingZeros64(uint64(result.mask))
			if child := result.children[i]; child.isLeaf() {
				return child
			}
		}
		return &result
	}
}

func (n *node) String() string {
	if n == nil {
		return "∅"
	}
	if n.isLeaf() {
		return fmt.Sprintf("%v", n.elem)
	}
	var b strings.Builder
	b.WriteString("[")
	for i, v := range n.children {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(v.String())
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
				i.elem = b.elem
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
