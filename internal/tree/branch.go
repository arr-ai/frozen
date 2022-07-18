package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
	"github.com/arr-ai/frozen/internal/pkg/masker"
)

const (
	fanoutBits = depth.FanoutBits
	fanout     = depth.Fanout
)

var (
	// UseRHS returns its RHS arg.
	UseRHS = func(_, b elementT) elementT { return b }

	// UseLHS returns its LHS arg.
	UseLHS = func(a, _ elementT) elementT { return a }
)

type branch struct {
	p packer
}

func newBranch(p *packer) *branch {
	b := &branch{}
	if p != nil {
		b.p = *p
	}
	return b
}

func newBranchFrom(depth int, data ...elementT) *branch {
	b := &branch{}
	for _, e := range data {
		h := newHasher(e, depth)
		b.Add(DefaultNPCombineArgs, e, depth, h)
	}
	return b
}

func (b *branch) Add(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	i := h.hash()
	if b.p.data[i] == nil {
		l := newLeaf1(v)
		b.p.SetNonNilChild(i, l)
	} else {
		h2 := h.next()
		var n node
		n, matches = b.p.data[i].Add(args, v, depth+1, h2)
		b.p.SetNonNilChild(i, n)
	}
	return b, matches
}

func (b *branch) AppendTo(dest []elementT) []elementT {
	for _, child := range b.p.data {
		if child != nil {
			if dest = child.AppendTo(dest); dest == nil {
				break
			}
		}
	}
	return dest
}

func (b *branch) Canonical(_ int) node {
	var buf [maxLeafLen]elementT
	if data := b.AppendTo(buf[:0]); data != nil {
		return newTwig(data...).Canonical(0)
	}
	return b
}

func (b *branch) Combine(args *CombineArgs, n node, depth int) (_ node, matches int) {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		_, matches = args.Parallel(depth, b.p.mask|n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			if x == nil {
				ret.p.SetNonNilChild(i, y)
			} else if y == nil {
				ret.p.SetNonNilChild(i, x)
			} else {
				var n node
				n, matches = x.Combine(args, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		return ret, matches
	case *leaf:
		for _, e := range n.slice() {
			h := newHasher(e, depth)
			var m int
			b, m = b.with(args, e, depth, h)
			matches += m
		}
		return b, matches
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Difference(args *EqArgs, n node, depth int) (_ node, matches int) {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			if y == nil {
				ret.p.data[i] = x
			} else {
				var n node
				n, matches = x.Difference(args, y, depth+1)
				ret.p.data[i] = n
			}
			return true, matches
		})
		ret.p.updateMask()
		return ret.Canonical(depth), matches
	case *leaf:
		ret := node(b)
		for _, e := range n.slice() {
			h := newHasher(e, depth)
			var m int
			ret, m = ret.Without(args, e, depth, h)
			matches += m
		}
		return ret, matches
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Empty() bool {
	return false
}

func (b *branch) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*branch); is {
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

func (b *branch) Get(args *EqArgs, v elementT, h hasher) *elementT {
	if x := b.p.data[h.hash()]; x != nil {
		h2 := h.next()
		return x.Get(args, v, h2)
	}
	return nil
}

func (b *branch) Intersection(args *EqArgs, n node, depth int) (_ node, matches int) {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		_, matches = args.Parallel(depth, b.p.mask&n.p.mask, func(i int) (_ bool, matches int) {
			x, y := b.p.data[i], n.p.data[i]
			var n node
			n, matches = x.Intersection(args, y, depth+1)
			ret.p.data[i] = n
			return true, matches
		})
		ret.p.updateMask()
		return ret.Canonical(depth), matches
	case *leaf:
		return n.Intersection(args.Flip(), b, depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Iterator(buf [][]node) Iterator {
	return b.p.Iterator(buf)
}

func (b *branch) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	var results [fanout]elementT
	args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		x := b.p.data[i]
		results[i] = x.Reduce(args, depth+1, r)
		return true, 0
	})

	results2 := results[:0]
	for _, r := range results {
		if r != zero {
			results2 = append(results2, r)
		}
	}
	return r(results2...)
}

func (b *branch) Remove(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	i := h.hash()
	if n := b.p.data[i]; n != nil {
		var n node
		n, matches = b.p.data[i].Remove(args, v, depth+1, h.next())
		b := *b
		b.p.data[i] = n
		if _, is := n.(*branch); !is {
			var buf [maxLeafLen]elementT
			if data := b.AppendTo(buf[:0]); data != nil {
				return newLeaf(data...), matches
			}
		}
		b.p.updateMask()
		return &b, matches
	}
	return b, matches
}

func (b *branch) SubsetOf(args *EqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case *branch:
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

func (b *branch) Map(args *CombineArgs, depth int, f func(e elementT) elementT) (_ node, matches int) {
	var p packer
	_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		if x := b.p.data[i]; x != nil {
			var n node
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

func (b *branch) Where(args *WhereArgs, depth int) (_ node, matches int) {
	var nodes packer
	_, matches = args.Parallel(depth, b.p.mask, func(i int) (_ bool, matches int) {
		x := b.p.data[i]
		var n node
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

func (b *branch) Vet() int {
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
					panic(errors.WrapPrefix(r, fmt.Sprintf("branch[%d]", m.FirstIndex()), 0))
				}
			}()
			if n := p.GetChild(m); n != nil {
				count += p.GetChild(m).Vet()
			} else {
				panic(errors.Errorf("nil node for mask %b", b.p.mask))
			}
		}()
	}
	return count
}

func (b *branch) With(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	return b.with(args, v, depth, h)
}

func (b *branch) with(args *CombineArgs, v elementT, depth int, h hasher) (_ *branch, matches int) {
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

func (b *branch) Without(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	i := h.hash()
	g := h.next()
	if x := b.p.data[i]; x != nil {
		var x2 node
		if x2, matches = x.Without(args, v, depth+1, g); x2 != x {
			p := b.p
			p.updateMaskBit(masker.NewMasker(i))
			return newBranch(p.WithChild(i, x2)).Canonical(depth), matches
		}
	}
	return b, matches
}

var branchStringIndices = []string{
	"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹",
	"¹⁰", "¹¹", "¹²", "¹³", "¹⁴", "¹⁵",
}

func (b *branch) Format(f fmt.State, verb rune) {
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

	var buf [20]elementT
	shallow := b.AppendTo(buf[:]) != nil

	if shallow {
		write([]byte("\n"))
	}

	for i, x := range b.p.data {
		if x == nil {
			continue
		}
		index := branchStringIndices[i]
		if shallow {
			printf("   %s%s\n", index, fu.IndentBlock(x.String()))
		} else {
			if i > 0 {
				write([]byte(" "))
			}
			printf("%s", index)
			x.Format(f, verb)
		}
	}
	write([]byte("⁆"))

	fu.PadFormat(f, total)
}

func (b *branch) String() string {
	return fmt.Sprintf("%s", b)
}

func (b *branch) clone() node {
	ret := *b
	for m := ret.p.mask; m != 0; m = m.Next() {
		i := m.FirstIndex()
		ret.p.data[i] = ret.p.data[i].clone()
	}
	return &ret
}
