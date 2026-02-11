package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type leaf1[T any] struct {
	data T
}

func newLeaf1[T any](a T) *leaf1[T] {
	return &leaf1[T]{data: a}
}

// maxSplitDepth is the depth beyond which splitLeaf gives up and creates a
// collision node. Two full hash rounds is enough to separate any elements
// with a reasonable hash function; beyond this, we assume pathological
// collisions (e.g. custom hash functions that map distinct values identically).
const maxSplitDepth = levelsPerRound * 2

// splitLeaf creates the minimal branch nesting needed to separate two elements
// that coexisted at the same leaf. It descends until their hash indices diverge.
// If indices match across all rounds up to maxSplitDepth, a collision node is
// created as a fallback.
func splitLeaf[T any](existing, incoming T, depth int) node[T] {
	if depth >= maxSplitDepth {
		return newCollision(existing, incoming)
	}
	existingH := newHasher(existing, depth)
	incomingH := newHasher(incoming, depth)
	ei := existingH.hash()
	ii := incomingH.hash()
	if ei != ii {
		b := &branch[T]{}
		b.p.SetNonNilChild(ei, newLeaf1(existing))
		b.p.SetNonNilChild(ii, newLeaf1(incoming))
		return b
	}
	inner := splitLeaf(existing, incoming, depth+1)
	b := &branch[T]{}
	b.p.SetNonNilChild(ei, inner)
	return b
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

// node[T]

func (l *leaf1[T]) Add(args *CombineArgs[T], v T, depth int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		l.data = args.f(l.data, v)
		return l, 1
	}
	return splitLeaf(l.data, v, depth), 0
}

func (l *leaf1[T]) AddFast(v T, depth int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		l.data = v
		return l, 1
	}
	return splitLeaf(l.data, v, depth), 0
}

func (l *leaf1[T]) Combine(args *CombineArgs[T], n node[T], depth int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *branch[T]:
		return n.Combine(args.Flip(), l, depth)
	case *collision[T]:
		return n.Combine(args.Flip(), l, depth)
	case *leaf1[T]:
		if args.eq(l.data, n.data) {
			return newLeaf1(args.f(l.data, n.data)), 1
		}
		return splitLeaf(l.data, n.data, depth), 0
	default:
		panic("unexpected node type in leaf1.Combine")
	}
}

func (l *leaf1[T]) AppendTo(dest []T) []T {
	if len(dest)+1 > cap(dest) {
		return nil
	}
	return append(dest, l.data)
}

func (l *leaf1[T]) Difference(_ depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	if n.Get(l.data, newHasher(l.data, depth), depth) != nil {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) Empty() bool {
	return false
}

func (l *leaf1[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	l2, is := n.(*leaf1[T])
	return is && args.eq(l.data, l2.data)
}

func (l *leaf1[T]) Get(v T, _ hasher, _ int) *T {
	if value.Equal(l.data, v) {
		return &l.data
	}
	return nil
}

func (l *leaf1[T]) Intersection(_ depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	if n.Get(l.data, newHasher(l.data, depth), depth) != nil {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) Iterator([][]node[T]) Iterator[T] {
	// TODO: Avoid malloc.
	return newSliceIterator([]T{l.data})
}

func (l *leaf1[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(l.data)
}

func (l *leaf1[T]) Remove(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) SubsetOf(_ depth.Gauge, n node[T], depth int) bool {
	return n.Get(l.data, newHasher(l.data, depth), depth) != nil
}

func (l *leaf1[T]) Map(_ *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	return newLeaf1(f(l.data)), 1
}

func (l *leaf1[T]) Vet() int {
	return 1
}

func (l *leaf1[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	if args.Pred(l.data) {
		return l, 1
	}
	return nil, 0
}

func (l *leaf1[T]) With(args *CombineArgs[T], v T, depth int, _ hasher) (_ node[T], matches int) {
	if args.eq(l.data, v) {
		return newLeaf1(args.f(l.data, v)), 1
	}
	return splitLeaf(l.data, v, depth), 0
}

func (l *leaf1[T]) WithFast(v T, depth int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		return newLeaf1(v), 1
	}
	return splitLeaf(l.data, v, depth), 0
}

func (l *leaf1[T]) Without(v T, _ int, _ hasher) (_ node[T], matches int) {
	if value.Equal(l.data, v) {
		return nil, 1
	}
	return l, 0
}

func (l *leaf1[T]) clone() node[T] {
	ret := *l
	return &ret
}
