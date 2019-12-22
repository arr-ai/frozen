package frozen

import "math/bits"

type bititer uint64

func (b bititer) next() bititer {
	return b & (b - 1)
}

func (b bititer) index() int {
	return bits.TrailingZeros64(uint64(b))
}
