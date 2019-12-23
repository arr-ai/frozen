package frozen

import (
	"fmt"
	"strings"
	"unsafe"
)

const (
	leafElems   = int((unsafe.Sizeof(node{}) - 2*unsafe.Sizeof(uintptr(0))) / unsafe.Sizeof(interface{}(nil)))
	leafPadding = unsafe.Sizeof(node{}) - unsafe.Sizeof(leafCore{})
)

// Compile-time assert that node and leaf have the same size and alignment.
const _ = -uint(unsafe.Sizeof(node{}) ^ unsafe.Sizeof(leaf{}))
const _ = -uint(unsafe.Alignof(node{}) ^ unsafe.Alignof(leaf{}))

type extraLeafElems []interface{}

// Used to compute leaf padding [...]byte
type leafCore struct {
	_, _  uintptr // mask only accessed via *node
	elems [leafElems]interface{}
}

type leaf struct { //nolint:maligned
	leafCore
	_ [leafPadding]byte
}

func newLeaf(elem interface{}) *node {
	l := &leaf{}
	l.elems[0] = elem
	return l.node()
}

func (l *leaf) node() *node {
	return (*node)(unsafe.Pointer(l))
}

func (l *leaf) last() *interface{} { //nolint:critic
	return &l.elems[leafElems-1]
}

func (l *leaf) extras() (extraLeafElems, bool) {
	extras, ok := (*l.last()).(extraLeafElems)
	return extras, ok
}

func (l *leaf) append(v interface{}, mutate bool) *leaf {
	var result *leaf
	if mutate {
		result = l
	} else {
		result = &leaf{}
		*result = *l
	}
	for i, elem := range result.elems {
		if elem == nil {
			result.elems[i] = v
			return result
		}
	}
	if extras, ok := result.extras(); ok {
		if mutate {
			extras = append(make([]interface{}, 0, len(extras)+1), extras...)
		}
		*result.last() = append(extras, v)
	} else {
		*result.last() = extraLeafElems([]interface{}{*result.last(), v})
	}
	return result
}

func (l *leaf) equal(m *leaf, eq func(a, b interface{}) bool) bool {
	return l.isSubsetOf(m, eq) && m.isSubsetOf(l, eq)
}

func (l *leaf) applyImpl(v interface{}, c *composer, depth int, h hasher) *node {
	if elem := l.get(v, Equal); elem != nil {
		*c.middleIn++
		if composed := c.compose(*elem, v); composed != nil {
			*c.middleOut++
			if c.mutate {
				*elem = composed
				return l.node()
			}
			return newLeaf(composed)
		}
		return nil
	}
	if c.keep&rightSideOnly == 0 {
		if c.keep&leftSideOnly == 0 {
			return nil
		}
		return l.node()
	}
	if c.keep&leftSideOnly == 0 {
		return newLeaf(v)
	}
	if *l.last() == nil {
		// return l.append(v, c.mutate).node()
	}
	result := &node{}
	last := result
	nh := newHasher(l.elems[0], depth)
	noffset, offset := nh.hash(), h.hash()
	for noffset == offset {
		last.mask = uintptr(1) << offset
		newLast := &node{}
		last.children[offset] = newLast
		last = newLast
		nh, h = nh.next(l.elems[0]), h.next(v)
		noffset, offset = nh.hash(), h.hash()
	}
	last.mask = uintptr(1)<<noffset | uintptr(1)<<offset
	last.children[noffset] = l.node()
	last.children[offset] = newLeaf(v)
	return result
}

func (l *leaf) isSubsetOf(m *leaf, eq func(a, b interface{}) bool) bool {
	for i := l.iterator(); i.next(); {
		if m.get(*i.elem(), eq) == nil {
			return false
		}
	}
	return true
}

func (l *leaf) get(v interface{}, eq func(a, b interface{}) bool) *interface{} { //nolint:gocritic
	for i := l.iterator(); i.next(); {
		if elem := i.elem(); eq(*elem, v) {
			return elem
		}
	}
	return nil
}

func (l *leaf) String() string {
	var b strings.Builder
	b.WriteString("(")
	for i, j := l.iterator(), 0; i.next(); j++ {
		if j > 0 {
			b.WriteString(",")
		}
		fmt.Fprint(&b, *i.elem())
	}
	b.WriteString(")")
	return b.String()
}

func (l *leaf) iterator() leafIterator {
	return leafIterator{l: l, index: -1}
}

type leafIterator struct {
	l      *leaf
	index  int
	extras extraLeafElems
}

func (i *leafIterator) next() bool {
	i.index++
	switch {
	case i.index < leafElems-1:
		return i.l.elems[i.index] != nil
	case i.index == leafElems-1:
		e := i.l.elems[i.index]
		if e == nil {
			return false
		}
		if extras, ok := e.(extraLeafElems); ok {
			i.index++
			i.extras = extras
		}
		return true
	default:
		return i.index-leafElems < len(i.extras)
	}
}

func (i *leafIterator) elem() *interface{} { //nolint:gocritic
	if i.index < leafElems {
		return &i.l.elems[i.index]
	}
	return &i.extras[i.index-leafElems]
}
