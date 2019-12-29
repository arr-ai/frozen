package frozen

import (
	"math/bits"
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

var brailleBytes = func() [0x100]rune {
	// 0 -> 0 |• •| 3 <- 4
	// 1 -> 1 |• •| 4 <- 5
	// 2 -> 2 |• •| 5 <- 6
	// 3 -> 6 |• •| 7 <- 7
	mappings := [][2]int{
		{0, 0}, {4, 3},
		{1, 1}, {5, 4},
		{2, 2}, {6, 5},
		{3, 6}, {7, 7},
	}
	var bytes [0x100]rune
	for i := 0; i < 0x100; i++ {
		r := rune(0x2800)
		for _, m := range mappings {
			r |= rune(i) >> m[0] & 1 << m[1]
		}
		bytes[i] = r
	}
	return bytes
}()

func (b BitIterator) String() string {
	var sb strings.Builder
	for ; b != 0; b >>= 8 {
		sb.WriteRune(brailleBytes[b%0x100])
	}
	return sb.String()
}
