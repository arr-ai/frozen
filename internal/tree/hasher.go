package tree

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/arr-ai/frozen/internal/fmtutil"
	"github.com/arr-ai/hash"
)

const (
	hashBits       = 8 * int(unsafe.Sizeof(uintptr(0)))
	hashBitsOffset = hashBits - nodeBits
)

type Hasher uintptr

func NewHasher(key interface{}, depth int) Hasher {
	if depth < 0 {
		panic("invalid depth")
	}
	return Hasher(hash.Interface(key, 0)) << (depth * nodeBits)
}

func (h Hasher) Next() Hasher {
	return h << nodeBits
}

func (h Hasher) Hash() int {
	return int(h >> hashBitsOffset)
}

func (h Hasher) String() string {
	const dregs = hashBits % nodeBits
	var s string
	switch nodeBits {
	case 2:
		// TODO(if we care): Output a base-4 number.
		s = fmt.Sprintf("%0*x", hashBits/4, h>>dregs)
	case 3:
		var sb strings.Builder
		sb.WriteByte('#')
		// Braille-encode octal digits in pairs.
		for ; h != 0; h <<= 6 {
			sb.WriteRune(rune(0x2800 + h.Hash() + h.Next().Hash()<<3))
		}
		return sb.String()
	case 4:
		return "#" + fmtutil.BrailleEncoded(uint64(h))
	default:
		panic("not implemented")
	}
	if dregs != 0 {
		s += fmt.Sprintf("%d", h<<(nodeBits-dregs)%NodeCount)
	}
	return strings.TrimRight(s, "0")
}
