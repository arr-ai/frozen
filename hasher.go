package frozen

import (
	"unsafe"

	"github.com/marcelocantos/hash"
)

const hashBitsOffset = 8*int(unsafe.Sizeof(uintptr(0))) - nodeBits

type hasher uintptr

func newHasher(key interface{}, depth int) hasher {
	return hasher(hash.Interface(key, 0)) << (depth * nodeBits)
}

func (h hasher) next() hasher {
	return h << nodeBits
}

func (h hasher) hash() int {
	return int(h >> hashBitsOffset)
}
