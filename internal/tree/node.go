package tree

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/arr-ai/frozen/internal/fmtutil"
	"github.com/arr-ai/frozen/slave/proto/slave"
	"github.com/arr-ai/frozen/types"
)

const NodeCount = 1 << nodeBits

type Node struct {
	mask     types.MaskIterator
	_        uint16
	children [NodeCount]*Node
}

func (n *Node) IsLeaf() bool {
	return n.mask == 0
}

func (n *Node) Leaf() *Leaf {
	return (*Leaf)(unsafe.Pointer(n))
}

func (n *Node) canonical() *Node {
	if n.mask == 0 {
		return nil
	}
	if n.mask.Count() == 1 {
		if child := n.children[n.mask.Index()]; child.IsLeaf() {
			return child
		}
	}
	return n
}

func (n *Node) setChild(i int, child *Node) *Node {
	mask := types.MaskIterator(1) << i
	if child != nil {
		n.mask |= mask
	} else {
		n.mask &^= mask
	}
	n.children[i] = child
	return n
}

func (n *Node) setChildAsync(i int, child *Node, m sync.Locker) {
	m.Lock()
	defer m.Unlock()
	n.setChild(i, child)
}

func (n *Node) SetChildren(mask types.MaskIterator, children *[NodeCount]*Node) {
	n.mask |= mask
	for ; mask != 0; mask = mask.Next() {
		i := mask.Index()
		n.children[i] = children[i]
	}
}

func (n *Node) GetChildren() (types.MaskIterator, [NodeCount]*Node) {
	return n.mask, n.children
}

func (n *Node) clearChildren(mask types.MaskIterator) {
	n.mask &^= mask
	for ; mask != 0; mask = mask.Next() {
		i := mask.Index()
		n.children[i] = nil
	}
}

func (n *Node) hash(depth int) Hasher {
	if i := n.Iterator(0); i.Next() {
		return NewHasher(i.Value(), 0) >> (hashBits - 3*depth)
	}
	return 0
}

func (n *Node) opCanonical(
	o *Node,
	depth int,
	count *int,
	c *Cloner,
	result **Node,
	op func(a, b *Node, count *int, result **Node),
) {
	if depth == c.parallelDepth {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			var m sync.Mutex
			var wg sync.WaitGroup
			for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				wg.Add(1)
				c.run(func() {
					defer wg.Done()
					count := 0
					var child *Node
					op(n.children[i], o.children[i], &count, &child)
					(*result).setChildAsync(i, child, &m)
					c.update(count)
				})
			}
			wg.Wait()
			*result = (*result).canonical()
		}()
		*result = promiseNode
	} else {
		promised := false
		for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			var child *Node
			op(n.children[i], o.children[i], count, &child)
			if child == promiseNode {
				promised = true
			} else {
				(*result).setChild(i, child)
			}
		}
		if !promised {
			*result = (*result).canonical()
		}
	}
}

func (n *Node) Equal(o *Node, eq func(a, b interface{}) bool, depth int, c *Cloner) bool {
	switch {
	case n == o:
		return true
	case n == nil || o == nil || n.mask != o.mask:
		return false
	case n.IsLeaf():
		return n.Leaf().equal(o.Leaf(), eq)
	default:
		if depth == c.parallelDepth {
			c.run(func() {
				for mask := n.mask; mask != 0; mask = mask.Next() {
					i := mask.Index()
					c.update(n.children[i].Equal(o.children[i], eq, depth, c))
				}
			})
		} else {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				if !n.children[i].Equal(o.children[i], eq, depth, c) {
					return false
				}
			}
		}
		return true
	}
}

func (n *Node) Get(elem interface{}) interface{} {
	return n.getImpl(elem, NewHasher(elem, 0))
}

func (n *Node) getImpl(v interface{}, h Hasher) interface{} {
	switch {
	case n == nil:
		return nil
	case n.IsLeaf():
		if elem, _ := n.Leaf().get(v, types.Equal); elem != nil {
			return elem
		}
		return nil
	default:
		i := h.Hash()
		return n.children[i].getImpl(v, h.Next())
	}
}

func (n *Node) IsSubsetOf(o *Node, depth int, c *Cloner) bool {
	switch {
	case n == nil:
		return true
	case o == nil:
		return false
	case n.IsLeaf() && o.IsLeaf():
		return n.Leaf().isSubsetOf(o.Leaf(), types.Equal)
	case n.IsLeaf():
		for i := n.Leaf().iterator(); i.Next(); {
			v := i.Value()
			if o.getImpl(v, NewHasher(v, depth)) == nil {
				return false
			}
		}
		return true
	case o.IsLeaf():
		return false
	default:
		if depth == c.parallelDepth {
			c.run(func() {
				for mask := n.mask; mask != 0; mask = mask.Next() {
					i := mask.Index()
					c.update(n.children[i].IsSubsetOf(o.children[i], depth+1, c))
				}
			})
			return true
		}
		for mask := n.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
			if !n.children[i].IsSubsetOf(o.children[i], depth+1, c) {
				return false
			}
		}
		return true
	}
}

