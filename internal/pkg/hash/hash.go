package hash

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Any returns a hash for i.
//
//nolint:cyclop
func Any(i any) uintptr {
	// The order below guesses frequency rank in real world code.
	switch k := i.(type) {
	case Hashable:
		return k.Hash()
	case int:
		return Int(k)
	case string:
		return String(k)
	case []byte:
		return Bytes(k)
	case uint64:
		return Uint64(k)
	case float64:
		return Float64(k)
	case bool:
		return Bool(k)
	case uintptr:
		return Uintptr(k)
	case uint:
		return Uint(k)
	case int64:
		return Int64(k)
	case []any:
		return sliceInterfaceHash(k)
	case reflect.Value:
		return Value(k)
	default:
		return Value(reflect.ValueOf(k))
	}
}

// Width salts keep values of different widths apart when they share a
// numeric value (int(5) and uint8(5) are distinct dynamic values in a
// Set[any], and a Set[any] should not chain them together). floatSalt
// keeps 0.0 from hashing like int(0).
const (
	salt8     = 0xA0761D6478BD642F
	salt16    = 0xE7037ED1A0B428DB
	salt32    = 0x8EBC6AF09C88C6E3
	floatSalt = 0xC2B2AE3D27D4EB4F
)

// Bool returns a hash for b.
func Bool(b bool) uintptr {
	var x uint64
	if b {
		x = 1
	}
	return uintptr(mix64(x ^ salt8))
}

// Int returns a hash for x.
func Int(x int) uintptr {
	return uintptr(mix64(uint64(x))) //nolint:gosec // reinterpreting bits is the point
}

// Int8 returns a hash for x.
func Int8(x int8) uintptr {
	return uintptr(mix64(uint64(uint8(x)) ^ salt8)) //nolint:gosec // reinterpreting bits is the point
}

// Int16 returns a hash for x.
func Int16(x int16) uintptr {
	return uintptr(mix64(uint64(uint16(x)) ^ salt16)) //nolint:gosec // reinterpreting bits is the point
}

// Int32 returns a hash for x.
func Int32(x int32) uintptr {
	return uintptr(mix64(uint64(uint32(x)) ^ salt32)) //nolint:gosec // reinterpreting bits is the point
}

// Int64 returns a hash for x.
func Int64(x int64) uintptr {
	return uintptr(mix64(uint64(x))) //nolint:gosec // reinterpreting bits is the point
}

// Uint returns a hash for x.
func Uint(x uint) uintptr {
	return uintptr(mix64(uint64(x)))
}

// Uint8 returns a hash for x.
func Uint8(x uint8) uintptr {
	return uintptr(mix64(uint64(x) ^ salt8))
}

// Uint16 returns a hash for x.
func Uint16(x uint16) uintptr {
	return uintptr(mix64(uint64(x) ^ salt16))
}

// Uint32 returns a hash for x.
func Uint32(x uint32) uintptr {
	return uintptr(mix64(uint64(x) ^ salt32))
}

// Uint64 returns a hash for x.
func Uint64(x uint64) uintptr {
	return uintptr(mix64(x))
}

// Uintptr returns a hash for x.
func Uintptr(x uintptr) uintptr {
	return uintptr(mix64(uint64(x)))
}

// Float32 returns a hash for f. +0 and -0 hash alike; every NaN hashes
// differently, so that NaN != NaN holds for hashing as it does for
// comparison.
func Float32(f float32) uintptr {
	switch {
	case f == 0:
		return uintptr(mix64(floatSalt ^ salt32))
	case f != f:
		return uintptr(mix64(uint64(fastrand()) ^ floatSalt ^ salt32))
	default:
		return uintptr(mix64(uint64(*(*uint32)(unsafe.Pointer(&f))) ^ floatSalt ^ salt32))
	}
}

// Float64 returns a hash for f. +0 and -0 hash alike; every NaN hashes
// differently, so that NaN != NaN holds for hashing as it does for
// comparison.
func Float64(f float64) uintptr {
	switch {
	case f == 0:
		return uintptr(mix64(floatSalt))
	case f != f:
		return uintptr(mix64(uint64(fastrand()) ^ floatSalt))
	default:
		return uintptr(mix64(*(*uint64)(unsafe.Pointer(&f)) ^ floatSalt))
	}
}

// Complex64 returns a hash for c.
func Complex64(c complex64) uintptr {
	return Combine(Float32(real(c)), Float32(imag(c)))
}

// Complex128 returns a hash for c.
func Complex128(c complex128) uintptr {
	return Combine(Float64(real(c)), Float64(imag(c)))
}

// String returns a hash for s. Sequences of at most 16 bytes go through
// the inlined mixer; longer strings use the AES memhash path.
func String(s string) uintptr {
	n := uintptr(len(s)) //nolint:gosec // length is a size, not a secret
	if n <= shortBytes {
		return mixBytes((*stringStruct)(unsafe.Pointer(&s)).str, n, strSalt)
	}
	return algarray[algSTRING](noescape(unsafe.Pointer(&s)), asUintptr(strSalt))
}

