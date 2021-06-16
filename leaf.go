package frozen

import (
	"fmt"
	"strings"
)

const maxLeafLen = 8

type leaf []interface{}

func (l leaf) Canonical(depth int) node {
	switch {
	case len(l) == 0:
		return emptyNode{}
	case len(l) <= maxLeafLen || depth*fanoutBits >= 64:
		return l
	default:
		var matches int
		return (branch{}).Combine(defaultNPCombineArgs, l, depth, &matches)
	}
}

func (l leaf) Combine(args *combineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case emptyNode:
		return l
	case leaf:
		cloned := false
		for i, e := range n {
			if j := l.find(e, args.eq); j >= 0 {
				if !cloned {
					l = l.clone(0)
					cloned = true
				}
				l[j] = args.f(l[j], e)
			} else if len(l) < maxLeafLen {
				l = append(l, e)
			} else {
				return (branch{}).Combine(args, l, depth, matches).Combine(args, n[i:], depth, matches)
			}
		}
		if len(l) > maxLeafLen {
			panic(WTF)
		}
		return l.Canonical(depth)
	case branch:
		return n.Combine(args.flip, l, depth, matches)
	default:
		panic(WTF)
	}
}

func (l leaf) CountUpTo(int) int {
	return len(l)
}

func (l leaf) Difference(args *eqArgs, n node, depth int, removed *int) node {
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

func (l leaf) Equal(args *eqArgs, n node, depth int) bool {
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

func (l leaf) Get(args *eqArgs, v interface{}, h hasher) *interface{} {
	if i := l.find(v, args.eq); i != -1 {
		return &l[i]
	}
	return nil
}

func (l leaf) Intersection(args *eqArgs, n node, depth int, matches *int) node {
	var result leaf
	for _, e := range l {
		if n.Get(args, e, newHasher(e, depth)) != nil {
			*matches++
			result = append(result, e)
		}
	}
	return result.Canonical(depth)
}

func (l leaf) Iterator([]packed) Iterator {
	return newLeafIterator(l)
}

func (l leaf) Reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{} {
	return r(l...)
}

func (l leaf) SubsetOf(args *eqArgs, n node, depth int) bool {
	for _, e := range l {
		if n.Get(args, e, 0) == nil {
			return false
		}
	}
	return true
}

func (l leaf) Transform(args *combineArgs, depth int, counts *int, f func(v interface{}) interface{}) node {
	var nb nodeBuilder
	for _, e := range l {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	*counts = t.count
	return t.root
}

func (l leaf) Where(args *whereArgs, depth int, matches *int) node {
	var result leaf
	for _, e := range l {
		if args.pred(e) {
			result = append(result, e)
			*matches++
		}
	}
	return result.Canonical(depth)
}

func (l leaf) With(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node {
	if i := l.find(v, args.eq); i >= 0 {
		*matches++
		result := l.clone(0)
		result[i] = args.f(result[i], v)
		return result
	}
	return append(l, v).Canonical(depth)
}

func (l leaf) Without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) node {
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

func (l leaf) find(v interface{}, eq func(a, b interface{}) bool) int { //nolint:gocritic
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
