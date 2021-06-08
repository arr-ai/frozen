package frozen

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

const nodeCount = 1 << nodeBits

var useRHS = func(_, b interface{}) interface{} { return b }
var useLHS = func(a, _ interface{}) interface{} { return a }

type node struct {
	mask     MaskIterator
	_        uint16
	children [nodeCount]*node
}

func (n *node) isLeaf() bool {
	return n.mask == 0
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

func (n *node) setChild(i int, child *node) *node {
	mask := MaskIterator(1) << uint(i)
	if child != nil {
		n.mask |= mask
	} else {
		n.mask &^= mask
	}
	n.children[i] = child
	return n
}

func (n *node) setChildAsync(i int, child *node, m sync.Locker) {
	m.Lock()
	defer m.Unlock()
	n.setChild(i, child)
}

func (n *node) setChildren(mask MaskIterator, children *[nodeCount]*node) {
	n.mask |= mask
	for ; mask != 0; mask = mask.Next() {
		i := mask.Index()
		n.children[i] = children[i]
	}
}

func (n *node) calcMask() {
	mask := MaskIterator(0)
	for i, child := range n.children {
		if child != nil {
			mask |= MaskIterator(1 << i)
		}
	}
	n.mask = mask
}

func (n *node) opCanonical(
	o *node,
	result *node,
	depth int,
	count *int,
	c *cloner,
	op func(a, b *node, count *int) *node,
) *node {
	if depth <= c.parallelDepth {
		var wg sync.WaitGroup
		var counts [nodeCount]int
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			wg.Add(1)
			go func() {
				defer wg.Done()
				result.children[i] = op(n.children[i], o.children[i], &counts[i])
			}()
		}
		wg.Wait()
		for _, c := range counts {
			*count += c
		}
		result.calcMask()
	} else {
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			result.setChild(i, op(n.children[i], o.children[i], count))
		}
	}
	return result.canonical()
}

func (n *node) equal(o *node, eq func(a, b interface{}) bool, depth int, c *cloner) bool {
	switch {
	case n == o:
		return true
	case n == nil || o == nil || n.mask != o.mask:
		return false
	case n.isLeaf():
		return n.leaf().equal(o.leaf(), eq)
	default:
		if depth <= c.parallelDepth {
			results := make(chan bool, nodeCount)
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				go func() {
					results <- n.children[i].equal(o.children[i], eq, depth+1, c)
				}()
			}
			for mask := n.mask; mask != 0; mask = mask.Next() {
				if !<-results {
					return false
				}
			}
			return true
		}
		for mask := n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if !n.children[i].equal(o.children[i], eq, depth+1, c) {
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
		i := h.hash()
		return n.children[i].getImpl(v, h.next())
	}
}

func (n *node) isSubsetOf(o *node, depth int, c *cloner) bool {
	switch {
	case n == nil:
		return true
	case o == nil:
		return false
	case n.isLeaf() && o.isLeaf():
		return n.leaf().isSubsetOf(o.leaf(), Equal)
	case n.isLeaf():
		for i := n.leaf().iterator(); i.Next(); {
			v := i.Value()
			if o.getImpl(v, newHasher(v, depth)) == nil {
				return false
			}
		}
		return true
	case o.isLeaf():
		return false
	default:
		if depth <= c.parallelDepth {
			results := make(chan bool, nodeCount)
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				go func() {
					results <- n.children[i].isSubsetOf(o.children[i], depth+1, c)
				}()
			}
			for mask := n.mask; mask != 0; mask = mask.Next() {
				if !<-results {
					return false
				}
			}
			return true
		}
		for mask := n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if !n.children[i].isSubsetOf(o.children[i], depth+1, c) {
				return false
			}
		}
		return true
	}
}

func (n *node) where(pred func(elem interface{}) bool, depth int, matches *int, c *cloner) *node {
	switch {
	case n == nil:
		return n
	case n.isLeaf():
		return n.leaf().where(pred, matches)
	default:
		var prepared *node
		return n.opCanonical(n, c.node(n, &prepared), depth, matches, c, func(a, _ *node, matches *int) *node {
			return a.where(pred, depth+1, matches, c)
		})
	}
}

type foreacher struct {
	f     func(elem interface{})
	spawn func() *foreacher
}

func (n *node) foreach(f *foreacher, depth int, c *cloner) {
	switch {
	case n == nil:
	case n.isLeaf():
		n.leaf().foreach(f.f)
	default:
		if depth <= c.parallelDepth {
			// var wg sync.WaitGroup
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				g := f.spawn()
				// wg.Add(1)
				// go func() {
				// 	defer wg.Done()
				n.children[i].foreach(g, depth+1, c)
				// }()
			}
			// wg.Wait()
		} else {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				n.children[i].foreach(f, depth+1, c)
			}
		}
	}
}

type forbatcher struct {
	f     func(elem ...interface{})
	spawn func() *forbatcher
}

const forbatchSize = 1 << 16

type forbatch struct {
	singular bool
	f        func(elem ...interface{})
	buf      []interface{}
}

func newForbatch(singular bool, f func(elem ...interface{})) *forbatch {
	return &forbatch{
		singular: singular,
		f:        f,
		buf:      make([]interface{}, 0, forbatchSize),
	}
}

