package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/errors"
	"github.com/arr-ai/frozen/v2/internal/fu"
	"github.com/arr-ai/frozen/v2/internal/iterator"
)

type leaf[T any] struct {
	data [2]T
}

func newLeaf[T any](data ...T) *leaf[T] {
	switch len(data) {
	case 1:
		return newLeaf1(data[0])
	case 2:
		return newLeaf2(data[0], data[1])
	default:
		panic(errors.Errorf("data wrong size (%d) for leaf[T]", len(data)))
	}
}

func newLeaf1[T any](a T) *leaf[T] {
	var zero T
	return newLeaf2(a, zero)
}

func newLeaf2[T any](a, b T) *leaf[T] {
	if isZero(a) {
		panic(errors.WTF)
	}
	return &leaf[T]{data: [2]T{a, b}}
}

// fmt.Formatter

func (l *leaf[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	if !isZero(l.data[0]) {
		fu.Format(l.data[0], f, verb)
		if !isZero(l.data[1]) {
			fu.WriteString(f, ",")
			fu.Format(l.data[1], f, verb)
		}
	}
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *leaf[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

func (l *leaf[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	switch {
	case args.eq(l.data[0], v):
		l.data[0] = args.f(l.data[0], v)
		matches++
	case isZero(l.data[1]):
		l.data[1] = v
	case args.eq(l.data[1], v):
		l.data[1] = args.f(l.data[1], v)
		matches++
	case depth >= maxTreeDepth:
		return newTwig(l.data[0], l.data[1], v), 0
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), 0
	}
	return l, matches
}

func (l *leaf[T]) Canonical(depth int) node[T] {
	return l
}

func (l *leaf[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) { //nolint:cyclop
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf[T]:
		lr := func(a, b int) int { return a<<2 | b }
		masks := lr(l.mask(), n.mask())
		if masks == lr(3, 1) {
			masks, l, n, args = lr(1, 3), n, l, args.Flip()
		}
		l0, l1 := l.data[0], l.data[1]
		n0, n1 := n.data[0], n.data[1]
		if args.eq(l0, n0) { //nolint:nestif
			r0 := args.f(l0, n0)
			matches++
			switch masks {
			case lr(1, 1):
				return newLeaf1(r0), matches
			case lr(1, 3):
				return newLeaf2(r0, n1), matches
			default:
				if args.eq(l1, n1) {
					matches++
					return newLeaf2(r0, args.f(l1, n1)), matches
				}
				return newBranchFrom(depth, r0, l1, n1), matches
			}
		} else {
			switch masks {
			case lr(1, 1):
				return newLeaf2(l0, n0), matches
			case lr(1, 3):
				if args.eq(l0, n1) {
					matches++
					return newLeaf2(n0, args.f(l0, n1)), matches
				}
				return newBranchFrom(depth, l0, n0, n1), matches
			default:
				if args.eq(l1, n1) {
					matches++
					return newBranchFrom(depth, l0, n0, args.f(l1, n1)), matches
				}
				if args.eq(l0, n1) {
					r0 := args.f(l0, n1)
					matches++
					if args.eq(l1, n0) {
						matches++
						return newLeaf2(r0, args.f(l1, n0)), matches
					}
					return newBranchFrom(depth, r0, l1, n0), matches
				}
				if args.eq(l1, n0) {
					matches++
					return newBranchFrom(depth, l0, n1, args.f(l1, n0)), matches
				}
				return newBranchFrom(depth, l0, l1, n0, n1), matches
			}
		}
	default:
		panic(errors.WTF)
	}
}

func (l *leaf[T]) AppendTo(dest []T) []T {
	data := l.slice()
	if len(dest)+len(data) > cap(dest) {
		return nil
	}
	return append(dest, data...)
}

func (l *leaf[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	mask := l.mask()
	if n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil {
		matches++
		mask &^= 0b01
	}

	if !isZero(l.data[1]) {
		if n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil {
			matches++
			mask &^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf[T]) Empty() bool {
	return false
}

func (l *leaf[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	if n, is := n.(*leaf[T]); is {
		lm, nm := l.mask(), n.mask()
		if lm != nm {
			return false
		}
		l0, l1 := l.data[0], l.data[1]
		n0, n1 := n.data[0], n.data[1]
		if lm == 1 && nm == 1 {
			return args.eq(l0, n0)
		}
		return args.eq(l0, n0) && args.eq(l1, n1) ||
			args.eq(l0, n1) && args.eq(l1, n0)
	}
	return false
}

func (l *leaf[T]) Get(args *EqArgs[T], v T, _ hasher) *T {
	for i, e := range l.slice() {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *leaf[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	mask := 0
	if n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil {
		matches++
		mask |= 0b01
	}

	if !isZero(l.data[1]) {
		if n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil {
			matches++
			mask |= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf[T]) Iterator([][]node[T]) iterator.Iterator[T] {
	return newSliceIterator(l.slice())
}

func (l *leaf[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.slice()...)
}

func (l *leaf[T]) Remove(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	var zero T
	if args.eq(l.data[0], v) {
		matches++
		if isZero(l.data[1]) {
			return nil, matches
		}
		l.data = [2]T{l.data[1], zero}
	} else if !isZero(l.data[1]) {
		if args.eq(l.data[1], v) {
			matches++
			l.data[1] = zero
		}
	}
	return l, matches
}

func (l *leaf[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	a := l.data[0]
	h := newHasher(a, depth)
	if n.Get(args, a, h) == nil {
		return false
	}
	if b := l.data[1]; !isZero(b) {
		h := newHasher(b, depth)
		if n.Get(args, b, h) == nil {
			return false
		}
	}
	return true
}

func (l *leaf[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	var nb Builder[T]
	for _, e := range l.slice() {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	matches += t.count
	return t.root, matches
}

func (l *leaf[T]) Vet() int {
	if isZero(l.data[0]) {
		if !isZero(l.data[1]) {
			panic(errors.Errorf("data only in leaf[T] slot 1"))
		}
		panic(errors.Errorf("empty leaf[T]"))
	}
	return l.count()
}

func (l *leaf[T]) Where(args *WhereArgs[T], depth int) (_ node[T], matches int) {
	var mask int
	if args.Pred(l.data[0]) {
		matches++
		mask ^= 0b01
	}
	if !isZero(l.data[1]) {
		if args.Pred(l.data[1]) {
			matches++
			mask ^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf[T]) where(mask int) node[T] {
	switch mask {
	case 0b00:
		return nil
	case 0b01:
		if isZero(l.data[1]) {
			return l
		}
		return newLeaf1(l.data[0])
	case 0b10:
		return newLeaf1(l.data[1])
	default:
		return l
	}
}

func (l *leaf[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	switch {
	case args.eq(l.data[0], v):
		matches++
		return newLeaf2(args.f(l.data[0], v), l.data[1]), matches
	case isZero(l.data[1]):
		return newLeaf2(l.data[0], v), 0
	case args.eq(l.data[1], v):
		matches++
		return newLeaf2(l.data[0], args.f(l.data[1], v)), matches
	case depth >= maxTreeDepth:
		return newTwig(append(l.data[:], v)...), matches
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), matches
	}
}

func (l *leaf[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	mask := l.mask()
	if args.eq(l.data[0], v) {
		matches++
		mask ^= 0b01
	} else if !isZero(l.data[1]) {
		if args.eq(l.data[1], v) {
			matches++
			mask ^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf[T]) count() int {
	if isZero(l.data[1]) {
		return 1
	}
	return 2
}

func (l *leaf[T]) clone() node[T] {
	ret := *l
	return &ret
}

func (l *leaf[T]) mask() int {
	if isZero(l.data[1]) {
		return 1
	}
	return 3
}

func (l *leaf[T]) slice() []T {
	return l.data[:l.count()]
}
