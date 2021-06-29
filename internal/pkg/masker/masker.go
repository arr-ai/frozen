package masker

import (
	"math/bits"

	"github.com/arr-ai/frozen/errors"
)

// Masker represents a set of one-bits and the ability to enumerate them.
type Masker uint16

func NewMasker(i int) Masker {
	return Masker(1) << i
}

// First returns a masker with only the low bit of m.
func (m Masker) First() Masker {
	return m &^ (m - 1)
}

// FirstIsIn returns true if, and only if, the low bit of m is also in n.
func (m Masker) FirstIsIn(n Masker) bool {
	return m.First().SubsetOf(n)
}

func (m Masker) FirstIndex() int {
	return bits.TrailingZeros16(uint16(m))
}

// Next strips the low bit off a Masker.
func (m Masker) Next() Masker {
	return m & (m - 1)
}

func (m Masker) String() string {
	panic(errors.Unimplemented)
	// return brailleEncoded(bits.Reverse64(uint64(m)))
}

func (m Masker) SubsetOf(mask Masker) bool {
	return m&^mask == 0
}
