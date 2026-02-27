package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

const (
	maxLeafLen = 8

	// maxSplitDepth is the depth beyond which splitLeaf gives up and keeps
	// elements in a single leaf. Two full hash rounds is enough to separate
	// any elements with a reasonable hash function; beyond this, we assume
	// pathological collisions.
	maxSplitDepth = levelsPerRound * 2
)

type leaf[T any] struct {
	data []T
}

// splitLeaf distributes elements across branches by their hash at the given
// depth. Called when a leaf overflows maxLeafLen. If all elements hash to the
// same index, it recurses deeper. At maxSplitDepth, gives up and returns a
// large leaf (pathological hash collision).
func splitLeaf[T any](data []T, depth int, hf func(T, uintptr) uintptr) node[T] {
	if depth >= maxSplitDepth {
		return &leaf[T]{data: data}
	}

	var buckets [fanout][]T
	for _, e := range data {
		i := newHasherWith(e, depth, hf).hash()
		buckets[i] = append(buckets[i], e)
	}

	// Count occupied buckets.
	occupied := 0
	singleIdx := 0
	for i, b := range buckets {
		if len(b) > 0 {
			occupied++
			singleIdx = i
		}
	}

	if occupied == 1 {
		// All elements hash to the same index — recurse deeper.
		inner := splitLeaf(data, depth+1, hf)
		b := &branch[T]{count: len(data)}
		b.p.SetNonNilChild(singleIdx, inner)
		return b
	}

	b := &branch[T]{count: len(data)}
	for i, bucket := range buckets {
		if len(bucket) > maxLeafLen {
			b.p.SetNonNilChild(i, splitLeaf(bucket, depth+1, hf))
		} else if len(bucket) > 0 {
			b.p.SetNonNilChild(i, leafCanonical(bucket))
		}
	}
	return b
}

// leafCanonical returns the simplest node for the given data.
// Callers must pass a freshly-built slice (leafCanonical takes ownership).
func leafCanonical[T any](data []T) node[T] {
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &leaf1[T]{data: data[0]}
	case 2:
		return newLeaf2(data[0], data[1])
	default:
		return &leaf[T]{data: data}
	}
}

// fmt.Formatter

func (l *leaf[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "‹")
	for i, e := range l.data {
		if i > 0 {
			fu.WriteString(f, ",")
		}
		fu.Format(e, f, verb)
	}
	fu.WriteString(f, "›")
}

// fmt.Stringer

func (l *leaf[T]) String() string {
	return fmt.Sprintf("%s", l)
}

// node[T]

// Add inserts v into the leaf, mutating in place. Builder-only: must not be
// called on shared nodes (use With for immutable operations).
func (l *leaf[T]) Add(args *CombineArgs[T], v T, depth int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			l.data[i] = args.f(e, v)
			return l, 1
		}
	}
	if len(l.data) < maxLeafLen {
		l.data = append(l.data, v)
		return l, 0
	}
	return splitLeaf(append(l.data, v), depth, args.hash), 0
}

// AddFast inserts v into the leaf, mutating in place. Builder-only: must not
// be called on shared nodes (use WithFast for immutable operations).
func (l *leaf[T]) AddFast(v T, depth int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if value.Equal(e, v) {
			l.data[i] = v
			return l, 1
		}
	}
	if len(l.data) < maxLeafLen {
		l.data = append(l.data, v)
		return l, 0
	}
	return splitLeaf(append(l.data, v), depth, getHashFunc[T]()), 0
}

func (l *leaf[T]) AppendTo(dest []T) []T {
	if len(dest)+len(l.data) > cap(dest) {
		return nil
	}
	return append(dest, l.data...)
}

func (l *leaf[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf1[T]:
		merged, m := combineLeafSlices(args, append([]T(nil), l.data...), []T{n.data})
		if len(merged) > maxLeafLen {
			return splitLeaf(merged, depth, args.hash), m
		}
		return &leaf[T]{data: merged}, m
	case *leaf2[T]:
		merged, m := combineLeafSlices(args, append([]T(nil), l.data...), n.data[:])
		if len(merged) > maxLeafLen {
			return splitLeaf(merged, depth, args.hash), m
		}
		return leafCanonical(merged), m
	case *leaf[T]:
		merged, m := combineLeafSlices(args, append([]T(nil), l.data...), n.data)
		if len(merged) > maxLeafLen {
			return splitLeaf(merged, depth, args.hash), m
		}
		return &leaf[T]{data: merged}, m
	default:
		panic("unexpected node type in leaf.Combine")
	}
}

