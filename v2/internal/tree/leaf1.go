package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/errors"
	"github.com/arr-ai/frozen/v2/internal/fu"
	"github.com/arr-ai/frozen/v2/internal/iterator"
)

type leaf1[T comparable] struct {
	data T
}

func newLeaf[T comparable](data ...T) node[T] {
	switch len(data) {
	case 1:
		return newLeaf1(data[0])
	case 2:
		return newLeaf2(data[0], data[1])
	default:
		panic(errors.Errorf("data wrong size (%d) for leaf", len(data)))
	}
}

func newLeaf1[T comparable](a T) *leaf1[T] {
	return &leaf1[T]{data: a}
}

// fmt.Formatter

func (l *leaf1[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	fu.Format(l.data, f, verb)
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *leaf1[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

func (l *leaf1[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		l.data = args.f(l.data, v)
		return l, 1
	}
	return newLeaf2(l.data, v), 0
}

func (l *leaf1[T]) Canonical(depth int) node[T] {
	return l
}

func (l *leaf1[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) { //nolint:cyclop
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf2[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf1[T]:
		if args.eq(l.data, n.data) { //nolint:nestif
			return newLeaf1(args.f(l.data, n.data)), 1
		}
		return newLeaf2(l.data, n.data), 0
	default:
		panic(errors.WTF)
	}
}

func (l *leaf1[T]) AppendTo(dest []T) []T {
	if len(dest)+1 > cap(dest) {
		return nil
	}
	return append(dest, l.data)
}

func (l *leaf1[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	if n.Get(args, l.data, newHasher(l.data, depth)) != nil {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) Empty() bool {
	return false
}

func (l *leaf1[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	l2, is := n.(*leaf1[T])
	return is && args.eq(l.data, l2.data)
}

func (l *leaf1[T]) Get(args *EqArgs[T], v T, _ hasher) *T {
	if args.eq(l.data, v) {
		return &l.data
	}
	return nil
}

func (l *leaf1[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	if n.Get(args, l.data, newHasher(l.data, depth)) != nil {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) Iterator([][]node[T]) iterator.Iterator[T] {
	// TODO: Avoid malloc.
	return newSliceIterator([]T{l.data})
}

func (l *leaf1[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data)
}

func (l *leaf1[T]) Remove(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	return n.Get(args, l.data, newHasher(l.data, depth)) != nil
}

func (l *leaf1[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	return newLeaf1(f(l.data)), 1
}

func (l *leaf1[T]) Vet() int {
	return 1
}

func (l *leaf1[T]) Where(args *WhereArgs[T], depth int) (_ node[T], matches int) {
	if args.Pred(l.data) {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return newLeaf1(args.f(l.data, v)), 1
	}
	return newLeaf2(l.data, v), 0
}

func (l *leaf1[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) clone() node[T] {
	ret := *l
	return &ret
}
