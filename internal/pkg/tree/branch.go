package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
)

const (
	fanoutBits = depth.FanoutBits
	fanout     = depth.Fanout
)

// UseRHS returns its RHS arg.
func UseRHS[T any](_, b T) T { return b }

// UseLHS returns its LHS arg.
func UseLHS[T any](a, _ T) T { return a }

type branch[T any] struct {
	p     packer[T]
	count int
	h0    H128
}

func (b *branch[T]) H0() H128 { return b.h0 }

// childrenH0 computes the XOR of all children's H0 values.
func childrenH0[T any](p *packer[T]) H128 {
	var h0 H128
	for m := p.mask; m != 0; m = m.Next() {
		h0 = h0.xor(p.data[m.FirstIndex()].H0())
	}
	return h0
}

func newBranch[T any](p *packer[T]) *branch[T] {
	b := &branch[T]{}
	if p != nil {
		b.p = *p
	}
	return b
}

// nextHasher advances h for the next depth, recomputing when crossing a round boundary.
func nextHasher[T any](v T, h hasher, depth int) hasher {
	if (depth+1)%levelsPerRound == 0 {
		return newHasher(v, depth+1)
	}
	return h.next()
}

// nextHasherWith is like nextHasher but uses a pre-resolved hash function,
// avoiding the sync.Map lookup in newHasher.
func nextHasherWith[T any](v T, h hasher, depth int, hf func(T, uintptr) uintptr) hasher {
	if (depth+1)%levelsPerRound == 0 {
		return newHasherWith(v, depth+1, hf)
	}
	return h.next()
}

// collapse returns a simpler node if possible after child removal.
// Empty branches become nil; branches with ≤maxLeafLen total elements
// become a leaf1 or leaf.
func (b *branch[T]) collapse() node[T] {
	if b.count == 0 {
		return nil
	}
	if b.count > maxLeafLen {
		return b
	}
	var buf [maxLeafLen]T
	data := b.AppendTo(buf[:0])
	// Same elements → same XOR hash. Pass b.h0 through.
	return leafCanonicalWithHash(append([]T(nil), data...), b.h0)
}

