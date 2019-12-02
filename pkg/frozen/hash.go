package frozen

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
	"github.com/marcelocantos/frozen/pkg/value"
)

type HashMixer struct {
	seed uint64
}

func NewHashMixer(seed uint64) HashMixer {
	return HashMixer{seed: seed}
}

func (m HashMixer) Hash(h uint64) HashMixer {
	return HashMixer{seed: (m.seed + h) * 11694109732069118083}
}

func (m HashMixer) HashValue(h uint64) uint64 {
	return m.Hash(h).Value()
}

func (m HashMixer) Interface(i interface{}) HashMixer {
	return m.Hash(hash(i))
}

func (m HashMixer) InterfaceValue(h uint64) uint64 {
	return m.Interface(h).Value()
}

func (m HashMixer) Value() uint64 {
	return m.seed
}

//nolint:gochecknoglobals
var (
	// These mixers exist primarily to ensure that different types with the same
	// numeric values yield different hashes. We include all primitive types for
	// completeness.
	boolMixer       = valueMixer(reflect.ValueOf(false))
	intMixer        = valueMixer(reflect.ValueOf(int(0)))
	int8Mixer       = valueMixer(reflect.ValueOf(int8(0)))
	int16Mixer      = valueMixer(reflect.ValueOf(int16(0)))
	int32Mixer      = valueMixer(reflect.ValueOf(int32(0)))
	int64Mixer      = valueMixer(reflect.ValueOf(int64(0)))
	uintMixer       = valueMixer(reflect.ValueOf(uint(0)))
	uint8Mixer      = valueMixer(reflect.ValueOf(uint8(0)))
	uint16Mixer     = valueMixer(reflect.ValueOf(uint16(0)))
	uint32Mixer     = valueMixer(reflect.ValueOf(uint32(0)))
	uint64Mixer     = valueMixer(reflect.ValueOf(uint64(0)))
	uintptrMixer    = valueMixer(reflect.ValueOf(uintptr(0)))
	float32Mixer    = valueMixer(reflect.ValueOf(float32(0)))
	float64Mixer    = valueMixer(reflect.ValueOf(float64(0)))
	complex64Mixer  = valueMixer(reflect.ValueOf(complex64(0)))
	complex128Mixer = valueMixer(reflect.ValueOf(complex128(0)))
	stringMixer     = valueMixer(reflect.ValueOf(string(0)))
)

//nolint:gocyclo,funlen
func hash(i interface{}) uint64 {
	switch k := i.(type) {
	case value.Hashable:
		return k.Hash()
	case [2]interface{}: // Optimisation for hasher.next
		return NewHashMixer(9647128711510533157).HashValue(hashInterfaceSlice(k[:]))
	case bool:
		return boolMixer.HashValue(hashBool(k))
	case int:
		return intMixer.HashValue(hashInt(int64(k)))
	case int8:
		return int8Mixer.HashValue(hashInt(int64(k)))
	case int16:
		return int16Mixer.HashValue(hashInt(int64(k)))
	case int32:
		return int32Mixer.HashValue(hashInt(int64(k)))
	case int64:
		return int64Mixer.HashValue(hashInt(k))
	case uint:
		return uintMixer.HashValue(hashUint(uint64(k)))
	case uint8:
		return uint8Mixer.HashValue(hashUint(uint64(k)))
	case uint16:
		return uint16Mixer.HashValue(hashUint(uint64(k)))
	case uint32:
		return uint32Mixer.HashValue(hashUint(uint64(k)))
	case uint64:
		return uint64Mixer.HashValue(hashUint(k))
	case uintptr:
		return uintptrMixer.HashValue(hashUint(uint64(k)))
	case float32:
		return float32Mixer.HashValue(hashFloat(float64(k)))
	case float64:
		return float64Mixer.HashValue(hashFloat(k))
	case complex64:
		return complex64Mixer.HashValue(hashComplex(complex128(k)))
	case complex128:
		return complex128Mixer.HashValue(hashComplex(k))
	case string:
		return stringMixer.HashValue(hashString(k))
	case []interface{}:
		return NewHashMixer(17001635779303974173).HashValue(hashInterfaceSlice(k))
	default:
		return hashValue(reflect.ValueOf(k))
	}
}

func hashBool(b bool) uint64 {
	if b {
		return 13782953484732871189
	}
	return 60143609790115321
}

func hashInt(i int64) uint64 {
	return hashUint(uint64(i))
}

// https://gist.github.com/badboy/6267743
func hashUint(u uint64) uint64 {
	u = ^u + u<<21
	u ^= u >> 24
	u *= 265
	u ^= u >> 14
	u *= 21
	u ^= u >> 28
	u *= 1 + 1<<31
	return u
}

func hashFloat(f float64) uint64 {
	return xxhash.Sum64((*(*[unsafe.Sizeof(f)]byte)(unsafe.Pointer(&f)))[:])
}

func hashComplex(c complex128) uint64 {
	return xxhash.Sum64((*(*[unsafe.Sizeof(c)]byte)(unsafe.Pointer(&c)))[:])
}

func hashString(s string) uint64 {
	return xxhash.Sum64([]byte(s))
}

func hashValue(v reflect.Value) uint64 {
	switch v.Kind() {
	case reflect.Bool:
		return valueMixer(v).HashValue(hashBool(v.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return valueMixer(v).HashValue(hashInt(v.Int()))
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return valueMixer(v).HashValue(hashUint(v.Uint()))
	case reflect.UnsafePointer:
		return valueMixer(v).HashValue(hashUint(uint64(v.Pointer())))
	case reflect.Float32, reflect.Float64:
		return valueMixer(v).HashValue(hashFloat(v.Float()))
	case reflect.Complex64, reflect.Complex128:
		return valueMixer(v).HashValue(hashComplex(v.Complex()))
	case reflect.String:
		return valueMixer(v).HashValue(hashString(v.String()))
	case reflect.Struct:
		return hashStruct(v)
	case reflect.Array:
		return hashArray(v)
	case reflect.Ptr:
		return valueMixer(v).HashValue(hashUint(uint64(v.Pointer())))
	}
	panic(fmt.Sprintf("value %v has unhashable type %v", v, v.Type()))
}

func hashStruct(v reflect.Value) uint64 {
	t := v.Type()
	mixer := valueMixer(v)
	for i := v.NumField(); i > 0; {
		i--
		f := v.Field(i)
		mixer = mixer.Hash(hashString(t.Field(i).Name))
		mixer = mixer.Interface(f.Interface())
	}
	return mixer.Value()
}

func hashArray(v reflect.Value) uint64 {
	mixer := valueMixer(v)
	for i := v.Len(); i > 0; {
		i--
		mixer = mixer.Interface(v.Index(i).Interface())
	}
	return mixer.Value()
}

func hashInterfaceSlice(slice []interface{}) uint64 {
	mixer := NewHashMixer(8687379964562119237)
	for _, elem := range slice {
		mixer = mixer.Interface(elem)
	}
	return mixer.Value()
}

func valueMixer(val reflect.Value) HashMixer {
	t := val.Type()
	return NewHashMixer(756276936141926161).
		Hash(hashUint(uint64(t.Kind()))).
		Hash(hashString(t.PkgPath())).
		Hash(hashString(t.Name()))
}
