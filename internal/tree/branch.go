package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
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
		var matches int
		b = b.with(DefaultNPCombineArgs, e, depth, h, &matches)
	}
	return b
}

func (b *branch) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
	i := h.hash()
	if b.p[i] == nil {
		b.p[i] = newLeaf1(v)
	} else {
		b.p[i] = b.p[i].Add(args, v, depth+1, h.next(), matches)
	}
	return b
}

func (b *branch) AppendTo(dest []elementT) []elementT {
	for _, child := range b.p {
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

func (b *branch) Combine(args *CombineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		args.Parallel(depth, matches, func(i int, matches *int) bool {
			x, y := b.p[i], n.p[i]
			if x == nil {
				ret.p[i] = y
			} else if y == nil {
				ret.p[i] = x
			} else {
				ret.p[i] = x.Combine(args, y, depth+1, matches)
			}
			return true
		})
		return ret
	case *leaf:
		for _, e := range n.slice() {
			h := newHasher(e, depth)
			b = b.with(args, e, depth, h, matches)
		}
		return b
	case *twig:
		for _, e := range n.data {
			h := newHasher(e, depth)
			b = b.with(args, e, depth, h, matches)
		}
		return b
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Difference(args *EqArgs, n node, depth int, removed *int) node {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		args.Parallel(depth, removed, func(i int, removed *int) bool {
			x, y := b.p[i], n.p[i]
			if x == nil || y == nil {
				ret.p[i] = x
			} else {
				ret.p[i] = x.Difference(args, y, depth+1, removed)
			}
			return true
		})
		return ret.Canonical(depth)
	case *leaf:
		ret := node(b)
		for _, e := range n.slice() {
			h := newHasher(e, depth)
			ret = ret.Without(args, e, depth, h, removed)
		}
		return ret
	case *twig:
		ret := node(b)
		for _, e := range n.data {
			h := newHasher(e, depth)
			ret = ret.Without(args, e, depth, h, removed)
		}
		return ret
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Empty() bool {
	return false
}

func (b *branch) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*branch); is {
		return args.Parallel(depth, nil, func(i int, _ *int) bool {
			x, y := b.p[i], n.p[i]
			return (x == nil) && (y == nil) ||
				(x != nil) && (y != nil) && x.Equal(args, y, depth+1)
		})
	}
	return false
}

func (b *branch) Get(args *EqArgs, v elementT, h hasher) *elementT {
	if x := b.p[h.hash()]; x != nil {
		h2 := h.next()
		return x.Get(args, v, h2)
	}
	return nil
}

func (b *branch) Intersection(args *EqArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case *branch:
		ret := newBranch(nil)
		args.Parallel(depth, matches, func(i int, matches *int) bool {
			x, y := b.p[i], n.p[i]
			if x == nil || y == nil {
				ret.p[i] = nil
			} else {
				ret.p[i] = x.Intersection(args, y, depth+1, matches)
			}
			return true
		})
		return ret.Canonical(depth)
	case *leaf:
		return n.Intersection(args.flip, b, depth, matches)
	case *twig:
		return n.Intersection(args.flip, b, depth, matches)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Iterator(buf [][]node) Iterator {
	return b.p.Iterator(buf)
}

func (b *branch) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	var results [fanout]elementT
	args.Parallel(depth, nil, func(i int, _ *int) bool {
		if x := b.p[i]; x != nil {
			results[i] = x.Reduce(args, depth+1, r)
		}
		return true
	})

	results2 := results[:0]
	for _, r := range results {
		if r != zero {
			results2 = append(results2, r)
		}
	}
	return r(results2...)
}

func (b *branch) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
	i := h.hash()
	if n := b.p[i]; n != nil {
		child := b.p[i].Remove(args, v, depth+1, h.next(), matches)
		b.p[i] = child
		if _, is := child.(*branch); !is {
			var buf [maxLeafLen]elementT
			if data := b.AppendTo(buf[:0]); data != nil {
				return newLeaf(data...)
			}
		}
	}
	return b
}

func (b *branch) SubsetOf(args *EqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case *branch:
		return args.Parallel(depth, nil, func(i int, _ *int) bool {
			x, y := b.p[i], n.p[i]
			if x == nil {
				return true
			} else if y == nil {
				return false
			} else {
				return x.SubsetOf(args, y, depth+1)
			}
		})
	default:
		return false
	}
}

func (b *branch) Map(args *CombineArgs, depth int, count *int, f func(e elementT) elementT) node {
	var nodes [fanout]node
	args.Parallel(depth, count, func(i int, count *int) bool {
		if x := b.p[i]; x != nil {
			nodes[i] = x.Map(args, depth+1, count, f)
		}
		return true
	})

	i := 0
	for ; i < len(nodes); i++ {
		if nodes[i] != nil {
			break
		}
	}
	acc := nodes[i]
	var duplicates int
	for _, n := range nodes[i+1:] {
		if n != nil {
			acc = acc.Combine(args, n, 0, &duplicates)
		}
	}
	*count -= duplicates
	return acc
}

func (b *branch) Where(args *WhereArgs, depth int, matches *int) node {
	var nodes packer
	args.Parallel(depth, matches, func(i int, matches *int) bool {
		if x := b.p[i]; x != nil {
			nodes[i] = x.Where(args, depth+1, matches)
		}
		return true
	})
	if nodes != b.p {
		return newBranch(&nodes).Canonical(depth)
	}
	return b
}

func (b *branch) Vet() {
	for i, x := range b.p {
		func() {
			defer func() {
				if r := recover(); r != nil {
					panic(errors.WrapPrefix(r, fmt.Sprintf("branch[%d]", i), 0))
				}
			}()
			if x != nil {
				x.Vet()
			}
		}()
	}
}

func (b *branch) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
	return b.with(args, v, depth, h, matches)
}

func (b *branch) with(args *CombineArgs, v elementT, depth int, h hasher, matches *int) *branch {
	i := h.hash()
	g := h.next()
	if x := b.p[i]; x != nil {
		if x2 := x.With(args, v, depth+1, g, matches); x2 != x {
			return newBranch(b.p.With(i, x2))
		}
		return b
	}
	return newBranch(b.p.With(i, newLeaf1(v)))
}

func (b *branch) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
	i := h.hash()
	g := h.next()
	if x := b.p[i]; x != nil {
		if x2 := x.Without(args, v, depth+1, g, matches); x2 != x {
			return newBranch(b.p.With(i, x2)).Canonical(depth)
		}
	}
	return b
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

	for i, x := range b.p {
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
