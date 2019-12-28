package frozen

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/marcelocantos/hash"
)

const (
	hashBits       = 8 * int(unsafe.Sizeof(uintptr(0)))
	hashBitsOffset = hashBits - nodeBits
)

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

func (h hasher) String() string {
	const dregs = hashBits % nodeBits
	var s string
	switch nodeBits {
	case 2:
		// TODO(if we care): Output a base-4 number.
		s = fmt.Sprintf("%0*x", hashBits/4, h>>dregs)
	case 3:
		s = fmt.Sprintf("%0*o", hashBits/3, h>>dregs)
	case 4:
		s = fmt.Sprintf("%0*x", hashBits/4, h>>dregs)
	default:
		panic("not implemented")
	}
	if dregs != 0 {
		s += fmt.Sprintf("%d", h<<(nodeBits-dregs)%nodeCount)
	}
	return strings.TrimRight(s, "0")
}
