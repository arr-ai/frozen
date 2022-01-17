package fu

import (
	"strings"
)

var brailleBytes = func() [0x100]rune {
	// 7 -> 0 |• •| 3 <- 3
	// 6 -> 1 |• •| 4 <- 2
	// 5 -> 2 |• •| 5 <- 1
	// 4 -> 6 |• •| 7 <- 0
	mappings := [][2]int{
		{7, 0},
		{3, 3},
		{6, 1},
		{2, 4},
		{5, 2},
		{1, 5},
		{4, 6},
		{0, 7},
	}
	var bytes [0x100]rune
	for i := 0; i < 0x100; i++ {
		r := rune(0x2800)
		for _, m := range mappings {
			r |= rune(i) >> uint(m[0]) & 1 << uint(m[1])
		}
		bytes[i] = r
	}
	bytes[0] = '~'
	return bytes
}()

func BrailleEncoded(i uint64) string {
	var sb strings.Builder
	for ; i != 0; i <<= 8 {
		sb.WriteRune(brailleBytes[i>>(64-8)])
	}
	return sb.String()
}
