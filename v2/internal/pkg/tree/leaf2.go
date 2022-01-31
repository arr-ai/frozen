package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/pkg/errors"
	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
)

type leaf2[T comparable] struct {
	data [2]T
}

func newLeaf2[T comparable](a, b T) *leaf2[T] {
	return &leaf2[T]{data: [2]T{a, b}}
}

// fmt.Formatter

func (l *leaf2[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	fu.Format(l.data[0], f, verb)
	fu.WriteString(f, ",")
	fu.Format(l.data[1], f, verb)
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *leaf2[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

func (l *leaf2[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	switch {
	case args.eq(l.data[0], v):
		l.data[0] = args.f(l.data[0], v)
		return l, 1
	case args.eq(l.data[1], v):
		l.data[1] = args.f(l.data[1], v)
		return l, 1
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), 0
	}
}

func (l *leaf2[T]) Canonical(depth int) node[T] {
	return l
}

func (l *leaf2[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) { //nolint:cyclop
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf2[T]:
		l0, l1 := l.data[0], l.data[1]
		n0, n1 := n.data[0], n.data[1]
		for i := 0; i < 2; i++ {
			switch {
			case args.eq(l0, n0):
				r0 := args.f(l0, n0)
				matches++
				if args.eq(l1, n1) {
					matches++
					return newLeaf2(r0, args.f(l1, n1)), matches
				}
				return newBranchFrom(depth, r0, l1, n1), matches
			case args.eq(l1, n1):
				matches++
				return newBranchFrom(depth, l0, n0, args.f(l1, n1)), matches
			}
			n0, n1 = n1, n0
		}
		return newBranchFrom(depth, l0, l1, n0, n1), matches
	case *leaf1[T]:
		l0, l1 := l.data[0], l.data[1]
		nd := n.data
		switch {
		case args.eq(l0, nd):
			return newLeaf2(args.f(l0, nd), l1), 1
		case args.eq(l1, nd):
			return newLeaf2(l0, args.f(l1, nd)), 1
		default:
			return newBranchFrom(depth, l0, l1, nd), 0
		}
	default:
		panic(errors.WTF)
	}
}

func (l *leaf2[T]) AppendTo(dest []T) []T {
	if len(dest)+len(l.data) > cap(dest) {
		return nil
	}
	return append(dest, l.data[:]...)
}

func (l *leaf2[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	a := n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil
	b := n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil
	switch {
	case a && b:
		return nil, 2
	case a:
		return newLeaf1(l.data[1]), 1
	case b:
		return newLeaf1(l.data[0]), 1
	default:
		return l, 0
	}
}

func (l *leaf2[T]) Empty() bool {
	return false
}

func (l *leaf2[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	l2, is := n.(*leaf2[T])
	return is && (
		args.eq(l.data[0], l2.data[0]) && args.eq(l.data[1], l2.data[1]) ||
		args.eq(l.data[0], l2.data[1]) && args.eq(l.data[1], l2.data[0]))
}

func (l *leaf2[T]) Get(args *EqArgs[T], v T, _ hasher) *T {
	for i, e := range l.data {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *leaf2[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	g0 := n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil
	g1 := n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil
	switch {
	case g0 && g1:
		return l, 2
	case g0:
		return newLeaf1(l.data[0]), 1
	case g1:
		return newLeaf1(l.data[1]), 1
	default:
		return nil, 0
	}
}

func (l *leaf2[T]) Iterator([][]node[T]) iterator.Iterator[T] {
	return newSliceIterator(l.data[:])
}

func (l *leaf2[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data[:]...)
}

func (l *leaf2[T]) Remove(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	switch {
	case args.eq(l.data[0], v):
		return newLeaf1(l.data[1]), 1
	case args.eq(l.data[1], v):
		return newLeaf1(l.data[0]), 1
	default:
		return l, 0
	}
}

func (l *leaf2[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	return n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil &&
		n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil
}

func (l *leaf2[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	a, b := f(l.data[0]), f(l.data[1])
	if args.eq(a, b) {
		return newLeaf1(a), 1
	}
	return newLeaf2(a, b), 2
}

func (l *leaf2[T]) Vet() int {
	return 2
}

func (l *leaf2[T]) Where(args *WhereArgs[T], depth int) (_ node[T], matches int) {
	p0 := args.Pred(l.data[0])
	p1 := args.Pred(l.data[1])
	switch {
	case p0 && p1:
		return l, 2
	case p0:
		return newLeaf1(l.data[0]), 1
	case p1:
		return newLeaf1(l.data[1]), 1
	default:
		return nil, 0
	}
}

func (l *leaf2[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	switch {
	case args.eq(l.data[0], v):
		return newLeaf2(args.f(l.data[0], v), l.data[1]), 1
	case args.eq(l.data[1], v):
		return newLeaf2(l.data[0], args.f(l.data[1], v)), 1
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), 0
	}
}

func (l *leaf2[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return newLeaf1(l.data[1]), 1
	}
	if args.eq(l.data[1], v) {
		return newLeaf1(l.data[0]), 1
	}
	return l, 0
}

func (l *leaf2[T]) clone() node[T] {
	ret := *l
	return &ret
}
