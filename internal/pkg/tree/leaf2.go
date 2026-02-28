package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type leaf2[T any] struct {
	data [2]T
	h0   uintptr // hash(data[0],0) ^ hash(data[1],0)
	ha   uintptr // hash(data[0],0); hash(data[1],0) = h0 ^ ha
}

func newLeaf2[T any](a, b T) *leaf2[T] {
	hf := getHashFunc[T]()
	ha := hf(a, 0)
	return &leaf2[T]{data: [2]T{a, b}, h0: ha ^ hf(b, 0), ha: ha}
}

func newLeaf2WithHash[T any](a, b T, ha, hb uintptr) *leaf2[T] {
	return &leaf2[T]{data: [2]T{a, b}, h0: ha ^ hb, ha: ha}
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

func (l *leaf2[T]) H0() uintptr { return l.h0 }

// hb returns the cached hash of data[1].
func (l *leaf2[T]) hb() uintptr { return l.h0 ^ l.ha }

// node[T]

func (l *leaf2[T]) Add(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
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
			if args.eq(f, n.data) {
				combined := args.f(f, n.data)
				ch := args.hash(combined, 0)
				var otherH uintptr
				if j == 0 {
					otherH = l.hb()
				} else {
					otherH = l.ha
				}
				return &leaf2[T]{data: [2]T{combined, l.data[1-j]}, h0: ch ^ otherH, ha: ch}, 1
			}
		}
		return &leaf[T]{data: []T{l.data[0], l.data[1], n.data}, h0: l.h0 ^ n.h0}, 0
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
	hashes := [2]uintptr{l.ha, l.hb()}
	var found [2]bool
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth, e, args.hash)
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
	hashes := [2]uintptr{l.ha, l.hb()}
	var found [2]bool
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth, e, args.hash)
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
	if args.eq(a, b) {
		combined := args.f(a, b)
		return newLeaf1WithHash(combined, args.hash(combined, 0)), 1
	}
	return newLeaf2(a, b), 2
}

func (l *leaf2[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data[0], l.data[1])
}

func (l *leaf2[T]) Remove(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return newLeaf1WithHash(l.data[1], l.hb()), 1
	}
	if args.eq(l.data[1], v) {
		return newLeaf1WithHash(l.data[0], l.ha), 1
	}
	return l, 0
}

func (l *leaf2[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	hashes := [2]uintptr{l.ha, l.hb()}
	for i, e := range l.data {
		h := hasherFromCached(hashes[i], depth, e, args.hash)
		if n.Get(args, e, h, depth) == nil {
			return false
		}
	}
	return true
}

func (l *leaf2[T]) Vet() int {
	return 2
}

func (l *leaf2[T]) vetH0(hf func(T, uintptr) uintptr) {
	ha := hf(l.data[0], 0)
	hb := hf(l.data[1], 0)
	if want := ha ^ hb; l.h0 != want {
		panic(fmt.Errorf("leaf2 h0 mismatch: stored %x, computed %x", l.h0, want))
	}
	if l.ha != ha {
		panic(fmt.Errorf("leaf2 ha mismatch: stored %x, computed %x", l.ha, ha))
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
	if args.eq(l.data[0], v) {
		combined := args.f(l.data[0], v)
		ch := args.hash(combined, 0)
		return &leaf2[T]{data: [2]T{combined, l.data[1]}, h0: ch ^ l.hb(), ha: ch}, 1
	}
	if args.eq(l.data[1], v) {
		combined := args.f(l.data[1], v)
		ch := args.hash(combined, 0)
		return &leaf2[T]{data: [2]T{l.data[0], combined}, h0: l.ha ^ ch, ha: l.ha}, 1
	}
	vh := args.hash(v, 0)
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}, h0: l.h0 ^ vh}, 0
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
	hf := getHashFunc[T]()
	vh := hf(v, 0)
	return &leaf[T]{data: []T{l.data[0], l.data[1], v}, h0: l.h0 ^ vh}, 0
}

func (l *leaf2[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data[0], v) {
		return newLeaf1WithHash(l.data[1], l.hb()), 1
	}
	if args.eq(l.data[1], v) {
		return newLeaf1WithHash(l.data[0], l.ha), 1
	}
	return l, 0
}

func (l *leaf2[T]) clone() node[T] {
	ret := *l
	return &ret
}
