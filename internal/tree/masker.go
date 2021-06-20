package tree

import (
	"math/bits"

	"github.com/arr-ai/frozen/errors"
)

// masker represents a set of one-bits and the ability to enumerate them.
type masker uint16

func newMasker(i int) masker {
	return masker(1) << i
}

// first returns a masker with only the low bit of m.
func (m masker) first() masker {
	return m &^ (m - 1)
}

// firstIsIn returns true if, and only if, the low bit of m is also in n.
func (m masker) firstIsIn(n masker) bool {
	return m.first().subsetOf(n)
}

func (m masker) next() masker {
	return m & (m - 1)
}

func (m masker) index() int {
	return bits.TrailingZeros16(uint16(m))
}

// Offset returns the offset of an element in a packed slice whose indices are
// represented by m.
//
// Consider a branch logically containing [nil, node1, nil, leaf, node2, nil,
// nil, nil]. The packed slice will contain [node1, leaf, node2], and m will
// represent this as 0b00011010. Offset computes the low bit, b, of n and
// returns 0 for b ≤ 1, else 1 for b ≤ 3, else 2 for b ≤ 4, else 3.
func (m masker) offset(n masker) int {
	return ((n - 1) &^ n & m).count()
}

func (m masker) count() int {
	return bits.OnesCount16(uint16(m))
}

func (m masker) subsetOf(mask masker) bool {
	return m&^mask == 0
}

func (m masker) String() string {
	panic(errors.Unimplemented)
	// return brailleEncoded(bits.Reverse64(uint64(m)))
}
