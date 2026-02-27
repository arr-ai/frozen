package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type leaf2[T any] struct {
	data [2]T
}

func newLeaf2[T any](a, b T) *leaf2[T] {
	return &leaf2[T]{data: [2]T{a, b}}
}

// fmt.Formatter

func (l *leaf2[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "‹")
	fu.Format(l.data[0], f, verb)
	fu.WriteString(f, ",")
	fu.Format(l.data[1], f, verb)
	fu.WriteString(f, "›")
}

// fmt.Stringer

func (l *leaf2[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

func (l *leaf2[T]) Add(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			l.data[i] = args.f(e, v)
			return l, 1
		}
	}
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}}, 0
}

func (l *leaf2[T]) AddFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if value.Equal(e, v) {
			l.data[i] = v
			return l, 1
		}
	}
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}}, 0
}

func (l *leaf2[T]) AppendTo(dest []T) []T {
	if len(dest)+2 > cap(dest) {
		return nil
	}
	return append(dest, l.data[0], l.data[1])
}

func (l *leaf2[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf1[T]:
		for j, f := range l.data {
			if args.eq(f, n.data) {
				ret := *l
				ret.data[j] = args.f(f, n.data)
				return &ret, 1
			}
		}
		return &leaf[T]{data: []T{l.data[0], l.data[1], n.data}}, 0
	case *leaf2[T]:
		merged, m := combineLeafSlices(args, []T{l.data[0], l.data[1]}, n.data[:])
		return leafCanonical(merged), m
	case *leaf[T]:
		return n.Combine(args.Flip(), l, depth)
	default:
		panic("unexpected node type in leaf2.Combine")
	}
}

func (l *leaf2[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	var ret []T
	for _, e := range l.data {
		h := newHasherWith(e, depth, args.hash)
		if n.Get(args, e, h, depth) != nil {
			matches++
		} else {
			ret = append(ret, e)
		}
	}
	return leafCanonical(ret), matches
}

func (l *leaf2[T]) Empty() bool {
	return false
}

func (l *leaf2[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	if n, ok := n.(*leaf2[T]); ok {
		return (args.eq(l.data[0], n.data[0]) && args.eq(l.data[1], n.data[1])) ||
			(args.eq(l.data[0], n.data[1]) && args.eq(l.data[1], n.data[0]))
	}
	return false
}

func (l *leaf2[T]) Get(args *EqArgs[T], v T, _ hasher, _ int) *T {
	if args.eq(l.data[0], v) {
		return &l.data[0]
	}
	if args.eq(l.data[1], v) {
		return &l.data[1]
	}
	return nil
}

func (l *leaf2[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	var ret []T
	for _, e := range l.data {
		h := newHasherWith(e, depth, args.hash)
		if n.Get(args, e, h, depth) != nil {
			ret = append(ret, e)
			matches++
		}
	}
	return leafCanonical(ret), matches
}

func (l *leaf2[T]) Iterator([][]node[T]) Iterator[T] {
	return newSliceIterator(l.data[:])
}

func (l *leaf2[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	a, b := f(l.data[0]), f(l.data[1])
	if args.eq(a, b) {
		return &leaf1[T]{data: args.f(a, b)}, 1
	}
	return newLeaf2(a, b), 2
}

func (l *leaf2[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data[0], l.data[1])
}

func (l *leaf2[T]) Remove(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return &leaf1[T]{data: l.data[1]}, 1
	}
	if args.eq(l.data[1], v) {
		return &leaf1[T]{data: l.data[0]}, 1
	}
	return l, 0
}

func (l *leaf2[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	for _, e := range l.data {
		h := newHasherWith(e, depth, args.hash)
		if n.Get(args, e, h, depth) == nil {
			return false
		}
	}
	return true
}

func (l *leaf2[T]) Vet() int {
	return 2
}

func (l *leaf2[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	m0 := args.Pred(l.data[0])
	m1 := args.Pred(l.data[1])
	switch {
	case m0 && m1:
		return l, 2
	case m0:
		return &leaf1[T]{data: l.data[0]}, 1
	case m1:
		return &leaf1[T]{data: l.data[1]}, 1
	default:
		return nil, 0
	}
}

func (l *leaf2[T]) With(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return newLeaf2(args.f(l.data[0], v), l.data[1]), 1
	}
	if args.eq(l.data[1], v) {
		return newLeaf2(l.data[0], args.f(l.data[1], v)), 1
	}
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}}, 0
}

func (l *leaf2[T]) WithFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data[0], v) {
		return newLeaf2(v, l.data[1]), 1
	}
	if value.Equal(l.data[1], v) {
		return newLeaf2(l.data[0], v), 1
	}
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}}, 0
}

func (l *leaf2[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return &leaf1[T]{data: l.data[1]}, 1
	}
	if args.eq(l.data[1], v) {
		return &leaf1[T]{data: l.data[0]}, 1
	}
	return l, 0
}

func (l *leaf2[T]) clone() node[T] {
	ret := *l
	return &ret
}
