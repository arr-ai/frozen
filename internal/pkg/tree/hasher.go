package tree

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/pkg/fu"
)

const (
	hashBits       = 8 * int(unsafe.Sizeof(uintptr(0)))
	hashBitsOffset = hashBits - fanoutBits
	levelsPerRound = hashBits / fanoutBits
)

type hasher uintptr

var hashFuncCache sync.Map

// resolveHashFunc returns a non-boxing hash function for T, checked once per type.
func resolveHashFunc[T any]() func(T, uintptr) uintptr {
	var t T
	if _, ok := any(t).(hash.Hashable); ok {
		return func(key T, seed uintptr) uintptr {
			return any(key).(hash.Hashable).Hash(seed) //nolint:forcetypeassert
		}
	}
	var i any = t
	switch i.(type) {
	case float32:
		return func(key T, seed uintptr) uintptr {
			return hash.Float32(*(*float32)(unsafe.Pointer(&key)), seed)
		}
	case float64:
		return func(key T, seed uintptr) uintptr {
			return hash.Float64(*(*float64)(unsafe.Pointer(&key)), seed)
		}
	}
	switch unsafe.Sizeof(t) {
	case 1:
		return func(key T, seed uintptr) uintptr {
			return hash.Uint8(*(*uint8)(unsafe.Pointer(&key)), seed)
		}
	case 2:
		return func(key T, seed uintptr) uintptr {
			return hash.Uint16(*(*uint16)(unsafe.Pointer(&key)), seed)
		}
	case 4:
		return func(key T, seed uintptr) uintptr {
			return hash.Uint32(*(*uint32)(unsafe.Pointer(&key)), seed)
		}
	case 8:
		return func(key T, seed uintptr) uintptr {
			return hash.Uint64(*(*uint64)(unsafe.Pointer(&key)), seed)
		}
	}
	return func(key T, seed uintptr) uintptr {
		return hash.Any(key, seed)
	}
}

func getHashFunc[T any]() func(T, uintptr) uintptr {
	var t T
	rt := reflect.TypeOf(&t)
	if f, ok := hashFuncCache.Load(rt); ok {
		return f.(func(T, uintptr) uintptr) //nolint:forcetypeassert
	}
	fn := resolveHashFunc[T]()
	hashFuncCache.Store(rt, fn)
	return fn
}

func newHasher[T any](key T, depth int) hasher {
	return newHasherWith(key, depth, getHashFunc[T]())
}

func newHasherWith[T any](key T, depth int, hf func(T, uintptr) uintptr) hasher {
	round := depth / levelsPerRound
	level := depth % levelsPerRound
	return hasher(hf(key, uintptr(round))) << uint(level*fanoutBits)
}

func (h hasher) next() hasher {
	return h << fanoutBits
}

func (h hasher) hash() int {
	return int(h >> uint(hashBitsOffset))
}

func (h hasher) String() string {
	const dregs = hashBits % fanoutBits
	var s string
	switch fanoutBits {
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
		return "#" + fu.BrailleEncoded(uint64(h))
	default:
		panic("not implemented")
	}
	if dregs != 0 {
		s += fmt.Sprintf("%d", h<<uint(fanoutBits-dregs)%fanout)
	}
	return strings.TrimRight(s, "0")
}
