package frozen

import (
	"math/bits"
	"strconv"
	"strings"
)

// BitIterator represents a set of one-bits and the ability to enumerate them.
type BitIterator uint

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

func (b BitIterator) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; b != 0; i, b = i+1, b.Next() {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(b.Index()))
	}
	sb.WriteByte(']')
	return sb.String()
}
