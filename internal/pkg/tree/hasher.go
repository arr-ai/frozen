package tree

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/hash"
)

const (
	hashBits       = 8 * int(unsafe.Sizeof(uintptr(0)))
	hashBitsOffset = hashBits - fanoutBits
	levelsPerRound = hashBits / fanoutBits
)

type hasher uintptr

var hashFuncCache sync.Map

// resolveHashFunc returns a non-boxing hash function for T.
func resolveHashFunc[T any]() func(T) uintptr {
	var t T
	if _, ok := any(t).(hash.Hashable); ok {
		return func(key T) uintptr {
			return any(key).(hash.Hashable).Hash() //nolint:forcetypeassert
		}
	}
	// Use reflect.Kind to catch derived types (e.g., type MyFloat float64)
	// that need special hashing (±0, NaN) before falling through to the
	// size-based integer dispatch.
	rt := reflect.TypeOf(t)
	if rt == nil {
		return func(key T) uintptr {
			return hash.Any(key)
		}
	}
	switch rt.Kind() { //nolint:exhaustive
	case reflect.Float32:
		return func(key T) uintptr {
			f := *(*float32)(unsafe.Pointer(&key))
			return hash.Float32(f)
		}
	case reflect.Float64:
		return func(key T) uintptr {
			f := *(*float64)(unsafe.Pointer(&key))
			return hash.Float64(f)
		}
	case reflect.Complex64, reflect.Complex128:
		return func(key T) uintptr {
			return hash.Any(key)
		}
	case reflect.String:
		return func(key T) uintptr {
			s := *(*string)(unsafe.Pointer(&key))
			return hash.String(s)
		}
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			return func(key T) uintptr {
				return hash.Bytes(*(*[]byte)(unsafe.Pointer(&key)))
			}
		}
	}
	return resolveHashFuncBySize[T]()
}

func resolveHashFuncBySize[T any]() func(T) uintptr {
	var t T
	switch unsafe.Sizeof(t) {
	case 1:
		return func(key T) uintptr {
			v := *(*uint8)(unsafe.Pointer(&key))
			return hash.Uint8(v)
		}
	case 2:
		return func(key T) uintptr {
			v := *(*uint16)(unsafe.Pointer(&key))
			return hash.Uint16(v)
		}
	case 4:
		return func(key T) uintptr {
			v := *(*uint32)(unsafe.Pointer(&key))
			return hash.Uint32(v)
		}
	case 8:
		return func(key T) uintptr {
			v := *(*uint64)(unsafe.Pointer(&key))
			return hash.Uint64(v)
		}
	}
	return func(key T) uintptr {
		return hash.Any(key)
	}
}

// GetHashFunc returns a cached hash function for type T.
func GetHashFunc[T any]() func(T) uintptr {
	key := TypeKeyOf[T]()
	if f, ok := hashFuncCache.Load(key); ok {
		return f.(func(T) uintptr) //nolint:forcetypeassert
	}
	fn := resolveHashFunc[T]()
	hashFuncCache.Store(key, fn)
	return fn
}

// TypeKeyOf returns a uintptr that uniquely identifies the type T, suitable for
// use as a sync.Map key. It extracts the type-descriptor word from an eface
// wrapping *T, which is a compile-time constant. Using uintptr instead of
// reflect.Type avoids the expensive nilinterhash→typehash chain that sync.Map
// incurs when hashing interface keys.
func TypeKeyOf[T any]() uintptr {
	var t T
	i := any(&t)
	return *(*uintptr)(unsafe.Pointer(&i))
}

func newHasher[T any](key T, depth int) hasher {
	return newHasherWith(key, depth, GetHashFunc[T]())
}

// hasherFromCached derives the address bits for depth from an element's
// cached hash. Round 0 consumes the hash directly. Later rounds remix it with
// the round number, which is deterministic and needs no access to the
// element; a later round can only separate elements whose hashes differ in
// the bits round 0 leaves unused (see maxSplitDepth). The remix is a plain
// xor-multiply rather than a hash call so that this function, and the
// callers that inline it on the read path, stay within the inlining budget.
func hasherFromCached(h0 uintptr, depth int) hasher {
	round := depth / levelsPerRound
	level := depth % levelsPerRound
	bits := h0
	if round > 0 {
		bits = (h0 ^ uintptr(round)) * roundMix //nolint:gosec // round is a small non-negative depth quotient
	}
	return hasher(bits) << uint(level*fanoutBits)
}

// roundMix is an odd multiplier (the 64-bit golden ratio, truncated on
// 32-bit targets) used to remix a cached hash for addressing rounds after
// the first.
const roundMix = uintptr(0x9E3779B97F4A7C15 & uint64(^uintptr(0)))

func newHasherWith[T any](key T, depth int, hf func(T) uintptr) hasher {
	return hasherFromCached(hf(key), depth)
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
			sb.WriteRune(rune(0x2800 + h.hash() + h.next().hash()<<3)) //nolint:gosec // braille block offset, bounded by fanout
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
