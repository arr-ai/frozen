package frozen

import (
	"unsafe"

	"github.com/marcelocantos/hash"
)

const hashBits = 8*unsafe.Sizeof(uintptr(0)) - 4

type hasher uintptr

func newHasher(key interface{}, depth int) hasher {
	// Use the high four bits as the seed.
	h := hasher(0b1111<<hashBits | hash.Interface(key, 0))
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
	if h >>= nodeBits; h < nodeCount {
		return (h-1)<<hashBits | hasher(hash.Interface(key, uintptr(h))>>4)
	}
	return h
}

func (h hasher) hash() int {
	return int(h & (nodeCount - 1))
}
