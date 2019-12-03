package frozen

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/marcelocantos/frozen/pkg/value"
)

const (
	nodeBits = 3
	nodeSize = 1 << nodeBits
	nodeMask = nodeSize - 1
)

type hasher uint64

func newHasher(key interface{}, depth int) hasher {
	// Use the high four bits as the seed.
	h := hasher(0b1111<<60 | hash(key))
	for i := 0; i < depth; i++ {
		h = h.next(key)
	}
	return h
}

func (h hasher) next(key interface{}) hasher {
	if h >>= nodeBits; h < 0b1_0000 {
		return (h-1)<<60 | hasher(hash([2]interface{}{int(h), key})>>4)
	}
	return h
}

func (h hasher) hash() int {
	return int(h & nodeMask)
}

type entry struct {
	_    uintptr
	elem interface{}
	_    [nodeSize - 2]*node
}

func (e *entry) node() *node {
	return (*node)(unsafe.Pointer(e))
}

type node struct {
	mask     uintptr
	children [nodeSize]*node
}

func (n *node) entry() *entry {
	return (*entry)(unsafe.Pointer(n))
}

func (n *node) put(elem interface{}) (result *node, old interface{}) {
	return n.putImpl(elem, 0, newHasher(elem, 0))
}

func (n *node) putImpl(elem interface{}, depth int, h hasher) (result *node, old interface{}) {
	switch {
	case n == nil:
		return (&entry{elem: elem}).node(), nil
	case n.mask == 0:
		e := n.entry()
		if value.Equal(elem, e.elem) {
			return (&entry{elem: elem}).node(), e.elem
		}
		offset := newHasher(e.elem, depth).hash()
		result = &node{mask: 1 << offset}
		result.children[offset] = n
		return result.putImpl(elem, depth, h)
	default:
		offset := h.hash()
		t, old := n.children[offset].putImpl(elem, depth+1, h.next(elem))
		return n.update(offset, t), old
	}
}

func (n *node) get(elem interface{}) interface{} {
	return n.getImpl(elem, newHasher(elem, 0))
}

func (n *node) getImpl(elem interface{}, h hasher) interface{} {
	switch {
	case n == nil:
		return nil
	case n.mask == 0:
		e := n.entry()
		if value.Equal(elem, e.elem) {
			return e.elem
		}
		return nil
	default:
		return n.children[h.hash()].getImpl(elem, h.next(elem))
	}
}

func (n *node) delete(elem interface{}) (result *node, old interface{}) {
	return n.deleteImpl(elem, newHasher(elem, 0))
}

func (n *node) deleteImpl(elem interface{}, h hasher) (result *node, old interface{}) {
	switch {
	case n == nil:
	case n.mask == 0:
		e := n.entry()
		if value.Equal(elem, e.elem) {
			return nil, e.elem
		}
	default:
		offset := h.hash()
		if child, old := n.children[offset].deleteImpl(elem, h.next(elem)); old != nil {
			return n.update(offset, child), old
		}
	}
	return n, nil
}

func (n *node) update(offset int, child *node) *node {
	mask := uintptr(1) << offset
	if n.mask&^mask == 0 {
		if child == nil {
			return nil
		}
		if child.mask == 0 {
			return child
		}
	}
	result := *n
	result.children[offset] = child
	if child != nil {
		result.mask |= mask
	} else {
		result.mask &= ^mask
	}
	return &result
}

func (n *node) String() string {
	if n == nil {
		return "∅"
	}
	if n.mask == 0 {
		e := n.entry()
		return fmt.Sprintf("%v", e.elem)
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
	if n.mask == 0 {
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
			case b.mask == 0:
				e := b.entry()
				i.elem = e.elem
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