func (n *Node) Where(pred func(elem interface{}) bool, depth int, matches *int, c *Cloner, result **Node) {
	var prepared *Node
	switch {
	case n == nil:
		*result = n
	case n.IsLeaf():
		*result = n.Leaf().where(pred, matches)
	default:
		*result = Copier.node(n, &prepared)
		n.opCanonical(n, depth, matches, c, result, func(a, _ *Node, matches *int, result **Node) {
			a.Where(pred, depth+1, matches, c, result)
		})
	}
}

type ForEacher struct {
	f     func(elem interface{})
	spawn func() *ForEacher
}

func NewForEacher(f func(elem interface{}), spawn func() *ForEacher) *ForEacher {
	return &ForEacher{f: f, spawn: spawn}
}

func (n *Node) ForEach(f *ForEacher, depth int, c *Cloner) {
	switch {
	case n == nil:
	case n.IsLeaf():
		n.Leaf().foreach(f.f)
	default:
		if depth == c.parallelDepth {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				g := f.spawn()
				c.run(func() {
					n.children[i].ForEach(g, depth+1, c)
				})
			}
		} else {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				n.children[i].ForEach(f, depth+1, c)
			}
		}
	}
}

type ForBatcher struct {
	f     func(elem ...interface{})
	spawn func() *ForBatcher
}

func NewForBatcher(f func(elem ...interface{}), spawn func() *ForBatcher) *ForBatcher {
	return &ForBatcher{f: f, spawn: spawn}
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

func (n *Node) ForBatches(f *ForBatcher, depth int, c *Cloner) {
	b := newForbatch(false, f.f)
	defer b.flush()
	n.forbatchesImpl(f, depth, c, b)
}

func (n *Node) forbatchesImpl(f *ForBatcher, depth int, c *Cloner, fb *forbatch) {
	switch {
	case n == nil:
	case n.IsLeaf():
		for i := n.Leaf().iterator(); i.Next(); {
			fb.add(i.Value())
		}
	default:
		if depth == c.parallelDepth {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				g := f.spawn()
				c.run(func() {
					b := newForbatch(true, g.f)
					defer b.flush()
					// TODO: 1<<15 is based on heuristics in newCloner. Confirm.
					for i := n.children[i].Iterator(1 << 15); i.Next(); {
						b.add(i.Value())
					}
				})
			}
		} else {
			for mask := n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				n.children[i].forbatchesImpl(f, depth+1, c, fb)
			}
		}
	}
}

func (n *Node) Intersection(o *Node, r *Resolver, depth int, count *int, c *Cloner, result **Node) {
	switch {
	case n == nil || o == nil:
		*result = nil
	case n.IsLeaf():
		*result = n.Leaf().intersection(o, r, depth, count)
	case o.IsLeaf():
		*result = o.Leaf().intersection(n, r, depth, count)
	default:
		*result = &Node{}
		n.opCanonical(o, depth, count, c, result, func(a, b *Node, count *int, result **Node) {
			if intersection, ok := a.remote(b, slave.Work_OP_INTERSECTION, r, depth+1, count, c); ok {
				*result = intersection
			} else {
				a.Intersection(b, r, depth+1, count, c, result)
			}
		})
	}
}

func (n *Node) Union(o *Node, r *Resolver, depth int, matches *int, c *Cloner) *Node { //nolint:gocognit
	switch {
	case n == nil:
		return o
	case o == nil:
		return n
	case n.IsLeaf():
		n, o, r = o, n, r.Flip()
		fallthrough
	case o.IsLeaf():
		var prepared *Node
		for i := o.Leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.With(v, r, depth, NewHasher(v, depth), matches, c, &prepared)
		}
		return n
	default:
		var prepared *Node
		result := c.node(n, &prepared)
		result.SetChildren(o.mask&^n.mask, &o.children)
		if depth == c.parallelDepth {
			var m sync.Mutex
			for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				c.run(func() {
					matches := 0
					union, ok := n.children[i].remote(o.children[i], slave.Work_OP_UNION, r, depth+1, &matches, c)
					if !ok {
						union = n.children[i].Union(o.children[i], r, depth+1, &matches, c)
					} else {
						elts := map[interface{}]struct{}{}
						for i := union.Iterator(0); i.Next(); {
							elts[i.Value()] = struct{}{}
						}
						matches2 := 0
						union2 := n.children[i].Union(o.children[i], r, depth+1, &matches2, c)
						elts2 := map[interface{}]struct{}{}
						for i := union2.Iterator(0); i.Next(); {
							elts2[i.Value()] = struct{}{}
						}
						if len(elts) != len(elts2) {
							panic(fmt.Errorf("remote union wrong size: %d = %d", len(elts), len(elts2)))
						}
						remoteOnly := []interface{}{}
						for e := range elts {
							if _, has := elts2[e]; !has {
								remoteOnly = append(remoteOnly, e)
							}
						}
						localOnly := []interface{}{}
						for e := range elts2 {
							if _, has := elts[e]; !has {
								localOnly = append(localOnly, e)
							}
						}
						if len(remoteOnly) > 0 || len(localOnly) > 0 {
							panic(fmt.Errorf(
								"missing elements:\n\033[1;32mremote: %v\n\033[1;31mlocal: %v\033[0m",
								remoteOnly, localOnly,
							))
						}
					}
					result.setChildAsync(i, union, &m)
					c.update(matches)
				})
			}
		} else {
			for mask := o.mask & n.mask; mask != 0; mask = mask.Next() {
				i := mask.Index()
				result.setChild(i, n.children[i].Union(o.children[i], r, depth+1, matches, c))
			}
		}
		return result
	}
}

