package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type leaf1[T any] struct {
	data T
	h0   H128
}

func newLeaf1[T any](v T) *leaf1[T] {
	return &leaf1[T]{data: v, h0: newElemH128(v, getHashFunc[T]())}
}

func newLeaf1WithHash[T any](v T, h0 H128) *leaf1[T] {
	return &leaf1[T]{data: v, h0: h0}
}

// fmt.Formatter

func (l *leaf1[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "‹")
	fu.Format(l.data, f, verb)
	fu.WriteString(f, "›")
}

// fmt.Stringer

func (l *leaf1[T]) String() string {
	return fmt.Sprintf("%s", l)
}

func (l *leaf1[T]) H0() H128 { return l.h0 }

// node[T]

func (l *leaf1[T]) Add(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		l.data = args.f(l.data, v)
		// h0 left stale — Builder.Finish() recomputes.
		return l, 1
	}
	return &leaf2[T]{data: [2]T{l.data, v}}, 0
}

func (l *leaf1[T]) AddFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		l.data = v
		// h0 left stale — Builder.Finish() recomputes.
		return l, 1
	}
	return &leaf2[T]{data: [2]T{l.data, v}}, 0
}

func (l *leaf1[T]) AppendTo(dest []T) []T {
	if len(dest)+1 > cap(dest) {
		return nil
	}
	return append(dest, l.data)
}

func (l *leaf1[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf1[T]:
		if args.eq(l.data, n.data) {
			combined := args.f(l.data, n.data)
			return newLeaf1WithHash(combined, newElemH128(combined, args.hash)), 1
		}
		return newLeaf2WithHash(l.data, n.data, l.h0, n.h0), 0
	case *leaf2[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf[T]:
		return n.Combine(args.Flip(), l, depth)
	default:
		panic("unexpected node type in leaf1.Combine")
	}
}

func (l *leaf1[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	h := hasherFromCached(l.h0, depth)
	if n.Get(args, l.data, h, depth) != nil {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) Empty() bool {
	return false
}

func (l *leaf1[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	if n, ok := n.(*leaf1[T]); ok {
		if l.h0 != n.h0 {
			return false
		}
		if args.fullHash && !l.h0.isZero() {
			return true
		}
		return args.eq(l.data, n.data)
	}
	return false
}

func (l *leaf1[T]) Get(args *EqArgs[T], v T, _ hasher, _ int) *T {
	if args.eq(l.data, v) {
		return &l.data
	}
	return nil
}

func (l *leaf1[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	h := hasherFromCached(l.h0, depth)
	if n.Get(args, l.data, h, depth) != nil {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) Iterator([][]node[T]) Iterator[T] {
	return newSliceIterator([]T{l.data})
}

func (l *leaf1[T]) Map(_ *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	return newLeaf1(f(l.data)), 1
}

func (l *leaf1[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data)
}

func (l *leaf1[T]) Remove(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	h := hasherFromCached(l.h0, depth)
	return n.Get(args, l.data, h, depth) != nil
}

func (l *leaf1[T]) Vet() int {
	return 1
}

func (l *leaf1[T]) vetH0(hf func(T) H128) {
	if want := newElemH128(l.data, hf); l.h0 != want {
		panic(fmt.Errorf("leaf1 h0 mismatch: stored %v, computed %v", l.h0, want))
	}
}

func (l *leaf1[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	if args.Pred(l.data) {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) With(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		combined := args.f(l.data, v)
		return newLeaf1WithHash(combined, newElemH128(combined, args.hash)), 1
	}
	return newLeaf2WithHash(l.data, v, l.h0, newElemH128(v, args.hash)), 0
}

func (l *leaf1[T]) WithFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		// Equal values have equal hashes, so reuse l.h0.
		return newLeaf1WithHash(v, l.h0), 1
	}
	hf := getHashFunc[T]()
	return newLeaf2WithHash(l.data, v, l.h0, newElemH128(v, hf)), 0
}

func (l *leaf1[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) clone() node[T] {
	ret := *l
	return &ret
}
