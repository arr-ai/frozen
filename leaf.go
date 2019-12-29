package frozen

import (
	"fmt"
	"strings"
	"unsafe"
)

const (
	leafElems = nodeCount / 2
)

// Compile-time assert that node and leaf have the same size and alignment.
const _ = -uint(unsafe.Sizeof(node{}) ^ unsafe.Sizeof(leaf{}))
const _ = -uint(unsafe.Alignof(node{}) ^ unsafe.Alignof(leaf{}))

type extraLeafElems []interface{}

type leaf struct { //nolint:maligned
	_         uint16 // mask only accessed via *node
	lastIndex int16
	elems     [leafElems]interface{}
}

func newLeaf(elem interface{}) *leaf {
	l := &leaf{lastIndex: 0}
	l.elems[0] = elem
	return l
}

func (l *leaf) node() *node {
	return (*node)(unsafe.Pointer(l))
}

func (l *leaf) last() *interface{} { //nolint:gocritic
	return &l.elems[leafElems-1]
}

func (l *leaf) extras() extraLeafElems {
	extras, _ := (*l.last()).(extraLeafElems)
	return extras
}

func (l *leaf) elem(i int) *interface{} { //nolint:gocritic
	if i < leafElems {
		return &l.elems[i]
	}
	return &l.extras()[i-leafElems]
}

func (l *leaf) prepareForUpdate(mutate bool) *leaf {
	if mutate {
		return l
	}
	result := *l
	return &result
}

// Assumptions: leaf already prepared and `l.extras() != nil`.
func (l *leaf) prepareExtrasForUpdate(mutate bool, capacityIncrease int) extraLeafElems {
	extras := l.extras()
	if !mutate {
		extras = append(make([]interface{}, 0, len(extras)+capacityIncrease), extras...)
	}
	return extras
}

func (l *leaf) set(i int, v interface{}) *leaf {
	*l.elem(i) = v
	return l
}

func (l *leaf) push(v interface{}, mutate bool) *leaf {
	l = l.prepareForUpdate(mutate)
	l.lastIndex++
	switch {
	case l.lastIndex < leafElems:
		l.elems[l.lastIndex] = v
	case l.lastIndex == leafElems:
		*l.last() = extraLeafElems([]interface{}{*l.last(), v})
		l.lastIndex++
	default:
		*l.last() = append(l.prepareExtrasForUpdate(mutate, 1), v)
	}
	return l
}

func (l *leaf) pop(mutate bool) (result *leaf, v interface{}) {
	if l.lastIndex == 0 {
		return nil, l.elems[0]
	}

	l = l.prepareForUpdate(mutate)
	switch {
	case l.lastIndex < leafElems:
		v = l.elems[l.lastIndex]
		l.elems[l.lastIndex] = nil
	default:
		extras := l.extras()
		v = extras[len(extras)-1]
		extras[len(extras)-1] = nil
		if l.lastIndex == leafElems+1 {
			*l.last() = extras[0]
			l.lastIndex--
		} else {
			*l.last() = extras[:len(extras)-1]
		}
	}
	l.lastIndex--
	return l, v
}

func (l *leaf) remove(i int, mutate bool) *leaf {
	if i == int(l.lastIndex) {
		result, _ := l.pop(mutate)
		return result
	}
	if result, v := l.pop(mutate); result != nil {
		if i > int(result.lastIndex) {
			i--
		}
		if i > leafElems && !mutate {
			*result.last() = append([]interface{}{}, l.extras()...)
		}
		result.set(i, v)
		return result
	}
	return nil
}

func (l *leaf) equal(m *leaf, eq func(a, b interface{}) bool) bool {
	return l.isSubsetOf(m, eq) && m.isSubsetOf(l, eq)
}

func (l *leaf) only(v interface{}, count *int) *node {
	if elem, _ := l.get(v, Equal); elem != nil {
		*count++
		return newLeaf(v).node()
	}
	return nil
}

func (l *leaf) with(v interface{}, mutate, useRHS bool, depth int, h hasher, matches *int) *node {
	if elem, i := l.get(v, Equal); elem != nil {
		*matches++
		if useRHS {
			return l.prepareForUpdate(mutate).set(i, v).node()
		}
		return l.node()
	}
	h0 := newHasher(l.elems[0], depth)
	if h == h0 {
		return l.push(v, mutate).node()
	}
	result := &node{}
	last := result
	noffset, offset := h0.hash(), h.hash()
	for noffset == offset {
		last.mask = BitIterator(1) << offset
		newLast := &node{}
		last.children[offset] = newLast
		last = newLast
		h0, h = h0.next(), h.next()
		noffset, offset = h0.hash(), h.hash()
	}
	last.mask = BitIterator(1)<<noffset | BitIterator(1)<<offset
	last.children[noffset] = l.node()
	last.children[offset] = newLeaf(v).node()
	return result
}

func (l *leaf) difference(n *node, depth int, matches *int) *node {
	result := leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		h := newHasher(v, depth)
		if n.getImpl(v, h) == nil {
			result.push(v, true)
		} else {
			*matches++
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.node()
}

func (l *leaf) without(v interface{}, mutate bool, matches *int) *node {
	if elem, i := l.get(v, Equal); elem != nil {
		*matches++
		return l.remove(i, mutate).node()
	}
	return l.node()
}

func (l *leaf) isSubsetOf(m *leaf, eq func(a, b interface{}) bool) bool {
	for i := l.iterator(); i.Next(); {
		if elem, _ := m.get(*i.elem(), eq); elem == nil {
			return false
		}
	}
	return true
}

func (l *leaf) get(v interface{}, eq func(a, b interface{}) bool) (_ interface{}, index int) { //nolint:gocritic
	for i := l.iterator(); i.Next(); {
		if elem := i.elem(); eq(*elem, v) {
			return *elem, i.index
		}
	}
	return nil, -1
}

func (l *leaf) String() string {
	var b strings.Builder
	b.WriteString("(")
	for i, j := l.iterator(), 0; i.Next(); j++ {
		if j > 0 {
			b.WriteString(",")
		}
		fmt.Fprint(&b, *i.elem())
	}
	b.WriteString(")")
	return b.String()
}

func (l *leaf) iterator() *leafIterator {
	return newLeafIterator(l)
}