func (b *branch[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := &leaf1[T]{data: v} // h0 left stale — Builder.Finish() recomputes.
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := nextHasher(v, h, depth)
		var n node[T]
		n, matches = b.p.data[i].Add(args, v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
	b.count += 1 - matches
	// h0 left stale — Builder.Finish() recomputes.
	return b, matches
}

func (b *branch[T]) AddFast(v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := &leaf1[T]{data: v} // h0 left stale — Builder.Finish() recomputes.
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := nextHasher(v, h, depth)
		var n node[T]
		n, matches = b.p.data[i].AddFast(v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
	b.count += 1 - matches
	// h0 left stale — Builder.Finish() recomputes.
	return b, matches
}

func (b *branch[T]) AppendTo(dest []T) []T {
	for _, child := range b.p.data {
		if child != nil {
			if dest = child.AppendTo(dest); dest == nil {
				break
			}
		}
	}
	return dest
}

func (b *branch[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		ret := newBranch[T](nil)
		_, matches = args.Parallel(depth, b.p.mask|n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			if x == nil {
				ret.p.data[i] = y
			} else if y == nil {
				ret.p.data[i] = x
			} else {
				var n node[T]
				n, matches = x.Combine(args, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		ret.count = b.count + n.count - matches
		ret.h0 = childrenH0(&ret.p)
		return ret, matches
	case *leaf1[T]:
		return b.with(args, n.data, depth, hasherFromCached(n.h0, depth, n.data, args.hash))
	case *leaf2[T]:
		hashes := [2]H128{n.ha, n.hb()}
		ret := b
		for i, e := range n.data {
			var m int
			ret, m = ret.with(args, e, depth, hasherFromCached(hashes[i], depth, e, args.hash))
			matches += m
		}
		return ret, matches
	case *leaf[T]:
		ret := b
		for _, e := range n.data {
			var m int
			ret, m = ret.with(args, e, depth, newHasherWith(e, depth, args.hash))
			matches += m
		}
		return ret, matches
	default:
		panic("unexpected node type in branch.Combine")
	}
}

func (b *branch[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		if b == n || (b.count == n.count && b.Equal(args, node[T](n), depth)) {
			return nil, b.count
		}
		ret := newBranch[T](nil)
		_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			if y == nil {
				ret.p.data[i] = x
			} else {
				var n node[T]
				n, matches = x.Difference(args, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		ret.count = b.count - matches
		ret.h0 = childrenH0(&ret.p)
		return ret.collapse(), matches
	case *leaf1[T]:
		h := hasherFromCached(n.h0, depth, n.data, args.hash)
		return b.Without(args, n.data, depth, h)
	case *leaf2[T]:
		return differenceByElement(args, node[T](b), n.data[:], depth)
	case *leaf[T]:
		return differenceByElement(args, node[T](b), n.data, depth)
	default:
		panic("unexpected node type in branch.Difference")
	}
}

func differenceByElement[T any](args *EqArgs[T], base node[T], elements []T, depth int) (node[T], int) {
	var matches int
	ret := base
	for _, e := range elements {
		if ret == nil {
			break
		}
		h := newHasherWith(e, depth, args.hash)
		var m int
		ret, m = ret.Without(args, e, depth, h)
		matches += m
	}
	return ret, matches
}

func (b *branch[T]) Empty() bool {
	return false
}

func (b *branch[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	if n, is := n.(*branch[T]); is {
		if b.h0 != n.h0 || b.p.mask != n.p.mask {
			return false
		}
		if args.fullHash && !b.h0.isZero() {
			return true
		}
		equal, _ := args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			return x.Equal(args, y, depth+1), 0
		})
		return equal
	}
	return false
}

func (b *branch[T]) Get(args *EqArgs[T], v T, h hasher, depth int) *T {
	if x := b.p.data[h.hash()]; x != nil {
		h2 := nextHasherWith(v, h, depth, args.hash)
		return x.Get(args, v, h2, depth+1)
	}
	return nil
}

func (b *branch[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		if b == n || (b.count == n.count && b.Equal(args, node[T](n), depth)) {
			return b, b.count
		}
		ret := newBranch[T](nil)
		_, matches = args.Parallel(depth, b.p.mask&n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			var n node[T]
			n, matches = x.Intersection(args, y, depth+1)
			ret.p.data[i] = n
			return true, matches
		})
		ret.p.updateMask()
		ret.count = matches
		ret.h0 = childrenH0(&ret.p)
		return ret.collapse(), matches
	case *leaf1[T]:
		return n.Intersection(args, b, depth)
	case *leaf2[T]:
		return n.Intersection(args, b, depth)
	case *leaf[T]:
		return n.Intersection(args, b, depth)
	default:
		panic("unexpected node type in branch.Intersection")
	}
}

func (b *branch[T]) Iterator(buf [][]node[T]) Iterator[T] {
	return b.p.Iterator(buf)
}

func (b *branch[T]) Reduce(args NodeArgs, depth int, r func(values ...T) T) T {
	var results [fanout]T
	args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		x := b.p.data[i]
		results[i] = x.Reduce(args, depth+1, r)
		return true, 0
	})

	results2 := results[:0]
	for i := b.p.mask; i != 0; i = i.Next() {
		results2 = append(results2, results[i.FirstIndex()])
	}
	return r(results2...)
}

func (b *branch[T]) Remove(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if n := b.p.data[i]; n != nil {
		h2 := nextHasherWith(v, h, depth, args.hash)
		var n2 node[T]
		n2, matches = n.Remove(args, v, depth+1, h2)
		b := *b
		b.p.data[i] = n2
		b.p.updateMask()
		b.count -= matches
		// h0 left stale — Builder.Finish() recomputes.
		return b.collapse(), matches
	}
	return b, matches
}

func (b *branch[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	switch n := n.(type) {
	case *branch[T]:
		ok, _ := args.Parallel(depth, b.p.mask|n.p.mask, func(i int) (bool, int) {
			x, y := b.p.data[i], n.p.data[i]
			if x == nil {
				return true, 0
			}
			if y == nil {
				return false, 0
			}
			return x.SubsetOf(args, y, depth+1), 0
		})
		return ok
	default:
		return false
	}
}

func (b *branch[T]) Map(args *CombineArgs[T], depth int, f func(e T) T) (_ node[T], matches int) {
	var p packer[T]
	_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		if x := b.p.data[i]; x != nil {
			var n node[T]
			n, matches = x.Map(args, depth+1, f)
			p.data[i] = n
		}
		return true, matches
	})
	p.updateMask()
	if p.mask == 0 {
		return
	}

	acc := p.GetChild(p.mask)
	var duplicates int
	for m := p.mask.Next(); m != 0; m = m.Next() {
		var d int
		acc, d = acc.Combine(args, p.GetChild(m), 0)
		duplicates += d
	}
	matches -= duplicates
	return acc, matches
}

func (b *branch[T]) Where(args *WhereArgs[T], depth int) (_ node[T], matches int) {
	var nodes packer[T]
	_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		x := b.p.data[i]
		var n node[T]
		n, matches = x.Where(args, depth+1)
		nodes.data[i] = n
		return true, matches
	})
	nodes.updateMask()
	if nodes != b.p {
		ret := newBranch(&nodes)
		ret.count = matches
		ret.h0 = childrenH0(&ret.p)
		return ret.collapse(), matches
	}
	return b, matches
}

func (b *branch[T]) Vet() int {
	p := b.p
	p.updateMask()
	if p.mask != b.p.mask {
		panic("stale mask")
	}
	count := 0
	for m := b.p.mask; m != 0; m = m.Next() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if err := r.(error); err != nil {
						panic(fmt.Errorf("branch[T][%d]: %w", m.FirstIndex(), err))
					}
					panic(fmt.Errorf("branch[T][%d]: %v", m.FirstIndex(), r))
				}
			}()
			if n := p.GetChild(m); n != nil {
				count += p.GetChild(m).Vet()
			} else {
				panic(fmt.Errorf("nil node[T] for mask %b", b.p.mask))
			}
		}()
	}
	if count != b.count {
		panic(fmt.Errorf("branch count mismatch: tracked %d, measured %d", b.count, count))
	}
	return count
}

