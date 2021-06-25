package tree

import (
	"github.com/arr-ai/frozen/errors"
)

// masker represents a set of one-bits and the ability to enumerate them.
type masker uint16

// First returns a masker with only the low bit of m.
func (m masker) First() masker {
	return m &^ (m - 1)
}

// FirstIsIn returns true if, and only if, the low bit of m is also in n.
func (m masker) FirstIsIn(n masker) bool {
	return m.First().SubsetOf(n)
}

func (m masker) SubsetOf(mask masker) bool {
	return m&^mask == 0
}

func (m masker) String() string {
	panic(errors.Unimplemented)
	// return brailleEncoded(bits.Reverse64(uint64(m)))
}
