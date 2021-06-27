package kvt

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/iterator/kvi"
	"github.com/arr-ai/frozen/internal/value"
	"github.com/arr-ai/frozen/pkg/kv"
)

type (
	elementT = kv.KeyValue
	Iterator = kvi.Iterator
)

var emptyIterator = kvi.Empty

func elementEqual(a, b kv.KeyValue) bool {
	return value.Equal(a.Key, b.Key)
}

func formatElement(f fmt.State, verb rune, e elementT) {
	e.Format(f, verb)
}

func interfaceAsElement(i interface{}) elementT {
	return i.(kv.KeyValue)
}

func isBlank(kv kv.KeyValue) bool {
	return kv.Key == nil
}

func newSliceIterator(slice []kv.KeyValue) Iterator {
	return kvi.NewSliceIterator(slice)
}

// KeyHash hashes using the KeyValue's own key.
var KeyHash = keyHasher(func(kv kv.KeyValue, seed uintptr) uintptr { return kv.Hash(seed) })

// KV creates a kv.KeyValue.
func KV(key, val interface{}) kv.KeyValue {
	return kv.KeyValue{Key: key, Value: val}
}

func hashValue(v kv.KeyValue, seed uintptr) uintptr {
	return hash.Interface(v.Key, seed)
}

func keyHasher(hash func(v kv.KeyValue, seed uintptr) uintptr) func(v kv.KeyValue, seed uintptr) uintptr {
	return func(v kv.KeyValue, seed uintptr) uintptr {
		return hash(v, seed)
	}
}

func KeyEqual(a, b kv.KeyValue) bool {
	return elementEqual(a, b)
}

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable interface {
	Equal(kv.KeyValue) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key interface {
	Equatable
	hash.Hashable
}