func (l *leaf[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
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

func (l *leaf[T]) Empty() bool {
	return len(l.data) == 0
}

func (l *leaf[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	n2, ok := n.(*leaf[T])
	if !ok || len(l.data) != len(n2.data) {
		return false
	}
outer:
	for _, e := range l.data {
		for _, f := range n2.data {
			if args.eq(e, f) {
				continue outer
			}
		}
		return false
	}
	return true
}

func (l *leaf[T]) Get(args *EqArgs[T], v T, _ hasher, _ int) *T {
	for i, e := range l.data {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *leaf[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
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

func (l *leaf[T]) Iterator([][]node[T]) Iterator[T] {
	return newSliceIterator(l.data)
}

func (l *leaf[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	var b Builder[T]
	for _, e := range l.data {
		b.add(args, f(e))
	}
	t := b.Finish()
	return t.root, t.count
}

func (l *leaf[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data...)
}

func (l *leaf[T]) Remove(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			last := len(l.data) - 1
			if last == 0 {
				return nil, 1
			}
			if i < last {
				l.data[i] = l.data[last]
			}
			l.data = l.data[:last]
			switch len(l.data) {
			case 1:
				return &leaf1[T]{data: l.data[0]}, 1
			case 2:
				return newLeaf2(l.data[0], l.data[1]), 1
			default:
				return l, 1
			}
		}
	}
	return l, 0
}

func (l *leaf[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	for _, e := range l.data {
		h := newHasherWith(e, depth, args.hash)
		if n.Get(args, e, h, depth) == nil {
			return false
		}
	}
	return true
}

func (l *leaf[T]) Vet() int {
	if len(l.data) == 0 {
		panic(fmt.Errorf("empty leaf"))
	}
	if len(l.data) <= 2 {
		panic(fmt.Errorf("leaf with %d elements should be leaf1/leaf2", len(l.data)))
	}
	return len(l.data)
}

func (l *leaf[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	var ret []T
	for _, e := range l.data {
		if args.Pred(e) {
			ret = append(ret, e)
			matches++
		}
	}
	return leafCanonical(ret), matches
}

func (l *leaf[T]) With(args *CombineArgs[T], v T, depth int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			ret := &leaf[T]{data: append([]T(nil), l.data...)}
			ret.data[i] = args.f(e, v)
			return ret, 1
		}
	}
	if len(l.data) < maxLeafLen {
		return &leaf[T]{data: append(append([]T(nil), l.data...), v)}, 0
	}
	all := make([]T, len(l.data)+1)
	copy(all, l.data)
	all[len(l.data)] = v
	return splitLeaf(all, depth, args.hash), 0
}

func (l *leaf[T]) WithFast(v T, depth int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if value.Equal(e, v) {
			ret := &leaf[T]{data: append([]T(nil), l.data...)}
			ret.data[i] = v
			return ret, 1
		}
	}
	if len(l.data) < maxLeafLen {
		return &leaf[T]{data: append(append([]T(nil), l.data...), v)}, 0
	}
	all := make([]T, len(l.data)+1)
	copy(all, l.data)
	all[len(l.data)] = v
	return splitLeaf(all, depth, getHashFunc[T]()), 0
}

func (l *leaf[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			switch len(l.data) {
			case 1:
				return nil, 1
			case 2:
				return &leaf1[T]{data: l.data[1-i]}, 1
			case 3:
				var d [2]T
				copy(d[:], l.data[:i])
				copy(d[i:], l.data[i+1:])
				return &leaf2[T]{data: d}, 1
			default:
				ret := make([]T, len(l.data)-1)
				copy(ret, l.data[:i])
				copy(ret[i:], l.data[i+1:])
				return &leaf[T]{data: ret}, 1
			}
		}
	}
	return l, 0
}

func (l *leaf[T]) clone() node[T] {
	return &leaf[T]{data: append([]T(nil), l.data...)}
}

// combineLeafSlices merges src into dest using args.eq/f, returning the merged
// slice and match count.
func combineLeafSlices[T any](args *CombineArgs[T], dest, src []T) ([]T, int) {
	var matches int
	for _, e := range src {
		found := false
		for j, f := range dest {
			if args.eq(f, e) {
				dest[j] = args.f(f, e)
				matches++
				found = true
				break
			}
		}
		if !found {
			dest = append(dest, e)
		}
	}
	return dest, matches
}
