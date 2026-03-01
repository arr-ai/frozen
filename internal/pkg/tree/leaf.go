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
	h0   H128
}

func (l *leaf[T]) H0() H128 { return l.h0 }

// splitLeaf distributes elements across branches by their hash at the given
// depth. Called when a leaf overflows maxLeafLen. If all elements hash to the
// same index, it recurses deeper. At maxSplitDepth, gives up and returns a
// large leaf (pathological hash collision).
func splitLeaf[T any](data []T, depth int, hf func(T) H128) node[T] {
	if depth >= maxSplitDepth {
		var h0 H128
		for _, e := range data {
			h0 = h0.xor(newElemH128(e, hf))
		}
		return &leaf[T]{data: data, h0: h0}
	}

	buckets, bucketH0, occupied, singleIdx := distributeLeaf(data, depth, hf)

	if occupied == 1 {
		// All elements hash to the same index — recurse deeper.
		inner := splitLeaf(data, depth+1, hf)
		b := &branch[T]{count: len(data), h0: inner.H0()}
		b.p.SetNonNilChild(singleIdx, inner)
		return b
	}

	var branchH0 H128
	b := &branch[T]{count: len(data)}
	for i, bucket := range buckets {
		if len(bucket) > maxLeafLen {
			child := splitLeaf(bucket, depth+1, hf)
			b.p.SetNonNilChild(i, child)
			branchH0 = branchH0.xor(child.H0())
		} else if len(bucket) > 0 {
			child := leafCanonicalWithHash(bucket, bucketH0[i])
			b.p.SetNonNilChild(i, child)
			branchH0 = branchH0.xor(child.H0())
		}
	}
	b.h0 = branchH0
	return b
}

// distributeLeaf buckets data by hash at the given depth, returning the
// buckets, per-bucket h0 values, occupied count, and the single index if
// only one bucket is occupied.
func distributeLeaf[T any](
	data []T, depth int, hf func(T) H128,
) (buckets [fanout][]T, bucketH0 [fanout]H128, occupied, singleIdx int) {
	for _, e := range data {
		eh := newElemH128(e, hf)
		idx := hasherFromCached(eh, depth).hash()
		buckets[idx] = append(buckets[idx], e)
		bucketH0[idx] = bucketH0[idx].xor(eh)
	}
	for i, b := range buckets {
		if len(b) > 0 {
			occupied++
			singleIdx = i
		}
	}
	return
}

// leafCanonical returns the simplest node for the given data.
// Callers must pass a freshly-built slice (leafCanonical takes ownership).
// h0 is not set — use leafCanonicalWithHash when the hash is known.
func leafCanonical[T any](data []T) node[T] {
	switch len(data) {
	case 0:
		return nil
	case 1:
		return newLeaf1(data[0])
	case 2:
		return newLeaf2(data[0], data[1])
	default:
		hf := getHashFunc[T]()
		var h0 H128
		for _, e := range data {
			h0 = h0.xor(newElemH128(e, hf))
		}
		return &leaf[T]{data: data, h0: h0}
	}
}

// leafCanonicalWithHash is like leafCanonical but uses a pre-computed h0.
func leafCanonicalWithHash[T any](data []T, h0 H128) node[T] {
	switch len(data) {
	case 0:
		return nil
	case 1:
		return newLeaf1WithHash(data[0], h0)
	case 2:
		// h0 = ha ^ hb. We need ha individually. Must compute it.
		hf := getHashFunc[T]()
		ha := newElemH128(data[0], hf)
		return &leaf2[T]{data: [2]T{data[0], data[1]}, h0: h0, ha: ha}
	default:
		return &leaf[T]{data: data, h0: h0}
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
		return leafCanonical(merged), m
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
		return leafCanonical(merged), m
	default:
		panic("unexpected node type in leaf.Combine")
	}
}

func (l *leaf[T]) elemHashes(hf func(T) H128) []H128 {
	h := make([]H128, len(l.data))
	for i, e := range l.data {
		h[i] = newElemH128(e, hf)
	}
	return h
}

func (l *leaf[T]) Difference(args *EqArgs[T], n node[T], depth int) (_ node[T], matches int) {
	eh := l.elemHashes(args.hash)
	var ret []T
	var retH0 H128
	for i, e := range l.data {
		h := hasherFromCached(eh[i], depth)
		if n.Get(args, e, h, depth) != nil {
			matches++
		} else {
			ret = append(ret, e)
			retH0 = retH0.xor(eh[i])
		}
	}
	return leafCanonicalWithHash(ret, retH0), matches
}