// Bytes returns a hash for b. Same split as String, with a distinct salt
// so []byte("ab") does not hash like "ab".
func Bytes(b []byte) uintptr {
	n := uintptr(len(b)) //nolint:gosec // length is a size, not a secret
	if n <= shortBytes {
		var p unsafe.Pointer
		if n != 0 {
			p = unsafe.Pointer(unsafe.SliceData(b))
		}
		return mixBytes(p, n, byteSalt)
	}
	return memhash(unsafe.Pointer(unsafe.SliceData(b)), asUintptr(byteSalt), n)
}

// UnsafePointer returns a hash for p.
func UnsafePointer(p unsafe.Pointer) uintptr {
	return Uintptr(uintptr(p))
}

// Value returns a hash for v. The type's name is mixed in so a defined
// type (type ID int) does not hash like its underlying type. PkgPath and
// Name are interned in the runtime type descriptor; hashing them is
// cheaper than a sync.Map keyed on reflect.Type, and Type is an interface
// whose own address is not a stable seed.
func Value(v reflect.Value) uintptr {
	return Combine(typeNameHash(v.Type()), valueHash(v))
}

func valueHash(v reflect.Value) uintptr {
	// These cause dependency cycles if added to valueHashes.
	switch kind := v.Kind(); kind { //nolint:exhaustive
	case reflect.Struct:
		return structHash(v)
	case reflect.Array:
		return arrayHash(v)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return Bytes(v.Bytes())
		}
		return valueHashes[kind](v)
	default:
		return valueHashes[kind](v)
	}
}

func typeNameHash(t reflect.Type) uintptr {
	if name := t.Name(); name != "" {
		if pkg := t.PkgPath(); pkg != "" {
			return Combine(String(pkg), String(name))
		}
		return String(name)
	}
	return String(t.String())
}

var valueHashes = func() []func(v reflect.Value) uintptr {
	m := map[reflect.Kind]func(v reflect.Value) uintptr{
		reflect.Bool:       func(v reflect.Value) uintptr { return Bool(v.Bool()) },
		reflect.Int:        func(v reflect.Value) uintptr { return Int(int(v.Int())) },
		reflect.Int8:       func(v reflect.Value) uintptr { return Int8(int8(v.Int())) },
		reflect.Int16:      func(v reflect.Value) uintptr { return Int16(int16(v.Int())) },
		reflect.Int32:      func(v reflect.Value) uintptr { return Int32(int32(v.Int())) },
		reflect.Int64:      func(v reflect.Value) uintptr { return Int64(v.Int()) },
		reflect.Uint:       func(v reflect.Value) uintptr { return Uint(uint(v.Uint())) },
		reflect.Uint8:      func(v reflect.Value) uintptr { return Uint8(uint8(v.Uint())) },
		reflect.Uint16:     func(v reflect.Value) uintptr { return Uint16(uint16(v.Uint())) },
		reflect.Uint32:     func(v reflect.Value) uintptr { return Uint32(uint32(v.Uint())) },
		reflect.Uint64:     func(v reflect.Value) uintptr { return Uint64(v.Uint()) },
		reflect.Uintptr:    func(v reflect.Value) uintptr { return Uintptr(uintptr(v.Uint())) },
		reflect.Float32:    func(v reflect.Value) uintptr { return Float32(float32(v.Float())) },
		reflect.Float64:    func(v reflect.Value) uintptr { return Float64(v.Float()) },
		reflect.Complex64:  func(v reflect.Value) uintptr { return Complex64(complex64(v.Complex())) },
		reflect.Complex128: func(v reflect.Value) uintptr { return Complex128(v.Complex()) },
		reflect.Pointer:    func(v reflect.Value) uintptr { return Uintptr(v.Pointer()) },
		reflect.String:     func(v reflect.Value) uintptr { return String(v.String()) },
		reflect.UnsafePointer: func(v reflect.Value) uintptr {
			return UnsafePointer(unsafe.Pointer(v.Pointer()))
		},
	}
	s := make([]func(v reflect.Value) uintptr, reflect.UnsafePointer+1)
	for k, v := range m {
		s[k] = v
	}
	for i, f := range s {
		if f == nil {
			s[i] = func(v reflect.Value) uintptr {
				panic(fmt.Sprintf("value %v has unhashable type %v", v, v.Type()))
			}
		}
	}
	return s
}()

// Salts that start the fold for each composite shape, so an empty struct,
// an empty array and an empty slice hash differently.
const (
	structSalt = 0x1D8E4E27C47D124F
	arraySalt  = 0x589965CC75374CC3
	sliceSalt  = 0x2D358DCCAA6C78A5
)

func structHash(v reflect.Value) uintptr {
	t := v.Type()
	h := asUintptr(structSalt)
	for i, n := 0, v.NumField(); i < n; i++ {
		h = Combine(h, String(t.Field(i).Name))
		h = Combine(h, Any(v.Field(i).Interface()))
	}
	return h
}

func arrayHash(v reflect.Value) uintptr {
	h := asUintptr(arraySalt)
	for i, n := 0, v.Len(); i < n; i++ {
		h = Combine(h, Value(v.Index(i)))
	}
	return h
}

func sliceInterfaceHash(slice []any) uintptr {
	h := asUintptr(sliceSalt)
	for _, elem := range slice {
		h = Combine(h, Any(elem))
	}
	return h
}
