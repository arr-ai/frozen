package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/pkg/errors"
	"github.com/arr-ai/frozen/v2/internal/pkg/depth"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/masker"
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
	p packer[T]
}

func newBranch[T any](p *packer[T]) *branch[T] {
	b := &branch[T]{}
	if p != nil {
		b.p = *p
	}
	return b
}

func newBranchFrom[T any](depth int, data ...T) node[T] {
	if depth >= maxTreeDepth {
		return newTwig(data...)
	}
	b := &branch[T]{}
	for _, e := range data {
		h := newHasher(e, depth)
		b.Add(DefaultNPCombineArgs[T](), e, depth, h)
	}
	return b
}

func (b *branch[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := newLeaf1(v)
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := h.next()
		var n node[T]
		n, matches = b.p.data[i].Add(args, v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
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

func (b *branch[T]) Canonical(_ int) node[T] {
	var buf [maxLeafLen]T
	if data := b.AppendTo(buf[:0]); data != nil {
		return newTwig(data...).Canonical(0)
	}
	return b
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
		return ret, matches
	case *leaf1[T]:
		return b.with(args, n.data, depth, newHasher(n.data, depth))
	case *leaf2[T]:
		var m1, m2 int
		b, m1 = b.with(args, n.data[0], depth, newHasher(n.data[0], depth))
		b, m2 = b.with(args, n.data[1], depth, newHasher(n.data[1], depth))
		return b, m1 + m2
	default:
		panic(errors.WTF)
	}
}

func (b *branch[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
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
		return ret.Canonical(depth), matches
	case *leaf1[T]:
		return b.Without(args, n.data, depth, newHasher(n.data, depth))
	case *leaf2[T]:
		ret := node[T](b)
		var m1, m2 int
		ret, m1 = ret.Without(args, n.data[0], depth, newHasher(n.data[0], depth))
		ret, m2 = ret.Without(args, n.data[1], depth, newHasher(n.data[1], depth))
		return ret, m1 + m2
	default:
		panic(errors.WTF)
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

func (b *branch[T]) Get(args *EqArgs[T], v T, h hasher) *T {
	if x := b.p.data[h.hash()]; x != nil {
		h2 := h.next()
		return x.Get(args, v, h2)
	}
	return nil
}

func (b *branch[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		ret := newBranch[T](nil)
		_, matches = args.Parallel(depth, b.p.mask&n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			var n node[T]
			n, matches = x.Intersection(args, y, depth+1)
			ret.p.data[i] = n
			return true, matches
		})
		ret.p.updateMask()
		return ret.Canonical(depth), matches
	case *leaf1[T]:
		return n.Intersection(args.Flip(), b, depth)
	case *leaf2[T]:
		return n.Intersection(args.Flip(), b, depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch[T]) Iterator(buf [][]node[T]) iterator.Iterator[T] {
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
		var n node[T]
		n, matches = b.p.data[i].Remove(args, v, depth+1, h.next())
		b.p.data[i] = n
		if _, is := n.(*branch[T]); !is {
			var buf [maxLeafLen]T
			if data := b.AppendTo(buf[:0]); data != nil {
				return newLeaf(data...), matches
			}
		}
	}
	b.p.updateMask()
	return b, matches
}

func (b *branch[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	switch n := n.(type) {
	case *branch[T]:
		ok, _ := args.Parallel(depth, b.p.mask|n.p.mask, func(i int) (bool, int) {
			x, y := b.p.data[i], n.p.data[i]
			if x == nil {
				return true, 0
			} else if y == nil {
				return false, 0
			} else {
				return x.SubsetOf(args, y, depth+1), 0
			}
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
		return newBranch(&nodes).Canonical(depth), matches
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
					panic(errors.WrapPrefix(r, fmt.Sprintf("branch[T][%d]", m.FirstIndex()), 0))
				}
			}()
			if n := p.GetChild(m); n != nil {
				count += p.GetChild(m).Vet()
			} else {
				panic(errors.Errorf("nil node[T] for mask %b", b.p.mask))
			}
		}()
	}
	return count
}

func (b *branch[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	return b.with(args, v, depth, h)
}

func (b *branch[T]) with(args *CombineArgs[T], v T, depth int, h hasher) (_ *branch[T], matches int) {
	i := h.hash()
	g := h.next()
	if x := b.p.data[i]; x != nil {
		x2, matches := x.With(args, v, depth+1, g)
		if x2 != x {
			return newBranch(b.p.WithChild(i, x2)), matches
		}
		return b, matches
	}
	return newBranch(b.p.WithChild(i, newLeaf1(v))), 0
}

func (b *branch[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	i := h.hash()
	g := h.next()
	if x := b.p.data[i]; x != nil {
		var x2 node[T]
		if x2, matches = x.Without(args, v, depth+1, g); x2 != x {
			b.p.updateMaskBit(masker.NewMasker(i))
			return newBranch(b.p.WithChild(i, x2)).Canonical(depth), matches
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

	printf := func(format string, args ...interface{}) {
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