func (l *leaf[T]) Empty() bool {
	return len(l.data) == 0
}

func (l *leaf[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	n2, ok := n.(*leaf[T])
	if !ok || l.h0 != n2.h0 || len(l.data) != len(n2.data) {
		return false
	}
	if args.fullHash && !l.h0.isZero() {
		return true
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
	eh := l.elemHashes(args.hash)
	var ret []T
	var retH0 H128
	for i, e := range l.data {
		h := hasherFromCached(eh[i], depth)
		if n.Get(args, e, h, depth) != nil {
			ret = append(ret, e)
			retH0 = retH0.xor(eh[i])
			matches++
		}
	}
	return leafCanonicalWithHash(ret, retH0), matches
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
			// h0 left stale — Builder.Finish() recomputes.
			switch len(l.data) {
			case 1:
				return &leaf1[T]{data: l.data[0]}, 1
			case 2:
				return &leaf2[T]{data: [2]T{l.data[0], l.data[1]}}, 1
			default:
				return l, 1
			}
		}
	}
	return l, 0
}

func (l *leaf[T]) SubsetOf(args *EqArgs[T], n node[T], depth int) bool {
	eh := l.elemHashes(args.hash)
	for i, e := range l.data {
		h := hasherFromCached(eh[i], depth)
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

func (l *leaf[T]) vetH0(hf func(T) H128) {
	var want H128
	for _, e := range l.data {
		want = want.xor(newElemH128(e, hf))
	}
	if l.h0 != want {
		panic(fmt.Errorf("leaf h0 mismatch: stored %v, computed %v", l.h0, want))
	}
}

func (l *leaf[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	var ret []T
	var retH0 H128
	hf := getHashFunc[T]()
	for _, e := range l.data {
		if args.Pred(e) {
			ret = append(ret, e)
			retH0 = retH0.xor(newElemH128(e, hf))
			matches++
		}
	}
	return leafCanonicalWithHash(ret, retH0), matches
}

func (l *leaf[T]) With(args *CombineArgs[T], v T, depth int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			ret := &leaf[T]{data: append([]T(nil), l.data...)}
			combined := args.f(e, v)
			ret.data[i] = combined
			// h0: remove old element hash, add new.
			ret.h0 = l.h0.xor(newElemH128(e, args.hash)).xor(newElemH128(combined, args.hash))
			return ret, 1
		}
	}
	vh := newElemH128(v, args.hash)
	if len(l.data) < maxLeafLen {
		return &leaf[T]{data: append(append([]T(nil), l.data...), v), h0: l.h0.xor(vh)}, 0
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
			// Equal values have equal hashes — h0 unchanged.
			ret.h0 = l.h0
			return ret, 1
		}
	}
	hf := getHashFunc[T]()
	vh := newElemH128(v, hf)
	if len(l.data) < maxLeafLen {
		return &leaf[T]{data: append(append([]T(nil), l.data...), v), h0: l.h0.xor(vh)}, 0
	}
	all := make([]T, len(l.data)+1)
	copy(all, l.data)
	all[len(l.data)] = v
	return splitLeaf(all, depth, hf), 0
}

func (l *leaf[T]) Without(args *EqArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			eh := newElemH128(e, args.hash)
			remH0 := l.h0.xor(eh) // h0 of remaining elements
			switch len(l.data) {
			case 1:
				return nil, 1
			case 2:
				return newLeaf1WithHash(l.data[1-i], remH0), 1
			case 3:
				var d [2]T
				copy(d[:], l.data[:i])
				copy(d[i:], l.data[i+1:])
				ha := newElemH128(d[0], args.hash)
				return &leaf2[T]{data: d, h0: remH0, ha: ha}, 1
			default:
				ret := make([]T, len(l.data)-1)
				copy(ret, l.data[:i])
				copy(ret[i:], l.data[i+1:])
				return &leaf[T]{data: ret, h0: remH0}, 1
			}
		}
	}
	return l, 0
}

func (l *leaf[T]) clone() node[T] {
	return &leaf[T]{data: append([]T(nil), l.data...), h0: l.h0}
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
