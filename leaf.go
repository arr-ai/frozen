package frozen

import (
	"fmt"
	"strings"
	"unsafe"
)

// Compile-time assert that branch and leaf have the same size and alignment.
const (
	leafElems = fanout / 2

	_ = -uint(unsafe.Sizeof(branch{}) ^ unsafe.Sizeof(leaf{}))
	_ = -uint(unsafe.Alignof(branch{}) ^ unsafe.Alignof(leaf{}))
)

var emptyLeaf = newLeaf()

type extraLeafElems []interface{}

type leaf struct { //nolint:maligned
	_         uint16 // mask only accessed via *branch
	lastIndex int16
	elems     [leafElems]interface{}
}

func newLeaf(elems ...interface{}) *leaf {
	l := &leaf{lastIndex: int16(len(elems) - 1)}
	copy(l.elems[:], elems)
	if len(elems) > leafElems {
		panic("too many elems")
	}
	return l
}

func (l *leaf) branch() *branch {
	return (*branch)(unsafe.Pointer(l))
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

func (l *leaf) set(i int, v interface{}) *leaf {
	*l.elem(i) = v
	return l
}

func (l *leaf) push(v interface{}, c *cloner) *leaf {
	l = c.leaf(l)
	l.lastIndex++
	switch {
	case l.lastIndex < leafElems:
		l.elems[l.lastIndex] = v
	case l.lastIndex == leafElems:
		*l.last() = extraLeafElems([]interface{}{*l.last(), v})
		l.lastIndex++
	default:
		*l.last() = append(c.extras(l, 1), v)
	}
	return l
}

func (l *leaf) pop(c *cloner) (result *leaf, v interface{}) {
	if l.lastIndex == 0 {
		return nil, l.elems[0]
	}

	l = c.leaf(l)
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

func (l *leaf) remove(i int, c *cloner) *leaf {
	if i == int(l.lastIndex) {
		result, _ := l.pop(c)
		return result
	}
	if result, v := l.pop(c); result != nil {
		if i > int(result.lastIndex) {
			i--
		}
		if i > leafElems {
			*result.last() = c.extras(l, 0)
		}
		result.set(i, v)
		return result
	}
	return nil
}

func (l *leaf) equal(m *leaf, eq func(a, b interface{}) bool) bool {
	return l.isSubsetOf(m, eq) && m.isSubsetOf(l, eq)
}

func (l *leaf) where(pred func(elem interface{}) bool, matches *int) *branch {
	result := leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		if pred(v) {
			*matches++
			result.push(v, theMutator)
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.branch()
}

func (l *leaf) foreach(f func(elem interface{})) {
	for i := l.iterator(); i.Next(); {
		f(*i.elem())
	}
}

func (l *leaf) intersection(n *branch, depth int, count *int) *branch {
	result := leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		h := newHasher(v, depth)
		if n.getImpl(v, h) != nil {
			*count++
			result.push(v, theMutator)
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.branch()
}

func (l *leaf) with(
	v interface{},
	f func(a, b interface{}) interface{},
	depth int,
	h hasher,
	matches *int,
	c *cloner,
) *branch {
	if elem, i := l.get(v, Equal); elem != nil {
		*matches++
		res := f(elem, v)
		return c.leaf(l).set(i, res).branch()
	}
	h0 := newHasher(l.elems[0], depth)
	if h == h0 {
		return l.push(v, c).branch()
	}
	result := &branch{}
	last := result
	noffset, offset := h0.hash(), h.hash()
	for noffset == offset {
		last.mask = 1 << uint(offset)
		newLast := &branch{}
		last.children[offset] = newLast
		last = newLast
		h0, h = h0.next(), h.next()
		noffset, offset = h0.hash(), h.hash()
	}
	last.mask = 1<<uint(noffset) | 1<<uint(offset)
	last.children[noffset] = l.branch()
	last.children[offset] = newLeaf(v).branch()
	return result
}

func (l *leaf) difference(n *branch, depth int, matches *int) *branch {
	result := leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		h := newHasher(v, depth)
		if n.getImpl(v, h) == nil {
			result.push(v, theMutator)
		} else {
			*matches++
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.branch()
}

func (l *leaf) without(v interface{}, matches *int, c *cloner) *branch {
	if elem, i := l.get(v, Equal); elem != nil {
		*matches++
		return l.remove(i, c).branch()
	}
	return l.branch()
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

func (l *leaf) iterator() leafIterator {
	return newLeafIterator(l)
}
