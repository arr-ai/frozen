//nolint:dupl
package types

import (
	"math/bits"

	"github.com/arr-ai/frozen/internal/fmtutil"
)

// BitIterator represents a set of one-bits and the ability to enumerate them.
type BitIterator uintptr

func (b BitIterator) Next() BitIterator {
	return b & (b - 1)
}

func (b BitIterator) Index() int {
	return bits.TrailingZeros64(uint64(b))
}

func (b BitIterator) Count() int {
	return bits.OnesCount64(uint64(b))
}

func (b BitIterator) Has(i int) bool {
	return b&(BitIterator(1)<<i) != 0
}

func (b BitIterator) With(i int) BitIterator {
	return b | BitIterator(1)<<i
}

func (b BitIterator) Without(i int) BitIterator {
	return b &^ BitIterator(1) << i
}

func (b BitIterator) String() string {
	return fmtutil.BrailleEncoded(bits.Reverse64(uint64(b)))
}
