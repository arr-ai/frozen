package frozen

import (
	"fmt"
	"log"
	"strings"
)

const fanout = 1 << fanoutBits

var (
	useRHS = func(_, b interface{}) interface{} { return b }
	useLHS = func(a, _ interface{}) interface{} { return a }
)

type branch struct {
	p packed
}

func (b branch) Canonical(_ int) node {
	if b.p.mask.count() == 0 {
		return emptyNode{}
	}
	if n := b.CountUpTo(9); n < 9 {
		l := make(leaf, 0, n)
		for i := b.Iterator(make([]packed, 0, 8)); i.Next(); {
			l = append(l, i.Value())
		}
		return l
	}
	return b
}

func (b branch) Combine(args *combineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case emptyNode:
		return b
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.With(args, e, depth, newHasher(e, depth), matches)
		}
		return result
	case branch:
		return b.transformPair(n, b.p.mask|n.p.mask, args.parallel(depth), matches,
			func(_ masker, x, y node, matches *int) node {
				return x.Combine(args, y, depth+1, matches)
			}).Canonical(depth)
	default:
		panic(WTF)
	}
}

func (b branch) CountUpTo(max int) int {
	total := 0
	for _, child := range b.p.data {
		total += child.CountUpTo(max)
		if total >= max {
			break
		}
	}
	return total
}

func (b branch) Difference(args *eqArgs, n node, depth int, removed *int) node {
	switch n := n.(type) {
	case emptyNode:
		return b
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.Without(args, e, depth, newHasher(e, depth), removed)
		}
		return result
	case branch:
		return b.transformPair(n, b.p.mask, args.parallel(depth), removed,
			func(_ masker, a, b node, matches *int) node {
				return a.Difference(args, b, depth+1, matches)
			}).Canonical(depth)
	default:
		panic(WTF)
	}
}

func (b branch) Equal(args *eqArgs, n node, depth int) bool {
	if n, is := n.(branch); is {
		return b.allPair(n, b.p.mask|n.p.mask, args.parallel(depth),
			func(m masker, x, y node) bool {
				return x.Equal(args, y, depth+1)
			})
	}
	return false
}

func (b branch) Get(args *eqArgs, v interface{}, h hasher) *interface{} {
	return b.p.Get(newMasker(h.hash())).Get(args, v, h.next())
}

func (b branch) Intersection(args *eqArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case emptyNode:
		return n
	case leaf:
		return n.Intersection(args.flip, b, depth, matches)
	case branch:
		return b.transformPair(n, b.p.mask&n.p.mask, args.parallel(depth), matches,
			func(_ masker, a, b node, matches *int) node {
				return a.Intersection(args, b, depth+1, matches)
			}).Canonical(depth)
	default:
		panic(WTF)
	}
}

func (b branch) Iterator(buf []packed) Iterator {
	return b.p.Iterator(buf)
}

func (b branch) Reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{} {
	var results [fanout]interface{}
	b.p.All(args.parallel(depth), func(m masker, child node) bool {
		results[m.index()] = child.Reduce(args, depth+1, r)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	for m = m.next(); m != 0; m = m.next() {
		acc = r(acc, results[m.index()])
	}
	return acc
}

func (b branch) SubsetOf(args *eqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case emptyNode, leaf:
		return false
	case branch:
		return b.allPair(n, b.p.mask, args.parallel(depth),
			func(m masker, x, y node) bool {
				return x.SubsetOf(args, y, depth+1)
			})
	default:
		panic(WTF)
	}
}

func (b branch) Transform(args *combineArgs, depth int, count *int, f func(v interface{}) interface{}) node {
	var results [fanout]node
	var counts [fanout]int
	b.p.All(args.parallel(depth), func(m masker, child node) bool {
		i := m.index()
		results[i] = child.Transform(args, depth+1, &counts[i], f)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	var duplicates int
	for m = m.next(); m != 0; m = m.next() {
		acc = acc.Combine(args, results[m.index()], 0, &duplicates)
	}

	for _, c := range counts {
		*count += c
	}
	*count -= duplicates

	return acc
}

func (b branch) Where(args *whereArgs, depth int, matches *int) node {
	return b.transformImpl(args.parallel(depth), matches,
		func(_ masker, n node, matches *int) node {
			return n.Where(args, depth+1, matches)
		}).Canonical(depth)
}

func (b branch) With(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node {
	i := newMasker(h.hash())
	return branch{p: b.p.With(i, b.p.Get(i).With(args, v, depth+1, h.next(), matches))}
}

func (b branch) Without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) node {
	i := newMasker(h.hash())
	child := b.p.Get(i).Without(args, v, depth+1, h.next(), matches)
	return branch{p: b.p.With(i, child)}.Canonical(depth)
}

func (b branch) allPair(
	o branch,
	mask masker,
	parallel bool,
	op func(m masker, a, b node) bool,
) bool {
	ok := b.p.AllPair(o.p, mask, parallel, func(m masker, x, y node) bool {
		return op(m, x, y)
	})
	return ok
}

func (b branch) transformPair(
	o branch,
	mask masker,
	parallel bool,
	matches *int,
	op func(m masker, a, b node, matches *int) node,
) node {
	var allMatches [fanout]int
	result := branch{p: b.p.TransformPair(o.p, mask, parallel, func(m masker, x, y node) node {
		return op(m, x, y, &allMatches[m.index()])
	})}
	for _, m := range allMatches {
		*matches += m
	}
	return result
}

func (b branch) transformImpl(
	parallel bool,
	matches *int,
	op func(m masker, n node, matches *int) node,
) node {
	var allMatches [fanout]int
	result := branch{p: b.p.Transform(parallel, func(m masker, n node) node {
		return op(m, n, &allMatches[m.index()])
	})}
	for _, m := range allMatches {
		*matches += m
	}
	return result
}

func (b branch) Format(f fmt.State, _ rune) {
	s := b.String()
	fmt.Fprint(f, s)
	padFormat(f, len(s))
}

var branchStringIndices = []string{
	"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹",
	"¹⁰", "¹¹", "¹²", "¹³", "¹⁴", "¹⁵",
}

func (b branch) String() string {
	var sb strings.Builder
	deep := b.CountUpTo(20) >= 20
	sb.WriteRune('⁅')
	if deep {
		sb.WriteString("\n")
	}

	m := b.p.mask

	defer func() {
		if recover() != nil {
			log.Print(m, b.p)
		}
	}()

	for i, child := range b.p.data {
		index := branchStringIndices[m.index()]
		m = m.next()
		if deep {
			fmt.Fprintf(&sb, "   %s%s\n", index, indentBlock(child.String()))
		} else {
			if i > 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprintf(&sb, "%s%v", index, child)
		}
	}
	sb.WriteRune('⁆')
	return sb.String()
}
