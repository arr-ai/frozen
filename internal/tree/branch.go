package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fmtutil"
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

func (b *branch) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) noderef {
	i := h.hash()
	n := b.p[i]
	if n == nil {
		n = newMutableLeaf().Node()
	}
	b.p[i] = n.Add(args, v, depth+1, h.next(), matches)
	return b.Node()
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

func (b *branch) Canonical(_ int) noderef {
	var buf [maxLeafLen]elementT
	if data := b.AppendTo(buf[:0]); data != nil {
		return newLeaf(append([]elementT{}, data...)...).Node()
	}
	return b.Node()
}

func (b *branch) Combine(args *CombineArgs, n noderef, depth int, matches *int) noderef {
	if l := n.Leaf(); l != nil {
		ret := b.Node()
		for _, e := range l.data {
			h := newHasher(e, depth)
			ret = ret.With(args, e, depth, h, matches)
		}
		return ret
	}
	b2 := n.Branch()
	ret := newBranch(nil)
	args.Parallel(depth, matches, func(i int, matches *int) bool {
		x := b.p.Get(i)
		y := b2.p.Get(i)
		ret.p[i] = x.Combine(args, y, depth+1, matches)
		return true
	})
	return ret.Node()
}

func (b *branch) Difference(args *EqArgs, n noderef, depth int, removed *int) noderef {
	if l := n.Leaf(); l != nil {
		ret := b.Node()
		for _, e := range l.data {
			h := newHasher(e, depth)
			ret = ret.Without(args, e, depth, h, removed)
		}
		return ret
	}
	ret := newBranch(nil)
	args.Parallel(depth, removed, func(i int, removed *int) bool {
		ret.p[i] = b.p.Get(i).Difference(args, n.Branch().p.Get(i), depth+1, removed)
		return true
	})
	return ret.Canonical(depth)
}

func (b *branch) Empty() bool {
	return false
}

func (b *branch) Equal(args *EqArgs, n noderef, depth int) bool {
	if n := n.Branch(); n != nil {
		return args.Parallel(depth, nil, func(i int, _ *int) bool {
			return b.p.Get(i).Equal(args, n.p.Get(i), depth+1)
		})
	}
	return false
}

func (b *branch) Get(args *EqArgs, v elementT, h hasher) *elementT {
	child := b.p.Get(h.hash())
	h2 := h.next()
	return child.Get(args, v, h2)
}

func (b *branch) Intersection(args *EqArgs, n noderef, depth int, matches *int) noderef {
	if l := n.Leaf(); l != nil {
		return l.Intersection(args.flip, b.Node(), depth, matches)
	}
	ret := newBranch(nil)
	args.Parallel(depth, matches, func(i int, matches *int) bool {
		ret.p[i] = b.p.Get(i).Intersection(args, n.Branch().p.Get(i), depth+1, matches)
		return true
	})
	return ret.Canonical(depth)
}

func (b *branch) Iterator(buf [][]noderef) Iterator {
	return b.p.Iterator(buf)
}

func (b *branch) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	var results [fanout]elementT
	args.Parallel(depth, nil, func(i int, _ *int) bool {
		if n := b.p.Get(i); !n.Empty() {
			results[i] = n.Reduce(args, depth+1, r)
		}
		return true
	})

	results2 := results[:0]
	for _, r := range results {
		if !isBlank(r) {
			results2 = append(results2, r)
		}
	}
	return r(results2...)
}

func (b *branch) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) noderef {
	i := h.hash()
	if n := b.p[i]; n != nil {
		child := b.p[i].Remove(args, v, depth+1, h.next(), matches)
		b.p[i] = child
		if child.Branch() == nil {
			var buf [maxLeafLen]elementT
			if data := b.AppendTo(buf[:0]); data != nil {
				return newMutableLeaf(data...).Node()
			}
		}
	}
	return b.Node()
}

func (b *branch) SubsetOf(args *EqArgs, n noderef, depth int) bool {
	if n.Leaf() != nil {
		return false
	}
	return args.Parallel(depth, nil, func(i int, _ *int) bool {
		return b.p.Get(i).SubsetOf(args, n.Branch().p.Get(i), depth+1)
	})
}

func (b *branch) Map(args *CombineArgs, depth int, count *int, f func(e elementT) elementT) noderef {
	var nodes [fanout]noderef
	args.Parallel(depth, count, func(i int, count *int) bool {
		nodes[i] = b.p.Get(i).Map(args, depth+1, count, f)
		return true
	})

	// log.Printf("%*s%v", 4*depth, "", nodes[0])
	acc := nodes[0]
	var duplicates int
	for _, n := range nodes[1:] {
		acc = acc.Combine(args, n, 0, &duplicates)
		// log.Printf("%*s%v -> %v", 4*depth, "", n, acc)
	}
	*count -= duplicates
	return acc
}

func (b *branch) Where(args *WhereArgs, depth int, matches *int) noderef {
	var nodes packer
	args.Parallel(depth, matches, func(i int, matches *int) bool {
		nodes[i] = b.p.Get(i).Where(args, depth+1, matches)
		return true
	})
	return (newBranch(&nodes)).Canonical(depth)
}

func (b *branch) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) noderef {
	i := h.hash()
	g := h.next()
	child := b.p.Get(i)
	return newBranch(b.p.With(i, child.With(args, v, depth+1, g, matches))).Node()
}

func (b *branch) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) noderef {
	i := h.hash()
	g := h.next()
	child := b.p.Get(i)
	return newBranch(b.p.With(i, child.Without(args, v, depth+1, g, matches))).Canonical(depth)
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

	for i, child := range b.p {
		if b.p.Get(i).Empty() {
			continue
		}
		index := branchStringIndices[i]
		if shallow {
			printf("   %s%s\n", index, fmtutil.IndentBlock(child.String()))
		} else {
			if i > 0 {
				write([]byte(" "))
			}
			printf("%s", index)
			child.Format(f, verb)
		}
	}
	write([]byte("⁆"))

	fmtutil.PadFormat(f, total)
}

func (b *branch) String() string {
	return fmt.Sprintf("%s", b)
}