func (n *Node) With(v interface{}, r *Resolver, depth int, h Hasher, matches *int, c *Cloner, prepared **Node) *Node {
	switch {
	case n == nil:
		return NewLeaf(v).Node()
	case n.IsLeaf():
		return n.Leaf().with(v, r, depth, h, matches, c)
	default:
		offset := h.Hash()
		var childPrepared *Node
		child := n.children[offset].With(v, r, depth+1, h.Next(), matches, c, &childPrepared)
		if child.IsLeaf() && (n.mask|types.MaskIterator(1)<<offset).Count() == 1 {
			return child
		}
		return c.node(n, prepared).setChild(offset, child)
	}
}

var promiseNode = &Node{}

func (n *Node) Difference(o *Node, depth int, matches *int, c *Cloner, result **Node) {
	var prepared *Node
	switch {
	case n == nil || o == nil:
		*result = n
		return
	case o.IsLeaf():
		for i := o.Leaf().iterator(); i.Next(); {
			v := *i.elem()
			n = n.Without(v, depth, NewHasher(v, depth), matches, c, &prepared)
		}
		*result = n
		return
	case n.IsLeaf():
		*result = n.Leaf().difference(o, depth, matches)
		return
	default:
		// TODO: use c?
		*result = Copier.node(n, &prepared)
		(*result).clearChildren(o.mask &^ n.mask)
		n.opCanonical(o, depth, matches, c, result, func(a, b *Node, matches *int, result **Node) {
			a.Difference(b, depth+1, matches, c, result)
		})
	}
}

func (n *Node) Without(v interface{}, depth int, h Hasher, matches *int, c *Cloner, prepared **Node) *Node {
	switch {
	case n == nil:
		return n
	case n.IsLeaf():
		return n.Leaf().without(v, matches, c)
	default:
		offset := h.Hash()
		var childPrepared *Node
		child := n.children[offset].Without(v, depth+1, h.Next(), matches, c, &childPrepared)
		return c.node(n, prepared).setChild(offset, child).canonical()
	}
}

func (n *Node) Format(f fmt.State, _ rune) {
	s := n.String()
	fmt.Fprint(f, s)
	fmtutil.PadFormat(f, len(s))
}

func (n *Node) String() string {
	if n == nil {
		return "∅"
	}
	if n.IsLeaf() {
		return n.Leaf().String()
	}
	var sb strings.Builder
	deep := false
	for mask := n.mask; mask != 0; mask = mask.Next() {
		if !n.children[mask.Index()].IsLeaf() {
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
			fmt.Fprintf(&sb, "%v\n", fmtutil.IndentBlock(v.String()))
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

func (n *Node) Iterator(count int) types.Iterator {
	if n == nil {
		return exhaustedIterator{}
	}
	if n.IsLeaf() {
		return newNodeIter([]*Node{n}, count)
	}
	return newNodeIter(n.children[:], count)
}

func (n *Node) Elements(count int) []interface{} {
	elems := []interface{}{}
	for i := n.Iterator(count); i.Next(); {
		elems = append(elems, i.Value())
	}
	return elems
}

func (n *Node) remote(o *Node, op slave.Work_Op, r *Resolver, depth int, matches *int, c *Cloner) (*Node, bool) {
	if len(c.clients) > 0 {
		a, err := ToSlaveTree(n)
		if err != nil {
			panic(err)
		}
		b, err := ToSlaveTree(o)
		if err != nil {
			panic(err)
		}
		req := &slave.Work{
			Op:       op,
			A:        a,
			B:        b,
			Resolver: r.Name(),
			Depth:    int32(depth),
		}
		i := int(uintptr(n.hash(depth)) % uintptr(len(c.clients)))
		resp, err := c.clients[i].Compute(c.ctx, req)
		if err != nil {
			panic(err)
		}
		result, err := FromSlaveTree(resp.Result)
		if err != nil {
			panic(err)
		}
		*matches = int(resp.Matches)
		return result, true
	}
	return nil, false
}
