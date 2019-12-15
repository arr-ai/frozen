package frozen

import (
	"unsafe"

	"github.com/marcelocantos/frozen/pkg/value"
)

const hashBits = 8*unsafe.Sizeof(uintptr(0)) - 4

type hasher uintptr

func newHasher(key interface{}, depth int) hasher {
	// Use the high four bits as the seed.
	h := hasher(0b1111<<hashBits | value.Hash(key))
	if depth > 0 {
		d := depth
		if uintptr(d)*nodeBits > hashBits {
			d = int(hashBits / nodeBits)
		}
		h >>= d * nodeBits
		for depth -= d; depth > 0; depth-- {
			h = h.next(key)
		}
	}
	return h
}

func (h hasher) next(key interface{}) hasher {
	if h >>= nodeBits; h < 0b1_0000 {
		return (h-1)<<hashBits | hasher(value.Hash([2]interface{}{int(h), key})>>4)
	}
	return h
}

func (h hasher) hash() int {
	return int(h % nodeSize)
}
