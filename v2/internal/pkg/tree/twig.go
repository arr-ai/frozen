package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/pkg/errors"
	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
)

type twig[T comparable] struct {
	data []T
}

func newTwig[T comparable](data ...T) *twig[T] {
	return &twig[T]{data: data}
}

// fmt.Formatter

func (l *twig[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	for i, e := range l.data {
		if i > 0 {
			fu.WriteString(f, ",")
		}
		fu.Format(e, f, verb)
	}
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *twig[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

func (l *twig[T]) Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			l.data[i] = args.f(e, v)
			return l, matches
		}
	}
	if len(l.data) < cap(l.data) || depth >= maxTreeDepth {
		l.data = append(l.data, v)
		return l, matches
	}

	b := &branch[T]{}
	for _, e := range l.data {
		_, m := b.Add(args, e, depth, newHasher(e, depth))
		matches += m
	}
	_, m := b.Add(args, v, depth, h)
	matches += m

	return b, matches
}

func (l *twig[T]) Canonical(depth int) node[T] {
	switch n := len(l.data); n {
	case 0:
		return nil
	case 1:
		return newLeaf1(l.data[0])
	case 2:
		return newLeaf2(l.data[0], l.data[1])
	default:
		if n <= maxLeafLen || depth*fanoutBits >= 64 {
			return l
		}
		return newBranchFrom(depth, l.data...)
	}
}

func (l *twig[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) { //nolint:cyclop
	var ndata []T
	switch n := n.(type) {
	case *twig[T]:
		ndata = n.data
	default:
		panic(errors.WTF)
	}

	cloned := false
scanning:
	for i, e := range ndata {
		for j, f := range l.data {
			if args.eq(f, e) {
				if !cloned {
					l = l.cloneWithExtra(0)
					cloned = true
				}
				l.data[j] = args.f(f, e)
				matches++
				continue scanning
			}
		}
		if len(l.data) < maxLeafLen {
			l = newTwig(append(l.data, e)...)
		} else {
			b := &branch[T]{}
			for _, e := range l.data {
				_, m := b.Add(args, e, depth, newHasher(e, depth))
				matches += m
			}
			for _, e := range ndata[i:] {
				_, m := b.Add(args, e, depth, newHasher(e, depth))
				matches += m
			}
		}
	}
	if len(l.data) > maxLeafLen {
		panic(errors.WTF)
	}
	return l.Canonical(depth), matches
}

func (l *twig[T]) AppendTo(dest []T) []T {
	if len(dest)+len(l.data) > cap(dest) {
		return nil
	}
	return append(dest, l.data...)
}

func (l *twig[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	ret := newTwig[T]()
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args.Flip(), e, h) == nil {
			ret.data = append(ret.data, e)
		} else {
			matches++
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig[T]) Empty() bool {
	return len(l.data) == 0
}

func (l *twig[T]) Equal(args *EqArgs[T], n node[T], depth int) bool {
	if n, is := n.(*twig[T]); is {
		if len(l.data) != len(n.data) {
			return false
		}
		for _, e := range l.data {
			if n.Get(args, e, 0) == nil {
				return false
			}
		}
		return true
	}
	return false
}

func (l *twig[T]) Get(args *EqArgs[T], v T, _ hasher) *T {
	for i, e := range l.data {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *twig[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	ret := newTwig[T]()
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args, e, h) != nil {
			matches++
			ret.data = append(ret.data, e)
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig[T]) Iterator([][]node[T]) iterator.Iterator[T] {
	return newSliceIterator(l.data)
}

func (l *twig[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data...)
}

func (l *twig[T]) Remove(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			last := len(l.data) - 1
			if last == 0 {
				return nil, matches
			}
			if i < last {
				l.data[i] = l.data[last]
			}
			l.data = l.data[:last]
			return l.Canonical(depth), matches
		}
	}
	return l, matches
}

func (l *twig[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args, e, h) == nil {
			return false
		}
	}
	return true
}

func (l *twig[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	var nb Builder[T]
	for _, e := range l.data {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	return t.root, t.count
}

func (l *twig[T]) Vet() int {
	if len(l.data) <= 2 {
		panic(errors.Errorf("twig[T] too small (%d)", len(l.data)))
	}
	h0 := newHasher(l.data[0], 0)
	for _, e := range l.data[1:] {
		if newHasher(e, 0) != h0 {
			panic(errors.Errorf("twig[T] with multiple hashes"))
		}
	}
	return len(l.data)
}

func (l *twig[T]) Where(args *WhereArgs[T], depth int) (_ node[T], matches int) {
	ret := newTwig[T]()
	for _, e := range l.data {
		if args.Pred(e) {
			ret.data = append(ret.data, e)
			matches++
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig[T]) With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			ret := l.cloneWithExtra(0)
			ret.data[i] = args.f(ret.data[i], v)
			return ret, matches
		}
	}
	return newTwig(append(l.data, v)...).Canonical(depth), matches
}

func (l *twig[T]) Without(args *EqArgs[T], v T, depth int, h hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			ret := newTwig(l.data[:len(l.data)-1]...).cloneWithExtra(0)
			if i != len(ret.data) {
				ret.data[i] = l.data[len(ret.data)]
			}
			return ret.Canonical(depth), matches
		}
	}
	return l, matches
}

func (l *twig[T]) clone() node[T] {
	return &twig[T]{data: append(([]T)(nil), l.data...)}
}

func (l *twig[T]) cloneWithExtra(extra int) *twig[T] {
	return newTwig(append(make([]T, 0, len(l.data)+extra), l.data...)...)
}
