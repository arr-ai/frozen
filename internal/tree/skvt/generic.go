package skvt

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/iterator/skvi"
	"github.com/arr-ai/frozen/pkg/kv/skv"
)

type (
	elementT = skv.KeyValue
	Iterator = skvi.Iterator
)

var (
	zero          = elementT{}
	emptyIterator = skvi.Empty
)

func elementEqual(a, b skv.KeyValue) bool {
	return a.Key == b.Key
}

func interfaceAsElement(i interface{}) elementT {
	return i.(skv.KeyValue)
}

func newSliceIterator(slice []skv.KeyValue) Iterator {
	return skvi.NewSliceIterator(slice)
}

// KeyHash hashes using the KeyValue's own key.
var KeyHash = keyHasher(func(skv skv.KeyValue, seed uintptr) uintptr { return skv.Hash(seed) })

// KV creates a skv.KeyValue.
func KV(key string, val interface{}) skv.KeyValue {
	return skv.KeyValue{Key: key, Value: val}
}

func hashValue(v skv.KeyValue, seed uintptr) uintptr {
	return hash.Interface(v.Key, seed)
}

func keyHasher(hash func(v skv.KeyValue, seed uintptr) uintptr) func(v skv.KeyValue, seed uintptr) uintptr {
	return func(v skv.KeyValue, seed uintptr) uintptr {
		return hash(v, seed)
	}
}

func KeyEqual(a, b skv.KeyValue) bool {
	return elementEqual(a, b)
}

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable interface {
	Equal(skv.KeyValue) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key interface {
	Equatable
	hash.Hashable
}
