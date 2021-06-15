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

func (b branch) canonical(_ int) (out node) {
	defer vet(&out)

	if b.p.mask.count() == 0 {
		return emptyNode{}
	}
	if n := b.countUpTo(9); n < 9 {
		result := make(leaf, 0, n)
		for i := b.iterator(make([]packed, 0, 8)); i.Next(); {
			result = append(result, i.Value())
		}
		return result
	}
	return b
}

func (b branch) combine(args *combineArgs, n node, depth int, matches *int) (out node) {
	defer vet(&out)

	switch n := n.(type) {
	case emptyNode:
		return b
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.with(args, e, depth, newHasher(e, depth), matches)
		}
		return result
	case branch:
		return b.transformPair(n, b.p.mask|n.p.mask, args.parallel(depth), matches,
			func(_ masker, x, y node, matches *int) node {
				return x.combine(args, y, depth+1, matches)
			}).canonical(depth)
	default:
		panic(wtf)
	}
}

func (b branch) countUpTo(max int) int {
	total := 0
	for _, child := range b.p.data {
		total += child.countUpTo(max)
		if total >= max {
			break
		}
	}
	return total
}

func (b branch) difference(args *eqArgs, n node, depth int, removed *int) (out node) {
	defer vet(&out)

	switch n := n.(type) {
	case emptyNode:
		return b
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.without(args, e, depth, newHasher(e, depth), removed)
		}
		return result
	case branch:
		return b.transformPair(n, b.p.mask, args.parallel(depth), removed,
			func(_ masker, a, b node, matches *int) node {
				return a.difference(args, b, depth+1, matches)
			}).canonical(depth)
	default:
		panic(wtf)
	}
}

func (b branch) equal(args *eqArgs, n node, depth int) bool {
	if n, is := n.(branch); is {
		return b.allPair(n, b.p.mask|n.p.mask, args.parallel(depth),
			func(m masker, x, y node) bool {
				return x.equal(args, y, depth+1)
			})
	}
	return false
}

func (b branch) get(args *eqArgs, v interface{}, h hasher) *interface{} {
	return b.p.get(newMasker(h.hash())).get(args, v, h.next())
}

func (b branch) intersection(args *eqArgs, n node, depth int, matches *int) (out node) {
	defer vet(&out)

	switch n := n.(type) {
	case emptyNode:
		return n
	case leaf:
		return n.intersection(args.flip, b, depth, matches)
	case branch:
		return b.transformPair(n, b.p.mask&n.p.mask, args.parallel(depth), matches,
			func(_ masker, a, b node, matches *int) node {
				return a.intersection(args, b, depth+1, matches)
			}).canonical(depth)
	default:
		panic(wtf)
	}
}

func (b branch) isSubsetOf(args *eqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case emptyNode, leaf:
		return false
	case branch:
		return b.allPair(n, b.p.mask, args.parallel(depth),
			func(m masker, x, y node) bool {
				return x.isSubsetOf(args, y, depth+1)
			})
	default:
		panic(wtf)
	}
}

func (b branch) iterator(buf []packed) Iterator {
	return b.p.iterator(buf)
}

func (b branch) reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{} {
	var results [fanout]interface{}
	b.p.all(args.parallel(depth), func(m masker, child node) bool {
		results[m.index()] = child.reduce(args, depth+1, r)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	for m = m.next(); m != 0; m = m.next() {
		acc = r(acc, results[m.index()])
	}
	return acc
}

func (b branch) transform(args *combineArgs, depth int, count *int, f func(v interface{}) interface{}) node {
	var results [fanout]node
	var counts [fanout]int
	b.p.all(args.parallel(depth), func(m masker, child node) bool {
		i := m.index()
		results[i] = child.transform(args, depth+1, &counts[i], f)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	var duplicates int
	for m = m.next(); m != 0; m = m.next() {
		acc = acc.combine(args, results[m.index()], 0, &duplicates)
	}

	for _, c := range counts {
		*count += c
	}
	*count -= duplicates

	return acc
}

func (b branch) vet() node {
	return b
}

func (b branch) where(args *whereArgs, depth int, matches *int) (out node) {
	defer vet(&out)

	return b.transformImpl(args.parallel(depth), matches,
		func(_ masker, n node, matches *int) node {
			return n.where(args, depth+1, matches)
		}).canonical(depth)
}

func (b branch) with(args *combineArgs, v interface{}, depth int, h hasher, matches *int) (out node) {
	defer vet(&out)

	i := newMasker(h.hash())
	return branch{p: b.p.with(i, b.p.get(i).with(args, v, depth+1, h.next(), matches))}
}

func (b branch) without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) (out node) {
	defer vet(&out)

	i := newMasker(h.hash())
	child := b.p.get(i).without(args, v, depth+1, h.next(), matches)
	return branch{p: b.p.with(i, child)}.canonical(depth)
}

func (b branch) allPair(
	o branch,
	mask masker,
	parallel bool,
	op func(m masker, a, b node) bool,
) bool {
	ok := b.p.allPair(o.p, mask, parallel, func(m masker, x, y node) bool {
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
) (out node) {
	defer vet(&out)

	var allMatches [fanout]int
	result := branch{p: b.p.transformPair(o.p, mask, parallel, func(m masker, x, y node) node {
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
) (out node) {
	defer vet(&out)

	var allMatches [fanout]int
	result := branch{p: b.p.transform(parallel, func(m masker, n node) node {
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
	deep := b.countUpTo(20) >= 20
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
