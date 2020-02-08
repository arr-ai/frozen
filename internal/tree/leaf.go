package tree

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/arr-ai/frozen/types"
)

const (
	leafElems = NodeCount / 2
)

// Compile-time assert that node and leaf have the same size and alignment.
const _ = -uint(unsafe.Sizeof(Node{}) ^ unsafe.Sizeof(Leaf{}))
const _ = -uint(unsafe.Alignof(Node{}) ^ unsafe.Alignof(Leaf{}))

type extraLeafElems []interface{}

type Leaf struct { //nolint:maligned
	_         uint16 // mask only accessed via *node
	lastIndex int16
	elems     [leafElems]interface{}
}

func newLeaf(elems ...interface{}) *Leaf {
	l := &Leaf{lastIndex: int16(len(elems) - 1)}
	copy(l.elems[:], elems)
	if len(elems) > leafElems {
		panic("too many elems")
	}
	return l
}

func (l *Leaf) node() *Node {
	return (*Node)(unsafe.Pointer(l))
}

func (l *Leaf) count() int {
	if l.lastIndex > leafElems {
		return int(l.lastIndex)
	}
	return int(l.lastIndex) + 1
}

func (l *Leaf) last() *interface{} { //nolint:gocritic
	return &l.elems[leafElems-1]
}

func (l *Leaf) extras() extraLeafElems {
	extras, _ := (*l.last()).(extraLeafElems)
	return extras
}

func (l *Leaf) elem(i int) *interface{} { //nolint:gocritic
	if i < leafElems {
		return &l.elems[i]
	}
	return &l.extras()[i-leafElems]
}

func (l *Leaf) set(i int, v interface{}) *Leaf {
	*l.elem(i) = v
	return l
}

func (l *Leaf) push(v interface{}, c *Cloner) *Leaf {
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

func (l *Leaf) pop(c *Cloner) (result *Leaf, v interface{}) {
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

func (l *Leaf) remove(i int, c *Cloner) *Leaf {
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

func (l *Leaf) equal(m *Leaf, eq func(a, b interface{}) bool) bool {
	return l.isSubsetOf(m, eq) && m.isSubsetOf(l, eq)
}

func (l *Leaf) where(pred func(elem interface{}) bool, matches *int) *Node {
	result := Leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		if pred(v) {
			*matches++
			result.push(v, Mutator)
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.node()
}

func (l *Leaf) foreach(f func(elem interface{})) {
	for i := l.iterator(); i.Next(); {
		f(*i.elem())
	}
}

func (l *Leaf) intersection(n *Node, depth int, count *int) *Node {
	result := Leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		h := NewHasher(v, depth)
		if n.getImpl(v, h) != nil {
			*count++
			result.push(v, Mutator)
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.node()
}

func (l *Leaf) with(
	v interface{},
	f func(a, b interface{}) interface{},
	depth int,
	h Hasher,
	matches *int,
	c *Cloner,
) *Node {
	if elem, i := l.get(v, types.Equal); elem != nil {
		*matches++
		res := f(elem, v)
		return c.leaf(l).set(i, res).node()
	}
	h0 := NewHasher(l.elems[0], depth)
	if h == h0 {
		return l.push(v, c).node()
	}
	result := &Node{}
	last := result
	noffset, offset := h0.Hash(), h.Hash()
	for noffset == offset {
		last.mask = 1 << offset
		newLast := &Node{}
		last.children[offset] = newLast
		last = newLast
		h0, h = h0.Next(), h.Next()
		noffset, offset = h0.Hash(), h.Hash()
	}
	last.mask = 1<<noffset | 1<<offset
	last.children[noffset] = l.node()
	last.children[offset] = newLeaf(v).node()
	return result
}

func (l *Leaf) difference(n *Node, depth int, matches *int) *Node {
	result := Leaf{lastIndex: -1}
	for i := l.iterator(); i.Next(); {
		v := *i.elem()
		h := NewHasher(v, depth)
		if n.getImpl(v, h) == nil {
			result.push(v, Mutator)
		} else {
			*matches++
		}
	}
	if result.lastIndex < 0 {
		return nil
	}
	return result.node()
}

func (l *Leaf) without(v interface{}, matches *int, c *Cloner) *Node {
	if elem, i := l.get(v, types.Equal); elem != nil {
		*matches++
		return l.remove(i, c).node()
	}
	return l.node()
}

func (l *Leaf) isSubsetOf(m *Leaf, eq func(a, b interface{}) bool) bool {
	for i := l.iterator(); i.Next(); {
		if elem, _ := m.get(*i.elem(), eq); elem == nil {
			return false
		}
	}
	return true
}

func (l *Leaf) get(v interface{}, eq func(a, b interface{}) bool) (_ interface{}, index int) { //nolint:gocritic
	for i := l.iterator(); i.Next(); {
		if elem := i.elem(); eq(*elem, v) {
			return *elem, i.index
		}
	}
	return nil, -1
}

func (l *Leaf) String() string {
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

func (l *Leaf) iterator() *leafIterator {
	return newLeafIterator(l)
}
