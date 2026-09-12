//go:build !frozen_vet

package tree

import (
	"fmt"
	"math"
	"testing"
)

// shape walks a tree and reports how well the element hashes spread the
// elements: the mean leaf depth, the deepest node, and how many collision
// leaves sit at or below maxSplitDepth (elements the hash could not tell
// apart).
type shape struct {
	elems          int
	depthSum       int
	maxDepth       int
	collisionLeafs int
}

func (s *shape) leaf(depth, n int) {
	s.elems += n
	s.depthSum += n * depth
	if depth > s.maxDepth {
		s.maxDepth = depth
	}
}

func walkShape[T any](n node[T], depth int, s *shape) {
	switch n := n.(type) {
	case *branch[T]:
		for m := n.p.mask; m != 0; m = m.Next() {
			walkShape(n.p.data[m.FirstIndex()], depth+1, s)
		}
	case *leaf1[T]:
		s.leaf(depth, 1)
	case *leaf2[T]:
		s.leaf(depth, 2)
	case *leaf[T]:
		s.leaf(depth, len(n.data))
		if depth >= maxSplitDepth {
			s.collisionLeafs++
		}
	}
}

func assertWellSpread[T any](t *testing.T, name string, keys []T) {
	t.Helper()
	var b Builder[T]
	for _, k := range keys {
		b.Add(k)
	}
	tr := b.Finish()
	var s shape
	walkShape(tr.root, 0, &s)
	if s.elems != len(keys) {
		t.Fatalf("%s: built %d elements, want %d", name, s.elems, len(keys))
	}
	// A perfectly uniform hash fills the trie to about log_fanout(n) levels;
	// leaves of up to maxLeafLen elements sit one or two levels above that.
	ideal := math.Log(float64(len(keys))) / math.Log(fanout)
	mean := float64(s.depthSum) / float64(s.elems)
	t.Logf("%s: n=%d mean depth %.2f (ideal %.2f) max depth %d collision leaves %d",
		name, len(keys), mean, ideal, s.maxDepth, s.collisionLeafs)
	if s.collisionLeafs != 0 {
		t.Errorf("%s: %d collision leaves; the hash failed to separate elements", name, s.collisionLeafs)
	}
	if mean > ideal+1 {
		t.Errorf("%s: mean depth %.2f exceeds ideal %.2f by more than one level", name, mean, ideal)
	}
	if float64(s.maxDepth) > ideal+4 {
		t.Errorf("%s: max depth %d is more than four levels below ideal %.2f", name, s.maxDepth, ideal)
	}
}

func TestHashSpreadsStructuredKeys(t *testing.T) {
	t.Parallel()
	const n = 1 << 16

	ints := func(f func(i int) int) []int {
		keys := make([]int, n)
		for i := range keys {
			keys[i] = f(i)
		}
		return keys
	}
	assertWellSpread(t, "sequential", ints(func(i int) int { return i }))
	assertWellSpread(t, "negative sequential", ints(func(i int) int { return -i }))
	assertWellSpread(t, "high bits only", ints(func(i int) int { return i << 40 }))
	assertWellSpread(t, "multiples of 4096", ints(func(i int) int { return i * 4096 }))
	assertWellSpread(t, "bit 63 toggled", ints(func(i int) int { return i ^ (i%2)<<62 }))

	floats := make([]float64, n)
	for i := range floats {
		floats[i] = float64(i)
	}
	assertWellSpread(t, "sequential float64", floats)

	strs := make([]string, n)
	for i := range strs {
		strs[i] = fmt.Sprintf("key-%d", i)
	}
	assertWellSpread(t, "formatted strings", strs)

	short := make([]string, n)
	for i := range short {
		short[i] = fmt.Sprintf("%04x", i)
	}
	assertWellSpread(t, "4-char hex strings", short)

	small := make([]uint16, 1<<16)
	for i := range small {
		small[i] = uint16(i) //nolint:gosec // i < 1<<16 by construction
	}
	assertWellSpread(t, "every uint16", small)
}
