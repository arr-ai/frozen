package frozen

import (
	"fmt"
	"strings"
)

const maxLeafLen = 8

type leaf []interface{}

func (l leaf) vet() node {
	// if len(l) == 0 || len(l) > maxLeafLen {
	// 	panic(wtf)
	// }
	return l
}

func (l leaf) canonical(depth int) (out node) {
	defer vet(&out)
	switch {
	case len(l) == 0:
		return emptyNode{}
	case len(l) <= maxLeafLen || depth*fanoutBits >= 64:
		return l
	default:
		var matches int
		return (branch{}).combine(defaultNPCombineArgs, l, depth, &matches)
	}
}

func (l leaf) combine(args *combineArgs, n node, depth int, matches *int) (out node) {
	l.vet()
	defer vet(&out)

	switch n := n.(type) {
	case emptyNode:
		return l.vet()
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
				return (branch{}).combine(args, l, depth, matches).combine(args, n[i:], depth, matches)
			}
		}
		if len(l) > maxLeafLen {
			panic(wtf)
		}
		return l.canonical(depth)
	case branch:
		return n.combine(args.flip, l, depth, matches)
	default:
		panic(wtf)
	}
}

func (l leaf) countUpTo(int) int {
	l.vet()
	return len(l)
}

func (l leaf) difference(args *eqArgs, n node, depth int, removed *int) (out node) {
	l.vet()
	defer vet(&out)

	var result leaf
	for _, e := range l {
		if n.get(args.flip, e, newHasher(e, depth)) == nil {
			result = append(result, e)
		} else {
			*removed++
		}
	}
	return result.canonical(depth).vet()
}

func (l leaf) equal(args *eqArgs, n node, depth int) bool {
	l.vet()

	if m, is := n.(leaf); is {
		if len(l) != len(m) {
			return false
		}
		for _, e := range l {
			if m.get(args, e, 0) == nil {
				return false
			}
		}
		return true
	}
	return false
}

func (l leaf) get(args *eqArgs, v interface{}, h hasher) *interface{} {
	l.vet()
	if i := l.find(v, args.eq); i != -1 {
		return &l[i]
	}
	return nil
}

func (l leaf) intersection(args *eqArgs, n node, depth int, matches *int) (out node) {
	l.vet()
	defer vet(&out)

	var result leaf
	for _, e := range l {
		if n.get(args, e, newHasher(e, depth)) != nil {
			*matches++
			result = append(result, e)
		}
	}
	return result.canonical(depth)
}

func (l leaf) isSubsetOf(args *eqArgs, n node, depth int) bool {
	l.vet()

	for _, e := range l {
		if n.get(args, e, 0) == nil {
			return false
		}
	}
	return true
}

func (l leaf) iterator([]packed) Iterator {
	l.vet()
	return newLeafIterator(l)
}

func (l leaf) reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{} {
	return r(l...)
}

func (l leaf) transform(args *combineArgs, depth int, counts *int, f func(v interface{}) interface{}) node {
	var nb nodeBuilder
	for _, e := range l {
		nb.Add(args, f(e))
	}
	root := nb.Finish()
	*counts = root.count
	return root.n
}

func (l leaf) where(args *whereArgs, depth int, matches *int) (out node) {
	l.vet()
	defer vet(&out)

	var result leaf
	for _, e := range l {
		if args.pred(e) {
			result = append(result, e)
			*matches++
		}
	}
	return result.canonical(depth)
}

func (l leaf) with(args *combineArgs, v interface{}, depth int, h hasher, matches *int) (out node) {
	l.vet()
	defer vet(&out)

	if i := l.find(v, args.eq); i >= 0 {
		*matches++
		result := l.clone(0)
		result[i] = args.f(result[i], v)
		return result
	}
	return append(l, v).canonical(depth)
}

func (l leaf) without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) (out node) {
	l.vet()
	defer vet(&out)

	if i := l.find(v, args.eq); i != -1 {
		*matches++
		result := l[:len(l)-1].clone(0)
		if i != len(result) {
			result[i] = l[len(result)]
		}
		return result.canonical(depth)
	}
	return l
}

func (l leaf) clone(extra int) leaf {
	return append(make(leaf, 0, len(l)+extra), l...)
}

func (l leaf) find(v interface{}, eq func(a, b interface{}) bool) int { //nolint:gocritic
	l.vet()
	for i, e := range l {
		if eq(e, v) {
			return i
		}
	}
	return -1
}

func (l leaf) String() string {
	l.vet()
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
