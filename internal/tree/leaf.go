package tree

import (
	"fmt"
	"strings"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/iterator"
)

var theEmptyNode leaf

type leaf []elementT

func (l leaf) Canonical(depth int) node {
	switch {
	case len(l) <= maxLeafLen || depth*fanoutBits >= 64:
		return l
	default:
		var matches int
		return (&branch{}).Combine(DefaultNPCombineArgs, l, depth, &matches)
	}
}

func (l leaf) Combine(args *CombineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case leaf:
		cloned := false
		for i, e := range n {
			if j := l.find(e, args.eq); j >= 0 {
				if !cloned {
					l = l.clone(0)
					cloned = true
				}
				l[j] = args.f(l[j], e)
				*matches++
			} else if len(l) < maxLeafLen {
				l = append(l, e)
			} else {
				return (&branch{}).Combine(args, l, depth, matches).Combine(args, n[i:], depth, matches)
			}
		}
		if len(l) > maxLeafLen {
			panic(errors.WTF)
		}
		return l.Canonical(depth)
	case *branch:
		return n.Combine(args.flip, l, depth, matches)
	default:
		panic(errors.WTF)
	}
}

func (l leaf) CopyTo(dest []elementT) []elementT {
	if len(dest)+len(l) > cap(dest) {
		return nil
	}
	return append(dest, l...)
}

func (l leaf) Defrost() unNode {
	panic(errors.Unimplemented)
}

func (l leaf) Difference(args *EqArgs, n node, depth int, removed *int) node {
	var result leaf
	for _, e := range l {
		if n.Get(args.flip, e, newHasher(e, depth)) == nil {
			result = append(result, e)
		} else {
			*removed++
		}
	}
	return result.Canonical(depth)
}

func (l leaf) Empty() bool {
	return len(l) == 0
}

func (l leaf) Equal(args *EqArgs, n node, depth int) bool {
	if m, is := n.(leaf); is {
		if len(l) != len(m) {
			return false
		}
		for _, e := range l {
			if m.Get(args, e, 0) == nil {
				return false
			}
		}
		return true
	}
	return false
}

func (l leaf) Get(args *EqArgs, v elementT, h hasher) *elementT {
	if i := l.find(v, args.eq); i != -1 {
		return &l[i]
	}
	return nil
}

func (l leaf) Intersection(args *EqArgs, n node, depth int, matches *int) node {
	var result leaf
	for _, e := range l {
		if n.Get(args, e, newHasher(e, depth)) != nil {
			*matches++
			result = append(result, e)
		}
	}
	return result.Canonical(depth)
}

func (l leaf) Iterator([][]node) iterator.Iterator {
	return newLeafIterator(l)
}

func (l leaf) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	return r(l...)
}

func (l leaf) SubsetOf(args *EqArgs, n node, depth int) bool {
	for _, e := range l {
		if n.Get(args, e, 0) == nil {
			return false
		}
	}
	return true
}

func (l leaf) Transform(args *CombineArgs, depth int, counts *int, f func(e elementT) elementT) node {
	var nb Builder
	for _, e := range l {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	*counts += t.count
	return t.root
}

func (l leaf) Where(args *WhereArgs, depth int, matches *int) node {
	var result leaf
	for _, e := range l {
		if args.Pred(e) {
			result = append(result, e)
			*matches++
		}
	}
	return result.Canonical(depth)
}

func (l leaf) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
	if i := l.find(v, args.eq); i >= 0 {
		*matches++
		result := l.clone(0)
		result[i] = args.f(result[i], v)
		return result
	}
	return append(l, v).Canonical(depth)
}

func (l leaf) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
	if i := l.find(v, args.eq); i != -1 {
		*matches++
		result := l[:len(l)-1].clone(0)
		if i != len(result) {
			result[i] = l[len(result)]
		}
		return result.Canonical(depth)
	}
	return l
}

func (l leaf) clone(extra int) leaf {
	return append(make(leaf, 0, len(l)+extra), l...)
}

func (l leaf) find(v elementT, eq func(a, b elementT) bool) int { //nolint:gocritic
	for i, e := range l {
		if eq(e, v) {
			return i
		}
	}
	return -1
}

func (l leaf) String() string {
	var b strings.Builder
	b.WriteByte('(')
	for i, e := range l {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprint(&b, e)
	}
	b.WriteByte(')')
	return b.String()
}
