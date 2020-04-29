//nolint:dupl
package frozen

import (
	"math/bits"
)

// MaskIterator represents a set of one-bits and the ability to enumerate them.
type MaskIterator uint16

func (b MaskIterator) Next() MaskIterator {
	return b & (b - 1)
}

func (b MaskIterator) Index() int {
	return bits.TrailingZeros16(uint16(b))
}

func (b MaskIterator) Count() int {
	return bits.OnesCount16(uint16(b))
}

func (b MaskIterator) Has(i int) bool {
	return b&(MaskIterator(1)<<uint(i)) != 0
}

func (b MaskIterator) With(i int) MaskIterator {
	return b | MaskIterator(1)<<uint(i)
}

func (b MaskIterator) Without(i int) MaskIterator {
	return b &^ MaskIterator(1) << uint(i)
}

func (b MaskIterator) String() string {
	return brailleEncoded(bits.Reverse64(uint64(b)))
}
