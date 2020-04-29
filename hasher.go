package frozen

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/arr-ai/hash"
)

const (
	hashBits       = 8 * int(unsafe.Sizeof(uintptr(0)))
	hashBitsOffset = hashBits - nodeBits
)

type hasher uintptr

func newHasher(key interface{}, depth int) hasher {
	return hasher(hash.Interface(key, 0)) << uint(depth*nodeBits)
}

func (h hasher) next() hasher {
	return h << nodeBits
}

func (h hasher) hash() int {
	return int(h >> uint(hashBitsOffset))
}

func (h hasher) String() string {
	const dregs = hashBits % nodeBits
	var s string
	switch nodeBits {
	case 2:
		// TODO(if we care): Output a base-4 number.
		s = fmt.Sprintf("%0*x", hashBits/4, h>>uint(dregs))
	case 3:
		var sb strings.Builder
		sb.WriteByte('#')
		// Braille-encode octal digits in pairs.
		for ; h != 0; h <<= 6 {
			sb.WriteRune(rune(0x2800 + h.hash() + h.next().hash()<<3))
		}
		return sb.String()
	case 4:
		return "#" + brailleEncoded(uint64(h))
	default:
		panic("not implemented")
	}
	if dregs != 0 {
		s += fmt.Sprintf("%d", h<<uint(nodeBits-dregs)%nodeCount)
	}
	return strings.TrimRight(s, "0")
}
