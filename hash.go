package frozen

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
)

const (
	hamtBits = 3
	hamtSize = 1 << hamtBits
	hamtMask = hamtSize - 1
)

type hasher uint64

func newHasher(key interface{}, depth int) *hasher {
	// Use the high four bits as the seed.
	h := hasher(0b1111<<60 | hash(key))
	for i := 0; i < depth; i++ {
		h.next(key)
	}
	return &h
}

func (h *hasher) next(key interface{}) int {
	if *h < 0b1_0000 {
		*h = (*h-1)<<60 | hasher(hash([2]interface{}{int(*h), key})>>4)
	}
	i := *h & hamtMask
	*h >>= hamtBits
	return int(i)
}

func hash(key interface{}) uint64 {
	switch k := key.(type) {
	case Hashable:
		return k.Hash()
	case int:
		return hash64shift(uint64(k))
	case int8:
		return hash64shift(uint64(k))
	case int16:
		return hash64shift(uint64(k))
	case int32:
		return hash64shift(uint64(k))
	case int64:
		return hash64shift(uint64(k))
	case uint:
		return hash64shift(uint64(k))
	case uint8:
		return hash64shift(uint64(k))
	case uint16:
		return hash64shift(uint64(k))
	case uint32:
		return hash64shift(uint64(k))
	case uint64:
		return hash64shift(uint64(k))
	case uintptr:
		return hash64shift(uint64(k))
	case float32:
		return xxhash.Sum64((*(*[unsafe.Sizeof(k)]byte)(unsafe.Pointer(&k)))[:])
	case float64:
		return xxhash.Sum64((*(*[unsafe.Sizeof(k)]byte)(unsafe.Pointer(&k)))[:])
	case string:
		return xxhash.Sum64([]byte(k))
	case []interface{}:
		return hashInterfaceSlice(k, 17001635779303974173)
	case [2]interface{}: // Optimisation for hasher.next
		return hashInterfaceSlice(k[:], 9647128711510533157)
	default:
		v := reflect.ValueOf(k)
		switch v.Kind() {
		case reflect.Struct:
			h := xxhash.New()
			t := v.Type()
			// go run github.com/marcelocantos/primal/cmd/random_primes 0x1
			fmt.Fprintf(h, "3ec747ed7761326f:%s:%s:", t.PkgPath(), t.Name())
			n := v.NumField()
			for i := 0; i < n; i++ {
				f := v.Field(i)
				fmt.Fprintf(h, "%s:%d:", t.Field(i).Name, hash(f.Interface()))
			}
			return h.Sum64()
		}
		panic(fmt.Sprintf("key %v has unhashable type %[1]T", key))
	}
}

// https://gist.github.com/badboy/6267743
func hash64shift(key uint64) uint64 {
	key = (^key) + (key << 21) // key = (key << 21) - key - 1;
	key = key ^ (key >> 24)
	key = (key + (key << 3)) + (key << 8) // key * 265
	key = key ^ (key >> 14)
	key = (key + (key << 2)) + (key << 4) // key * 21
	key = key ^ (key >> 28)
	key = key + (key << 31)
	return key
}

func hashInterfaceSlice(slice []interface{}, h uint64) uint64 {
	for _, e := range slice {
		h = 9926087995771043021*h + hash(e)
	}
	return h
}