func (b *branch[T]) vetH0(hf func(T, uintptr) uintptr) {
	want := childrenH0(&b.p)
	if b.h0 != want {
		panic(fmt.Errorf("branch h0 mismatch: stored %v, computed %v", b.h0, want))
	}
	for m := b.p.mask; m != 0; m = m.Next() {
		vetNodeH0(b.p.data[m.FirstIndex()], hf)
	}
}

func (b *branch[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	return b.with(args, v, depth, h)
}

func (b *branch[T]) with(args *CombineArgs[T], v T, depth int, h hasher) (_ *branch[T], matches int) {
	i := h.hash()
	g := nextHasher(v, h, depth)
	if x := b.p.data[i]; x != nil {
		x2, matches := x.With(args, v, depth+1, g)
		if x2 != x {
			ret := *b
			ret.p.data[i] = x2
			ret.count = b.count + 1 - matches
			ret.h0 = b.h0.xor(x.H0()).xor(x2.H0())
			return &ret, matches
		}
		return b, matches
	}
	l := newLeaf1WithHash(v, newElemH128(v, args.hash))
	ret := *b
	ret.p.SetNonNilChild(i, l)
	ret.count = b.count + 1
	ret.h0 = b.h0.xor(l.h0)
	return &ret, 0
}

func (b *branch[T]) WithFast(v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	g := nextHasher(v, h, depth)
	if x := b.p.data[i]; x != nil {
		x2, matches := x.WithFast(v, depth+1, g)
		if x2 != x {
			ret := *b
			ret.p.data[i] = x2
			ret.count = b.count + 1 - matches
			ret.h0 = b.h0.xor(x.H0()).xor(x2.H0())
			return &ret, matches
		}
		return b, matches
	}
	l := newLeaf1(v)
	ret := *b
	ret.p.SetNonNilChild(i, l)
	ret.count = b.count + 1
	ret.h0 = b.h0.xor(l.h0)
	return &ret, 0
}

type spineEntry[T any] struct {
	b     *branch[T]
	index int
}

// buildSpine batch-allocates branch copies for the recorded path and wires
// them bottom-up with child as the new node at the bottom level.
func buildSpine[T any](path []spineEntry[T], child node[T], countDelta int) *branch[T] {
	pathLen := len(path)
	spine := make([]branch[T], pathLen)
	bottom := pathLen - 1
	spine[bottom] = *path[bottom].b
	oldChild := path[bottom].b.p.data[path[bottom].index]
	spine[bottom].p.SetNonNilChild(path[bottom].index, child)
	spine[bottom].count = path[bottom].b.count + countDelta
	var oldH0 H128
	if oldChild != nil {
		oldH0 = oldChild.H0()
	}
	spine[bottom].h0 = path[bottom].b.h0.xor(oldH0).xor(child.H0())

	for j := bottom - 1; j >= 0; j-- {
		s := path[j]
		spine[j] = *s.b
		spine[j].p.data[s.index] = &spine[j+1]
		spine[j].count = s.b.count + countDelta
		spine[j].h0 = s.b.h0.xor(s.b.p.data[s.index].H0()).xor(spine[j+1].h0)
	}
	return &spine[0]
}

// withFastBatched performs a With operation starting from a branch root using
// batched allocation: all spine branch copies are allocated in a single slice
// instead of one heap allocation per level.
func withFastBatched[T any](root *branch[T], v T, h hasher, hf func(T, uintptr) uintptr) (node[T], int) {
	// Stack-allocated path buffer. levelsPerRound*2 covers two full hash
	// rounds, far more than any realistic tree depth.
	var path [levelsPerRound * 2]spineEntry[T]
	pathLen := 0

	cur := root
	curHash := h
	depth := 0

	for {
		i := curHash.hash()
		child := cur.p.data[i]
		path[pathLen] = spineEntry[T]{b: cur, index: i}
		pathLen++

		nextH := nextHasherWith(v, curHash, depth, hf)

		if child == nil {
			return buildSpine(path[:pathLen], node[T](newLeaf1WithHash(v, newElemH128(v, hf))), 1), 0
		}

		if b2, ok := child.(*branch[T]); ok {
			cur = b2
			curHash = nextH
			depth++
			continue
		}

		// Hit a leaf node — delegate to its WithFast.
		x2, matches := child.WithFast(v, depth+1, nextH)
		if x2 == child {
			return root, matches
		}
		return buildSpine(path[:pathLen], x2, 1-matches), matches
	}
}

// buildSpineWithout batch-allocates branch copies for Without. Unlike
// buildSpine, the bottom branch uses SetChild (handles nil) and is collapsed
// if its count drops to ≤maxLeafLen.
func buildSpineWithout[T any](path []spineEntry[T], child node[T], matches int) node[T] {
	pathLen := len(path)
	spine := make([]branch[T], pathLen)
	bottom := pathLen - 1

	spine[bottom] = *path[bottom].b
	oldChild := path[bottom].b.p.data[path[bottom].index]
	spine[bottom].p.SetChild(path[bottom].index, child)
	spine[bottom].count = path[bottom].b.count - matches
	var childH0 H128
	if child != nil {
		childH0 = child.H0()
	}
	spine[bottom].h0 = path[bottom].b.h0.xor(oldChild.H0()).xor(childH0)

	// The bottom branch may collapse to a leaf or nil.
	var bottomNode node[T] = &spine[bottom]
	if spine[bottom].count <= maxLeafLen {
		bottomNode = spine[bottom].collapse()
	}

	if pathLen == 1 {
		return bottomNode
	}

	var bottomH0 H128
	if bottomNode != nil {
		bottomH0 = bottomNode.H0()
	}

	// Wire the rest of the spine. If bottom collapsed to a non-branch,
	// the next level up must use SetChild (handles nil).
	spine[bottom-1] = *path[bottom-1].b
	spine[bottom-1].p.SetChild(path[bottom-1].index, bottomNode)
	spine[bottom-1].count = path[bottom-1].b.count - matches
	spine[bottom-1].h0 = path[bottom-1].b.h0.xor(path[bottom-1].b.p.data[path[bottom-1].index].H0()).xor(bottomH0)

	for j := bottom - 2; j >= 0; j-- {
		s := path[j]
		spine[j] = *s.b
		spine[j].p.data[s.index] = &spine[j+1]
		spine[j].count = s.b.count - matches
		spine[j].h0 = s.b.h0.xor(s.b.p.data[s.index].H0()).xor(spine[j+1].h0)
	}
	return &spine[0]
}

// withoutBatched performs a Without operation starting from a branch root using
// batched allocation: all spine branch copies are allocated in a single slice
// instead of one heap allocation per level.
func withoutBatched[T any](root *branch[T], args *EqArgs[T], v T, h hasher) (node[T], int) {
	var path [levelsPerRound * 2]spineEntry[T]
	pathLen := 0

	cur := root
	curHash := h
	curDepth := 0

	for {
		i := curHash.hash()
		child := cur.p.data[i]

		if child == nil {
			return root, 0
		}

		path[pathLen] = spineEntry[T]{b: cur, index: i}
		pathLen++

		nextH := nextHasherWith(v, curHash, curDepth, args.hash)

		if b2, ok := child.(*branch[T]); ok {
			cur = b2
			curHash = nextH
			curDepth++
			continue
		}

		// Hit a leaf node — delegate to its Without.
		x2, matches := child.Without(args, v, curDepth+1, nextH)
		if x2 == child {
			return root, 0
		}
		return buildSpineWithout(path[:pathLen], x2, matches), matches
	}
}

func (b *branch[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	g := nextHasherWith(v, h, depth, args.hash)
	if x := b.p.data[i]; x != nil {
		var x2 node[T]
		if x2, matches = x.Without(args, v, depth+1, g); x2 != x {
			ret := *b
			ret.p.SetChild(i, x2)
			ret.count = b.count - matches
			var x2H0 H128
			if x2 != nil {
				x2H0 = x2.H0()
			}
			ret.h0 = b.h0.xor(x.H0()).xor(x2H0)
			return ret.collapse(), matches
		}
	}
	return b, matches
}

var branchStringIndices = []string{
	"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹",
	"¹⁰", "¹¹", "¹²", "¹³", "¹⁴", "¹⁵",
}

func (b *branch[T]) Format(f fmt.State, verb rune) {
	total := 0

	printf := func(format string, args ...any) {
		n, err := fmt.Fprintf(f, format, args...)
		if err != nil {
			panic(err)
		}
		total += n
	}
	write := func(b []byte) {
		n, err := f.Write(b)
		if err != nil {
			panic(err)
		}
		total += n
	}

	write([]byte("⁅"))

	var buf [20]T
	shallow := b.AppendTo(buf[:]) != nil

	if shallow {
		write([]byte("\n"))
	}

	first := true
	for i, x := range b.p.data {
		if x == nil {
			continue
		}
		index := branchStringIndices[i]
		if shallow {
			printf("   %s%s\n", index, fu.IndentBlock(x.String()))
		} else {
			if !first {
				write([]byte(" "))
			} else {
				first = false
			}
			printf("%s", index)
			x.Format(f, verb)
		}
	}
	write([]byte("⁆"))

	fu.PadFormat(f, total)
}

func (b *branch[T]) String() string {
	return fmt.Sprintf("%s", b)
}

func (b *branch[T]) clone() node[T] {
	ret := *b
	for m := ret.p.mask; m != 0; m = m.Next() {
		i := m.FirstIndex()
		ret.p.data[i] = ret.p.data[i].clone()
	}
	return &ret
}
