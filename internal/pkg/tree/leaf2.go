package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type leaf2[T any] struct {
	data [2]T
	h0   H128 // hasH128(data[0]) ^ hasH128(data[1])
	ha   H128 // hasH128(data[0]); hasH128(data[1]) = h0.xor(ha)
}

func newLeaf2[T any](a, b T) *leaf2[T] {
	hf := GetHashFunc[T]()
	ha := newElemH128(a, hf)
	hb := newElemH128(b, hf)
	return &leaf2[T]{data: [2]T{a, b}, h0: ha.xor(hb), ha: ha}
}

func newLeaf2WithHash[T any](a, b T, ha, hb H128) *leaf2[T] {
	return &leaf2[T]{data: [2]T{a, b}, h0: ha.xor(hb), ha: ha}
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

func (l *leaf2[T]) H0() H128 { return l.h0 }

// hb returns the cached hash of data[1].
func (l *leaf2[T]) hb() H128 { return l.h0.xor(l.ha) }

// node[T]

func (l *leaf2[T]) Add(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.Equal(e, v) {
			l.data[i] = args.f(e, v)
			// h0 left stale — Builder.Finish() recomputes.
			return l, 1
		}
	}
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}}, 0
}

func (l *leaf2[T]) AddFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if value.Equal(e, v) {
			l.data[i] = v
			// h0 left stale — Builder.Finish() recomputes.
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
			if args.Equal(f, n.data) {
				combined := args.f(f, n.data)
				ch := newElemH128(combined, args.hf)
				var otherH H128
				if j == 0 {
					otherH = l.hb()
				} else {
					otherH = l.ha
				}
				return &leaf2[T]{data: [2]T{combined, l.data[1-j]}, h0: ch.xor(otherH), ha: ch}, 1
			}
		}
		return &leaf[T]{data: []T{l.data[0], l.data[1], n.data}, h0: l.h0.xor(n.h0)}, 0
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
	hashes := [2]H128{l.ha, l.hb()}
	var found [2]bool
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth)
		if n.Get(args, e, h, depth) != nil {
			found[i] = true
			matches++
		}
	}
	switch matches {
	case 0:
		return l, 0
	case 1:
		survivor := 0
		if found[0] {
			survivor = 1
		}
		return newLeaf1WithHash(l.data[survivor], hashes[survivor]), 1
	default:
		return nil, 2
	}
}

func (l *leaf2[T]) Empty() bool {
	return false
}

func (l *leaf2[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	if n, ok := n.(*leaf2[T]); ok {
		if l.h0 != n.h0 {
			return false
		}
		if args.FullHash() && !l.h0.isZero() {
			return true
		}
		return (args.Equal(l.data[0], n.data[0]) && args.Equal(l.data[1], n.data[1])) ||
			(args.Equal(l.data[0], n.data[1]) && args.Equal(l.data[1], n.data[0]))
	}
	return false
}

func (l *leaf2[T]) Get(args *EqArgs[T], v T, _ hasher, _ int) *T {
	if args.Equal(l.data[0], v) {
		return &l.data[0]
	}
	if args.Equal(l.data[1], v) {
		return &l.data[1]
	}
	return nil
}

func (l *leaf2[T]) Intersection(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	hashes := [2]H128{l.ha, l.hb()}
	var found [2]bool
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth)
		if n.Get(args, e, h, depth) != nil {
			found[i] = true
			matches++
		}
	}
	switch matches {
	case 0:
		return nil, 0
	case 1:
		survivor := 0
		if !found[0] {
			survivor = 1
		}
		return newLeaf1WithHash(l.data[survivor], hashes[survivor]), 1
	default:
		return l, 2
	}
}

func (l *leaf2[T]) Iterator([][]node[T]) Iterator[T] {
	return newSliceIterator(l.data[:])
}

func (l *leaf2[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	a, b := f(l.data[0]), f(l.data[1])
	if args.Equal(a, b) {
		combined := args.f(a, b)
		return newLeaf1WithHash(combined, newElemH128(combined, args.hf)), 1
	}
	return newLeaf2(a, b), 2
}

func (l *leaf2[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data[0], l.data[1])
}

func (l *leaf2[T]) Remove(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.Equal(l.data[0], v) {
		return newLeaf1WithHash(l.data[1], l.hb()), 1
	}
	if args.Equal(l.data[1], v) {
		return newLeaf1WithHash(l.data[0], l.ha), 1
	}
	return l, 0
}

func (l *leaf2[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	hashes := [2]H128{l.ha, l.hb()}
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth)
		if n.Get(args, e, h, depth) == nil {
			return false
		}
	}
	return true
}

func (l *leaf2[T]) Vet() int {
	return 2
}

func (l *leaf2[T]) vetH0(hf func(T) H128) {
	ha := newElemH128(l.data[0], hf)
	hb := newElemH128(l.data[1], hf)
	if want := ha.xor(hb); l.h0 != want {
		panic(fmt.Errorf("leaf2 h0 mismatch: stored %v, computed %v", l.h0, want))
	}
	if l.ha != ha {
		panic(fmt.Errorf("leaf2 ha mismatch: stored %v, computed %v", l.ha, ha))
	}
}

func (l *leaf2[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	m0 := args.Pred(l.data[0])
	m1 := args.Pred(l.data[1])
	switch {
	case m0 && m1:
		return l, 2
	case m0:
		return newLeaf1WithHash(l.data[0], l.ha), 1
	case m1:
		return newLeaf1WithHash(l.data[1], l.hb()), 1
	default:
		return nil, 0
	}
}

func (l *leaf2[T]) With(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.Equal(l.data[0], v) {
		if args.same != nil && args.same(l.data[0], v) {
			return l, 1
		}
		combined := args.f(l.data[0], v)
		ch := newElemH128(combined, args.hf)
		return &leaf2[T]{data: [2]T{combined, l.data[1]}, h0: ch.xor(l.hb()), ha: ch}, 1
	}
	if args.Equal(l.data[1], v) {
		if args.same != nil && args.same(l.data[1], v) {
			return l, 1
		}
		combined := args.f(l.data[1], v)
		ch := newElemH128(combined, args.hf)
		return &leaf2[T]{data: [2]T{l.data[0], combined}, h0: l.ha.xor(ch), ha: l.ha}, 1
	}
	vh := newElemH128(v, args.hf)
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}, h0: l.h0.xor(vh)}, 0
}

func (l *leaf2[T]) WithFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data[0], v) {
		// Equal values have equal hashes — reuse l.ha.
		return &leaf2[T]{data: [2]T{v, l.data[1]}, h0: l.h0, ha: l.ha}, 1
	}
	if value.Equal(l.data[1], v) {
		// Equal values have equal hashes — reuse l.hb().
		return &leaf2[T]{data: [2]T{l.data[0], v}, h0: l.h0, ha: l.ha}, 1
	}
	hf := GetHashFunc[T]()
	vh := newElemH128(v, hf)
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}, h0: l.h0.xor(vh)}, 0
}

func (l *leaf2[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.Equal(l.data[0], v) {
		return newLeaf1WithHash(l.data[1], l.hb()), 1
	}
	if args.Equal(l.data[1], v) {
		return newLeaf1WithHash(l.data[0], l.ha), 1
	}
	return l, 0
}

func (l *leaf2[T]) clone() node[T] {
	ret := *l
	return &ret
}
