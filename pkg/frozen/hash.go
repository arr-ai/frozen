package frozen

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
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

func (m HashMixer) Interface(i interface{}) HashMixer {
	return m.Hash(hash(i))
}

func (m HashMixer) Value() uint64 {
	return m.seed
}

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

func hash(i interface{}) uint64 {
	switch k := i.(type) {
	case Hashable:
		return k.Hash()
	case [2]interface{}: // Optimisation for hasher.next
		return NewHashMixer(9647128711510533157).Hash(hashInterfaceSlice(k[:])).Value()
	case bool:
		return boolMixer.Hash(hashBool(k)).Value()
	case int:
		return intMixer.Hash(hashInt(int64(k))).Value()
	case int8:
		return int8Mixer.Hash(hashInt(int64(k))).Value()
	case int16:
		return int16Mixer.Hash(hashInt(int64(k))).Value()
	case int32:
		return int32Mixer.Hash(hashInt(int64(k))).Value()
	case int64:
		return int64Mixer.Hash(hashInt(k)).Value()
	case uint:
		return uintMixer.Hash(hashUint(uint64(k))).Value()
	case uint8:
		return uint8Mixer.Hash(hashUint(uint64(k))).Value()
	case uint16:
		return uint16Mixer.Hash(hashUint(uint64(k))).Value()
	case uint32:
		return uint32Mixer.Hash(hashUint(uint64(k))).Value()
	case uint64:
		return uint64Mixer.Hash(hashUint(k)).Value()
	case uintptr:
		return uintptrMixer.Hash(hashUint(uint64(k))).Value()
	case float32:
		return float32Mixer.Hash(hashFloat(float64(k))).Value()
	case float64:
		return float64Mixer.Hash(hashFloat(k)).Value()
	case complex64:
		return complex64Mixer.Hash(hashComplex(complex128(k))).Value()
	case complex128:
		return complex128Mixer.Hash(hashComplex(k)).Value()
	case string:
		return stringMixer.Hash(hashString(k)).Value()
	case []interface{}:
		return NewHashMixer(17001635779303974173).Hash(hashInterfaceSlice(k)).Value()
	default:
		v := reflect.ValueOf(k)
		switch v.Kind() {
		case reflect.Bool:
			return valueMixer(v).Hash(hashBool(v.Bool())).Value()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return valueMixer(v).Hash(hashInt(v.Int())).Value()
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return valueMixer(v).Hash(hashUint(v.Uint())).Value()
		case reflect.UnsafePointer:
			return valueMixer(v).Hash(hashUint(uint64(v.Pointer()))).Value()
		case reflect.Float32, reflect.Float64:
			return valueMixer(v).Hash(hashFloat(v.Float())).Value()
		case reflect.Complex64, reflect.Complex128:
			return valueMixer(v).Hash(hashComplex(v.Complex())).Value()
		case reflect.String:
			return valueMixer(v).Hash(hashString(v.String())).Value()
		case reflect.Struct:
			t := v.Type()
			mixer := valueMixer(v)
			for i := v.NumField(); i > 0; {
				i--
				f := v.Field(i)
				mixer = mixer.Hash(hashString(t.Field(i).Name))
				mixer = mixer.Interface(f.Interface())
			}
			return mixer.Value()
		case reflect.Array:
			mixer := valueMixer(v)
			for i := v.Len(); i > 0; {
				i--
				mixer = mixer.Interface(v.Index(i).Interface())
			}
			return mixer.Value()
		case reflect.Ptr:
			mixer := valueMixer(v)
			mixer = mixer.Hash(hashUint(uint64(v.Pointer())))
			return mixer.Value()
		}
		panic(fmt.Sprintf("value %v has unhashable type %[1]T", i))
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
	u = (^u) + (u << 21) // i = (i << 21) - i - 1;
	u = u ^ (u >> 24)
	u = (u + (u << 3)) + (u << 8) // i * 265
	u = u ^ (u >> 14)
	u = (u + (u << 2)) + (u << 4) // i * 21
	u = u ^ (u >> 28)
	u = u + (u << 31)
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

func hashInterfaceSlice(slice []interface{}) uint64 {
	mixer := NewHashMixer(8687379964562119237)
	for _, elem := range slice {
		mixer = mixer.Interface(elem)
	}
	return mixer.Value()
}

func valueMixer(value reflect.Value) HashMixer {
	t := value.Type()
	return NewHashMixer(756276936141926161).
		Hash(hashUint(uint64(t.Kind()))).
		Hash(hashString(t.PkgPath())).
		Hash(hashString(t.Name()))
}
