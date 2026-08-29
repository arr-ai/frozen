package hash

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/arr-ai/hash/hash128"
)

// H128 holds a 128-bit hash computed in a single pass on AES-capable hardware.
type H128 struct{ Lo, Hi uintptr }

func (h H128) String() string {
	return fmt.Sprintf("%x:%x", h.Lo, h.Hi)
}

// Assembly stubs — defined in asm_{amd64,arm64}.s.
// These are seedless: the AES key schedule provides per-process randomisation.
func aeshash32H128(p unsafe.Pointer) H128
func aeshash64H128(p unsafe.Pointer) H128
func aeshashH128(p unsafe.Pointer, s uintptr) H128
func aeshashstrH128(p unsafe.Pointer) H128

// h128Funcs is the dispatch table for H128 hashing, initialised during
// package init. On AES-capable hardware the functions point to the
// single-pass assembly; everywhere else they fall back to two seeded calls.
var h128Funcs struct {
	mem32 func(unsafe.Pointer) H128
	mem64 func(unsafe.Pointer) H128
	memN  func(unsafe.Pointer, uintptr) H128
	str   func(unsafe.Pointer) H128
	f32   func(unsafe.Pointer) H128
	f64   func(unsafe.Pointer) H128
}

func initH128AES() {
	h128Funcs.mem32 = aeshash32H128
	h128Funcs.mem64 = aeshash64H128
	h128Funcs.memN = aeshashH128
	h128Funcs.str = aeshashstrH128
	h128Funcs.f32 = func(p unsafe.Pointer) H128 {
		f := *(*float32)(p)
		switch {
		case f == 0:
			return h128zero()
		case f != f:
			return h128nan()
		default:
			return aeshash32H128(p)
		}
	}
	h128Funcs.f64 = func(p unsafe.Pointer) H128 {
		f := *(*float64)(p)
		switch {
		case f == 0:
			return h128zero()
		case f != f:
			return h128nan()
		default:
			return aeshash64H128(p)
		}
	}
}

func initH128Software() {
	h128Funcs.mem32 = func(p unsafe.Pointer) H128 {
		return H128{memhash32(p, 0), memhash32(p, 1)}
	}
	h128Funcs.mem64 = func(p unsafe.Pointer) H128 {
		return H128{memhash64(p, 0), memhash64(p, 1)}
	}
	h128Funcs.memN = func(p unsafe.Pointer, s uintptr) H128 {
		return H128{memhash(p, 0, s), memhash(p, 1, s)}
	}
	h128Funcs.str = func(p unsafe.Pointer) H128 {
		x := (*stringStruct)(p)
		return H128{memhash(x.str, 0, uintptr(x.len)), memhash(x.str, 1, uintptr(x.len))}
	}
	h128Funcs.f32 = func(p unsafe.Pointer) H128 {
		return H128{f32hash(p, 0), f32hash(p, 1)}
	}
	h128Funcs.f64 = func(p unsafe.Pointer) H128 {
		return H128{f64hash(p, 0), f64hash(p, 1)}
	}
}

// h128zero returns a deterministic H128 for float ±0.
// Uses runtime multiplication to avoid constant overflow on 32-bit arches.
func h128zero() H128 {
	lo := c1
	lo *= c0
	hi := c1
	hi *= c0 ^ 1
	return H128{lo, hi}
}

// h128nan returns a random H128 for float NaN (each NaN hashes differently).
func h128nan() H128 {
	r := uintptr(fastrand())
	lo := c1
	lo *= c0 ^ r
	hi := c1
	hi *= c0 ^ r ^ 1
	return H128{lo, hi}
}

// Uint8H128 returns an H128 hash for x.
func Uint8H128(x uint8) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&x)), 1)
}

// Uint16H128 returns an H128 hash for x.
func Uint16H128(x uint16) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&x)), 2)
}

// Uint32H128 returns an H128 hash for x.
func Uint32H128(x uint32) H128 {
	return h128Funcs.mem32(noescape(unsafe.Pointer(&x)))
}

// Uint64H128 returns an H128 hash for x.
func Uint64H128(x uint64) H128 {
	return h128Funcs.mem64(noescape(unsafe.Pointer(&x)))
}

// Float32H128 returns an H128 hash for f (handles NaN and ±0).
func Float32H128(f float32) H128 {
	return h128Funcs.f32(noescape(unsafe.Pointer(&f)))
}

// Float64H128 returns an H128 hash for f (handles NaN and ±0).
func Float64H128(f float64) H128 {
	return h128Funcs.f64(noescape(unsafe.Pointer(&f)))
}

// StringH128 returns an H128 hash for s.
func StringH128(s string) H128 {
	return h128Funcs.str(noescape(unsafe.Pointer(&s)))
}

// AnyH128 returns an H128 hash for i, dispatching to the appropriate H128
// function by type. Types implementing hash128.Hashable hash themselves in
// one pass; types implementing only the seeded Hashable fall back to two
// seeded calls.
func AnyH128(i any) H128 { //nolint:cyclop
	switch k := i.(type) {
	case hash128.Hashable:
		h := k.Hash128()
		return H128{uintptr(h.Lo), uintptr(h.Hi)}
	case Hashable:
		return H128{k.Hash(0), k.Hash(1)}
	case int:
		return IntH128(k)
	case string:
		return StringH128(k)
	case uint64:
		return Uint64H128(k)
	case float64:
		return Float64H128(k)
	case bool:
		return BoolH128(k)
	case uintptr:
		return UintptrH128(k)
	case uint:
		return UintH128(k)
	case int64:
		return Int64H128(k)
	case reflect.Value:
		return H128{Value(k, 0), Value(k, 1)}
	default:
		return H128{Value(reflect.ValueOf(k), 0), Value(reflect.ValueOf(k), 1)}
	}
}

// BoolH128 returns an H128 hash for b.
func BoolH128(b bool) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&b)), 1)
}

// IntH128 returns an H128 hash for x.
func IntH128(x int) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&x)), unsafe.Sizeof(x))
}

// Int64H128 returns an H128 hash for x.
func Int64H128(x int64) H128 {
	return h128Funcs.mem64(noescape(unsafe.Pointer(&x)))
}

// UintH128 returns an H128 hash for x.
func UintH128(x uint) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&x)), unsafe.Sizeof(x))
}

// UintptrH128 returns an H128 hash for x.
func UintptrH128(x uintptr) H128 {
	return h128Funcs.memN(noescape(unsafe.Pointer(&x)), unsafe.Sizeof(x))
}
