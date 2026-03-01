package hash

import (
	"fmt"
	"reflect"
	"unsafe"
)

type any = interface{}

// Interface returns a hash for i.
//
// Deprecated: Use Any instead.
func Interface(a any, seed Seed) Seed {
	return Any(a, seed)
}

// Interface returns a hash for i.
func Any(i any, seed Seed) Seed {
	// The order below guesses frequency rank in real world code.
	switch k := i.(type) {
	case Hashable:
		return k.Hash(seed)
	case int:
		return Int(k, seed)
	case string:
		return String(k, seed)
	case uint64:
		return Uint64(k, seed)
	case float64:
		return Float64(k, seed)
	case bool:
		return Bool(k, seed)
	case uintptr:
		return Uintptr(k, seed)
	case uint:
		return Uint(k, seed)
	case int64:
		return Int64(k, seed)
	case []any:
		return sliceInterfaceHash(k, seed)
	case reflect.Value:
		return Value(k, seed)
	default:
		return Value(reflect.ValueOf(k), seed)
	}
}

// Bool returns a hash for b.
func Bool(b bool, seed Seed) Seed {
	return algarray[algMEM8](noescape(unsafe.Pointer(&b)), seed)
}

// Int returns a hash for x.
func Int(x int, seed Seed) Seed {
	return algarray[algINT](noescape(unsafe.Pointer(&x)), seed)
}

// Int8 returns a hash for x.
func Int8(x int8, seed Seed) Seed {
	return algarray[algMEM8](noescape(unsafe.Pointer(&x)), seed)
}

// Int16 returns a hash for x.
func Int16(x int16, seed Seed) Seed {
	return algarray[algMEM16](noescape(unsafe.Pointer(&x)), seed)
}

// Int32 returns a hash for x.
func Int32(x int32, seed Seed) Seed {
	return algarray[algMEM32](noescape(unsafe.Pointer(&x)), seed)
}

// Int64 returns a hash for x.
func Int64(x int64, seed Seed) Seed {
	return algarray[algMEM64](noescape(unsafe.Pointer(&x)), seed)
}

// Uint returns a hash for x.
func Uint(x uint, seed Seed) Seed {
	return algarray[algUINT](noescape(unsafe.Pointer(&x)), seed)
}

// Uint8 returns a hash for x.
func Uint8(x uint8, seed Seed) Seed {
	return algarray[algMEM8](noescape(unsafe.Pointer(&x)), seed)
}

// Uint16 returns a hash for x.
func Uint16(x uint16, seed Seed) Seed {
	return algarray[algMEM16](noescape(unsafe.Pointer(&x)), seed)
}

// Uint32 returns a hash for x.
func Uint32(x uint32, seed Seed) Seed {
	return algarray[algMEM32](noescape(unsafe.Pointer(&x)), seed)
}

// Uint64 returns a hash for x.
func Uint64(x uint64, seed Seed) Seed {
	return algarray[algMEM64](noescape(unsafe.Pointer(&x)), seed)
}

// Uintptr returns a hash for x.
func Uintptr(x, seed Seed) Seed {
	return algarray[algPTR](noescape(unsafe.Pointer(&x)), seed)
}

// Float32 returns a hash for f.
func Float32(f float32, seed Seed) Seed {
	return algarray[algFLOAT32](noescape(unsafe.Pointer(&f)), seed)
}

// Float64 returns a hash for f.
func Float64(f float64, seed Seed) Seed {
	return algarray[algFLOAT64](noescape(unsafe.Pointer(&f)), seed)
}

// Complex64 returns a hash for c.
func Complex64(c complex64, seed Seed) Seed {
	return algarray[algCPLX64](noescape(unsafe.Pointer(&c)), seed)
}

// Complex128 returns a hash for c.
func Complex128(c complex128, seed Seed) Seed {
	return algarray[algCPLX128](noescape(unsafe.Pointer(&c)), seed)
}

// String returns a hash for s.
func String(s string, seed Seed) Seed {
	return algarray[algSTRING](noescape(unsafe.Pointer(&s)), seed)
}

