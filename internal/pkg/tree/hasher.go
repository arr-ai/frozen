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

// toH128 converts a hash.H128 (exported fields) to tree's H128 (unexported).
func toH128(h hash.H128) H128 {
	return H128{h.Lo, h.Hi}
}

// resolveHashFunc returns a non-boxing hash function for T that computes both
// hash halves in a single call, returning H128. On AES-capable hardware the
// typed variants (Uint64H128, etc.) compute both halves from a single AES pass.
func resolveHashFunc[T any]() func(T) H128 {
	var t T
	if _, ok := any(t).(hash.Hashable); ok {
		return func(key T) H128 {
			h := any(key).(hash.Hashable) //nolint:forcetypeassert
			return H128{h.Hash(0), h.Hash(1)}
		}
	}
	// Use reflect.Kind to catch derived types (e.g., type MyFloat float64)
	// that need special hashing (±0, NaN) before falling through to the
	// size-based integer dispatch.
	rt := reflect.TypeOf(t)
	if rt == nil {
		return func(key T) H128 {
			return toH128(hash.AnyH128(key))
		}
	}
	switch rt.Kind() { //nolint:exhaustive
	case reflect.Float32:
		return func(key T) H128 {
			f := *(*float32)(unsafe.Pointer(&key))
			return toH128(hash.Float32H128(f))
		}
	case reflect.Float64:
		return func(key T) H128 {
			f := *(*float64)(unsafe.Pointer(&key))
			return toH128(hash.Float64H128(f))
		}
	case reflect.Complex64:
		return func(key T) H128 {
			return toH128(hash.AnyH128(key))
		}
	case reflect.Complex128:
		return func(key T) H128 {
			return toH128(hash.AnyH128(key))
		}
	case reflect.String:
		return func(key T) H128 {
			s := *(*string)(unsafe.Pointer(&key))
			return toH128(hash.StringH128(s))
		}
	}
	switch unsafe.Sizeof(t) {
	case 1:
		return func(key T) H128 {
			v := *(*uint8)(unsafe.Pointer(&key))
			return toH128(hash.Uint8H128(v))
		}
	case 2:
		return func(key T) H128 {
			v := *(*uint16)(unsafe.Pointer(&key))
			return toH128(hash.Uint16H128(v))
		}
	case 4:
		return func(key T) H128 {
			v := *(*uint32)(unsafe.Pointer(&key))
			return toH128(hash.Uint32H128(v))
		}
	case 8:
		return func(key T) H128 {
			v := *(*uint64)(unsafe.Pointer(&key))
			return toH128(hash.Uint64H128(v))
		}
	}
	return func(key T) H128 {
		return toH128(hash.AnyH128(key))
	}
}

// GetHashFunc returns a cached hash function for type T.
func GetHashFunc[T any]() func(T) H128 {
	key := TypeKeyOf[T]()
	if f, ok := hashFuncCache.Load(key); ok {
		return f.(func(T) H128) //nolint:forcetypeassert
	}
	fn := resolveHashFunc[T]()
	hashFuncCache.Store(key, fn)
	return fn
}

var seededHashFuncCache sync.Map

// resolveSeededHashFunc returns a non-boxing seeded hash function for T that
// produces the same values as hash.Any(t, seed). Used by mapEntryHashFunc to
// hash keys directly without boxing the enclosing mapEntry struct.
func resolveSeededHashFunc[T any]() func(T, uintptr) uintptr {
	var t T
	if _, ok := any(t).(hash.Hashable); ok {
		// For interface types, boxing to any is free.
		return func(key T, seed uintptr) uintptr {
			return any(key).(hash.Hashable).Hash(seed) //nolint:forcetypeassert
		}
	}
	rt := reflect.TypeOf(t)
	if rt == nil {
		return func(key T, seed uintptr) uintptr {
			return hash.Any(key, seed)
		}
	}
	switch rt.Kind() { //nolint:exhaustive
	case reflect.Float32:
		return func(key T, seed uintptr) uintptr {
			f := *(*float32)(unsafe.Pointer(&key))
			return hash.Float32(f, seed)
		}
	case reflect.Float64:
		return func(key T, seed uintptr) uintptr {
			f := *(*float64)(unsafe.Pointer(&key))
			return hash.Float64(f, seed)
		}
	case reflect.Complex64, reflect.Complex128:
		return func(key T, seed uintptr) uintptr {
			return hash.Any(key, seed)
		}
	case reflect.String:
		return func(key T, seed uintptr) uintptr {
			s := *(*string)(unsafe.Pointer(&key))
			return hash.String(s, seed)
		}
	}
	switch unsafe.Sizeof(t) {
	case 1:
		return func(key T, seed uintptr) uintptr {
			v := *(*uint8)(unsafe.Pointer(&key))
			return hash.Uint8(v, seed)
		}
	case 2:
		return func(key T, seed uintptr) uintptr {
			v := *(*uint16)(unsafe.Pointer(&key))
			return hash.Uint16(v, seed)
		}
	case 4:
		return func(key T, seed uintptr) uintptr {
			v := *(*uint32)(unsafe.Pointer(&key))
			return hash.Uint32(v, seed)
		}
	case 8:
		return func(key T, seed uintptr) uintptr {
			v := *(*uint64)(unsafe.Pointer(&key))
			return hash.Uint64(v, seed)
		}
	}
	return func(key T, seed uintptr) uintptr {
		return hash.Any(key, seed)
	}
}

// GetSeededHashFunc returns a cached seeded hash function for type T.
func GetSeededHashFunc[T any]() func(T, uintptr) uintptr {
	key := TypeKeyOf[T]()
	if f, ok := seededHashFuncCache.Load(key); ok {
		return f.(func(T, uintptr) uintptr) //nolint:forcetypeassert
	}
	fn := resolveSeededHashFunc[T]()
	seededHashFuncCache.Store(key, fn)
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

// hasherFromCached reconstructs a hasher from a cached H128 hash.
// Round 0 uses h0.lo; round 1 uses h0.hi. No rehashing needed for two rounds,
// which covers depths 0–41 (fanoutBits=3) — far beyond any realistic tree.
func hasherFromCached(h0 H128, depth int) hasher {
	round := depth / levelsPerRound
	level := depth % levelsPerRound
	var bits uintptr
	switch round {
	case 0:
		bits = h0.lo
	case 1:
		bits = h0.hi
	default:
		// Pathological depth (>= 2*levelsPerRound). XOR the halves with the
		// round number to produce a deterministic but distinct hash.
		bits = h0.lo ^ h0.hi ^ uintptr(round)
	}
	return hasher(bits) << uint(level*fanoutBits)
}

func newHasherWith[T any](key T, depth int, hf func(T) H128) hasher {
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
