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
	if len(data) == 1 {
		return &leaf1[T]{data: data[0]}
	}
	return &leaf[T]{data: append([]T(nil), data...)}
}

func (b *branch[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := newLeaf1(v)
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := nextHasher(v, h, depth)
		var n node[T]
		n, matches = b.p.data[i].Add(args, v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
	b.count += 1 - matches
	return b, matches
}

func (b *branch[T]) AddFast(v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := newLeaf1(v)
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := nextHasher(v, h, depth)
		var n node[T]
		n, matches = b.p.data[i].AddFast(v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
	b.count += 1 - matches
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
				ret.p.SetNonNilChild(i, y)
			} else if y == nil {
				ret.p.SetNonNilChild(i, x)
			} else {
				var n node[T]
				n, matches = x.Combine(args, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		ret.count = b.count + n.count - matches
		return ret, matches
	case *leaf1[T]:
		return b.with(args, n.data, depth, newHasher(n.data, depth))
	case *leaf[T]:
		ret := b
		for _, e := range n.data {
			var m int
			ret, m = ret.with(args, e, depth, newHasher(e, depth))
			matches += m
		}
		return ret, matches
	default:
		panic("unexpected node type in branch.Combine")
	}
}

func (b *branch[T]) Difference(gauge depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		ret := newBranch[T](nil)
		_, matches = gauge.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			if y == nil {
				ret.p.data[i] = x
			} else {
				var n node[T]
				n, matches = x.Difference(gauge, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		ret.count = b.count - matches
		return ret.collapse(), matches
	case *leaf1[T]:
		h := newHasher(n.data, depth)
		return b.Without(n.data, depth, h)
	case *leaf[T]:
		var ret node[T] = b
		for _, e := range n.data {
			if ret == nil {
				break
			}
			h := newHasher(e, depth)
			var m int
			ret, m = ret.Without(e, depth, h)
			matches += m
		}
		return ret, matches
	default:
		panic("unexpected node type in branch.Difference")
	}
}

func (b *branch[T]) Empty() bool {
	return false
}

func (b *branch[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	if n, is := n.(*branch[T]); is {
		if b.p.mask != n.p.mask {
			return false
		}
		equal, _ := args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			return x.Equal(args, y, depth+1), 0
		})
		return equal
	}
	return false
}

func (b *branch[T]) Get(v T, h hasher, depth int) *T {
	if x := b.p.data[h.hash()]; x != nil {
		h2 := nextHasher(v, h, depth)
		return x.Get(v, h2, depth+1)
	}
	return nil
}

func (b *branch[T]) Intersection(gauge depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		ret := newBranch[T](nil)
		_, matches = gauge.Parallel(depth, b.p.mask&n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			var n node[T]
			n, matches = x.Intersection(gauge, y, depth+1)
			ret.p.data[i] = n
			return true, matches
		})
		ret.p.updateMask()
		ret.count = matches
		return ret.collapse(), matches
	case *leaf1[T]:
		return n.Intersection(gauge, b, depth)
	case *leaf[T]:
		return n.Intersection(gauge, b, depth)
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

func (b *branch[T]) Remove(v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if n := b.p.data[i]; n != nil {
		h2 := nextHasher(v, h, depth)
		var n2 node[T]
		n2, matches = n.Remove(v, depth+1, h2)
		b := *b
		b.p.data[i] = n2
		b.p.updateMask()
		b.count -= matches
		return b.collapse(), matches
	}
	return b, matches
}

func (b *branch[T]) SubsetOf(gauge depth.Gauge, n node[T], depth int) bool {
	switch n := n.(type) {
	case *branch[T]:
		ok, _ := gauge.Parallel(depth, b.p.mask|n.p.mask, func(i int) (bool, int) {
			x, y := b.p.data[i], n.p.data[i]
			if x == nil {
				return true, 0
			}
			if y == nil {
				return false, 0
			}
			return x.SubsetOf(gauge, y, depth+1), 0
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
			return &ret, matches
		}
		return b, matches
	}
	ret := *b
	ret.p.SetNonNilChild(i, newLeaf1(v))
	ret.count = b.count + 1
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
			return &ret, matches
		}
		return b, matches
	}
	ret := *b
	ret.p.SetNonNilChild(i, newLeaf1(v))
	ret.count = b.count + 1
	return &ret, 0
}

func (b *branch[T]) Without(v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	g := nextHasher(v, h, depth)
	if x := b.p.data[i]; x != nil {
		var x2 node[T]
		if x2, matches = x.Without(v, depth+1, g); x2 != x {
			ret := *b
			ret.p.SetChild(i, x2)
			ret.count = b.count - matches
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