// UnsafePointer returns a hash for p
func UnsafePointer(p unsafe.Pointer, seed Seed) Seed {
	return Uintptr(uintptr(p), seed)
}

// Value returns a hash for v.
//
//nolint:funlen
func Value(v reflect.Value, seed Seed) Seed {
	// These cause dependency cycles if added to valueHashes.
	switch kind := v.Kind(); kind {
	case reflect.Struct:
		return structHash(v, seed)
	case reflect.Array:
		return arrayHash(v, seed)
	default:
		return valueHashes[kind](v, seed)
	}
}

var valueHashes = func() []func(v reflect.Value, seed Seed) Seed {
	m := map[reflect.Kind]func(v reflect.Value, seed Seed) Seed{
		reflect.Bool:          func(v reflect.Value, seed Seed) Seed { return Bool(v.Bool(), seed) },
		reflect.Int:           func(v reflect.Value, seed Seed) Seed { return Int(int(v.Int()), seed) },
		reflect.Int8:          func(v reflect.Value, seed Seed) Seed { return Int8(int8(v.Int()), seed) },
		reflect.Int16:         func(v reflect.Value, seed Seed) Seed { return Int16(int16(v.Int()), seed) },
		reflect.Int32:         func(v reflect.Value, seed Seed) Seed { return Int32(int32(v.Int()), seed) },
		reflect.Int64:         func(v reflect.Value, seed Seed) Seed { return Int64(int64(v.Int()), seed) },
		reflect.Uint:          func(v reflect.Value, seed Seed) Seed { return Uint(uint(v.Uint()), seed) },
		reflect.Uint8:         func(v reflect.Value, seed Seed) Seed { return Uint8(uint8(v.Uint()), seed) },
		reflect.Uint16:        func(v reflect.Value, seed Seed) Seed { return Uint16(uint16(v.Uint()), seed) },
		reflect.Uint32:        func(v reflect.Value, seed Seed) Seed { return Uint32(uint32(v.Uint()), seed) },
		reflect.Uint64:        func(v reflect.Value, seed Seed) Seed { return Uint64(uint64(v.Uint()), seed) },
		reflect.Uintptr:       func(v reflect.Value, seed Seed) Seed { return Uintptr(uintptr(v.Uint()), seed) },
		reflect.Float32:       func(v reflect.Value, seed Seed) Seed { return Float32(float32(v.Float()), seed) },
		reflect.Float64:       func(v reflect.Value, seed Seed) Seed { return Float64(float64(v.Float()), seed) },
		reflect.Complex64:     func(v reflect.Value, seed Seed) Seed { return Complex64(complex64(v.Complex()), seed) },
		reflect.Complex128:    func(v reflect.Value, seed Seed) Seed { return Complex128(complex128(v.Complex()), seed) },
		reflect.Pointer:       func(v reflect.Value, seed Seed) Seed { return Uintptr(v.Pointer(), seed) },
		reflect.String:        func(v reflect.Value, seed Seed) Seed { return String(v.String(), seed) },
		reflect.UnsafePointer: func(v reflect.Value, seed Seed) Seed { return UnsafePointer(unsafe.Pointer(v.Pointer()), seed) },
	}
	s := make([]func(v reflect.Value, seed Seed) Seed, reflect.UnsafePointer+1)
	for k, v := range m {
		s[k] = v
	}
	for i, f := range s {
		if f == nil {
			s[i] = func(v reflect.Value, seed Seed) Seed {
				panic(fmt.Sprintf("value %v has unhashable type %v", v, v.Type()))
			}
		}
	}
	return s
}()

func structHash(v reflect.Value, seed Seed) Seed {
	t := v.Type()
	for i, n := 0, v.NumField(); i < n; i++ {
		seed = String(t.Field(i).Name, seed)
		seed = Interface(v.Field(i).Interface(), seed)
	}
	return seed
}

func arrayHash(v reflect.Value, seed Seed) Seed {
	for i, n := 0, v.Len(); i < n; i++ {
		seed = Value(v.Index(i), seed)
	}
	return seed
}

func sliceInterfaceHash(slice []any, seed Seed) Seed {
	h := seed
	for _, elem := range slice {
		h = Interface(elem, h)
	}
	return h
}