func (b *forbatch) add(elem interface{}) {
	if cap(b.buf) == 0 && !b.singular {
		b.flush()
	}
	b.buf = append(b.buf, elem)
}

func (b *forbatch) flush() {
	if len(b.buf) > 0 {
		b.f(b.buf...)
		b.buf = b.buf[:0]
	}
}

func (n *node) forbatches(f *forbatcher, depth int, c *cloner) {
	b := newForbatch(false, f.f)
	defer b.flush()
	n.forbatchesImpl(f, depth, c, b)
}

func (n *node) forbatchesImpl(f *forbatcher, depth int, c *cloner, fb *forbatch) {
	switch {
	case n == nil:
	case n.isLeaf():
		for i := n.leaf().iterator(); i.Next(); {
			fb.add(i.Value())
		}
	default:
		if depth <= c.parallelDepth {
			var wg sync.WaitGroup
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				g := f.spawn()
				wg.Add(1)
				go func() {
					defer wg.Done()
					b := newForbatch(true, g.f)
					defer b.flush()
					// TODO: 1<<15 is based on heuristics in newCloner. Confirm.
					for i := n.children[i].iterator(1 << 15); i.Next(); {
						b.add(i.Value())
					}
				}()
			}
			wg.Wait()
		} else {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				n.children[i].forbatchesImpl(f, depth+1, c, fb)
			}
		}
	}
}

func (n *node) intersection(o *node, depth int, count *int, c *cloner) *node {
	switch {
	case n == nil || o == nil:
		return nil
	case n.isLeaf():
		return n.leaf().intersection(o, depth, count)
	case o.isLeaf():
		return o.leaf().intersection(n, depth, count)
	default:
		return n.opCanonical(o, &node{}, depth, count, c, func(a, b *node, count *int) *node {
			return a.intersection(b, depth+1, count, c)
		})
	}
}

func (n *node) union(o *node, f func(a, b interface{}) interface{}, depth int, matches *int, c *cloner) *node {
	var prepared *node
	transform := f
	switch {
	case n == nil:
		return o
	case o == nil:
		return n
	case n.isLeaf():
		n, o, transform = o, n, func(a, b interface{}) interface{} { return f(b, a) }
		fallthrough
	case o.isLeaf():
		for i := o.leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.with(v, transform, depth, newHasher(v, depth), matches, c, &prepared)
		}
		return n
	default:
		result := c.node(n, &prepared)
		result.setChildren(o.mask&^n.mask, &o.children)
		if depth <= c.parallelDepth {
			var m sync.Mutex
			var wg sync.WaitGroup
			var mm [nodeCount]int
			for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				wg.Add(1)
				go func() {
					defer wg.Done()
					result.setChildAsync(i, n.children[i].union(o.children[i], transform, depth+1, &mm[i], c), &m)
				}()
			}
			wg.Wait()
			for _, m := range mm {
				*matches += m
			}
		} else {
			for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				result.setChild(i, n.children[i].union(o.children[i], transform, depth+1, matches, c))
			}
		}
		return result
	}
}

func (n *node) with(
	v interface{},
	f func(a, b interface{}) interface{},
	depth int,
	h hasher,
	matches *int,
	c *cloner,
	prepared **node,
) *node {
	switch {
	case n == nil:
		return newLeaf(v).node()
	case n.isLeaf():
		return n.leaf().with(v, f, depth, h, matches, c)
	default:
		offset := h.hash()
		var childPrepared *node
		child := n.children[offset].with(v, f, depth+1, h.next(), matches, c, &childPrepared)
		if child.isLeaf() && (n.mask|MaskIterator(1)<<uint(offset)).Count() == 1 {
			return child
		}
		return c.node(n, prepared).setChild(offset, child)
	}
}

func (n *node) difference(o *node, depth int, matches *int, c *cloner) *node {
	var prepared *node
	switch {
	case n == nil || o == nil:
		return n
	case o.isLeaf():
		for i := o.leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.without(v, depth, newHasher(v, depth), matches, c, &prepared)
		}
		return n
	case n.isLeaf():
		return n.leaf().difference(o, depth, matches)
	default:
		result := theCopier.node(n, &prepared)
		return n.opCanonical(o, result, depth, matches, c, func(a, b *node, matches *int) *node {
			return a.difference(b, depth+1, matches, c)
		})
	}
}

func (n *node) without(v interface{}, depth int, h hasher, matches *int, c *cloner, prepared **node) *node {
	switch {
	case n == nil:
		return n
	case n.isLeaf():
		return n.leaf().without(v, matches, c)
	default:
		offset := h.hash()
		var childPrepared *node
		child := n.children[offset].without(v, depth+1, h.next(), matches, c, &childPrepared)
		return c.node(n, prepared).setChild(offset, child).canonical()
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
			fmt.Fprintf(&sb, "%v\n", indentBlock(v.String()))
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

func (n *node) iterator(count int) Iterator {
	if n == nil {
		return exhaustedIterator{}
	}
	if n.isLeaf() {
		return newNodeIter([]*node{n}, count)
	}
	return newNodeIter(n.children[:], count)
}

func (n *node) elements(count int) []interface{} {
	elems := []interface{}{}
	for i := n.iterator(count); i.Next(); {
		elems = append(elems, i.Value())
	}
	return elems
}
