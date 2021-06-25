package tree

import (
	"fmt"
	"strings"

	"github.com/arr-ai/frozen/errors"
)

var theEmptyNode = newLeaf()

type leaf struct {
	data []elementT
}

func newLeaf(data ...elementT) *leaf {
	return &leaf{data: data}
}

func (l *leaf) Canonical(depth int) node {
	if len(l.data) <= maxLeafLen || depth*fanoutBits >= 64 {
		return l
	}
	var matches int
	return (&branch{}).Combine(DefaultNPCombineArgs, l, depth, &matches)
}

func (l *leaf) Combine(args *CombineArgs, n node, depth int, matches *int) node { //nolint:cyclop
	switch n := n.(type) {
	case *leaf:
		if l.Empty() {
			return n
		}
		cloned := false
	scanning:
		for i, e := range n.data {
			for j, f := range l.data {
				if args.eq(f, e) {
					if !cloned {
						l = l.clone(0)
						cloned = true
					}
					l.data[j] = args.f(f, e)
					*matches++
					continue scanning
				}
			}
			if len(l.data) < maxLeafLen {
				l = newLeaf(append(l.data, e)...)
			} else {
				return (&branch{}).Combine(args, l, depth, matches).Combine(args, newLeaf(n.data[i:]...), depth, matches)
			}
		}
		if len(l.data) > maxLeafLen {
			panic(errors.WTF)
		}
		return l.Canonical(depth)
	case *branch:
		return n.Combine(args.flip, l, depth, matches)
	default:
		panic(errors.WTF)
	}
}

func (l *leaf) CopyTo(dest []elementT) []elementT {
	if len(dest)+len(l.data) > cap(dest) {
		return nil
	}
	return append(dest, l.data...)
}

func (l *leaf) Defrost() unNode {
	panic(errors.Unimplemented)
}

func (l *leaf) Difference(args *EqArgs, n node, depth int, removed *int) node {
	var result leaf
	for _, e := range l.data {
		if n.Get(args.flip, e, newHasher(e, depth)) == nil {
			result.data = append(result.data, e)
		} else {
			*removed++
		}
	}
	return result.Canonical(depth)
}

func (l *leaf) Empty() bool {
	return len(l.data) == 0
}

func (l *leaf) Equal(args *EqArgs, n node, depth int) bool {
	if m, is := n.(*leaf); is {
		if len(l.data) != len(m.data) {
			return false
		}
		for _, e := range l.data {
			if m.Get(args, e, 0) == nil {
				return false
			}
		}
		return true
	}
	return false
}

func (l *leaf) Get(args *EqArgs, v elementT, h hasher) *elementT {
	for i, e := range l.data {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *leaf) Intersection(args *EqArgs, n node, depth int, matches *int) node {
	var result leaf
	for _, e := range l.data {
		if n.Get(args, e, newHasher(e, depth)) != nil {
			*matches++
			result.data = append(result.data, e)
		}
	}
	return result.Canonical(depth)
}

func (l *leaf) Iterator([][]node) Iterator {
	return newSliceIterator(l.data)
}

func (l *leaf) Reduce(_ NodeArgs, _ int, r func(values ...elementT) elementT) elementT {
	return r(l.data...)
}

func (l *leaf) SubsetOf(args *EqArgs, n node, _ int) bool {
	for _, e := range l.data {
		if n.Get(args, e, 0) == nil {
			return false
		}
	}
	return true
}

func (l *leaf) Transform(args *CombineArgs, _ int, counts *int, f func(e elementT) elementT) node {
	var nb Builder
	for _, e := range l.data {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	*counts += t.count
	return t.root
}

func (l *leaf) Where(args *WhereArgs, depth int, matches *int) node {
	var result leaf
	for _, e := range l.data {
		if args.Pred(e) {
			result.data = append(result.data, e)
			*matches++
		}
	}
	return result.Canonical(depth)
}

func (l *leaf) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
	for i, e := range l.data {
		if args.eq(e, v) {
			*matches++
			ret := l.clone(0)
			ret.data[i] = args.f(ret.data[i], v)
			return ret
		}
	}
	return newLeaf(append(l.data, v)...).Canonical(depth)
}

func (l *leaf) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
	for i, e := range l.data {
		if args.eq(e, v) {
			*matches++
			ret := newLeaf(l.data[:len(l.data)-1]...).clone(0)
			if i != len(ret.data) {
				ret.data[i] = l.data[len(ret.data)]
			}
			return ret.Canonical(depth)
		}
	}
	return l
}

func (l *leaf) clone(extra int) *leaf {
	return newLeaf(append(make([]elementT, 0, len(l.data)+extra), l.data...)...)
}

func (l *leaf) String() string {
	var b strings.Builder
	b.WriteByte('(')
	for i, e := range l.data {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprint(&b, e)
	}
	b.WriteByte(')')
	return b.String()
}
