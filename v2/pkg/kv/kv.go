package kv

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/fu"
	"github.com/arr-ai/frozen/v2/internal/value"
)

// KeyValue[K, V] represents a key-value pair for insertion into a Map.
type KeyValue[K any, V any] struct {
	Key K
	Value V
}

// KV creates a KeyValue[K, V].
func KV[K any, V any](key K, val V) KeyValue[K, V] {
	return KeyValue[K, V]{Key: key, Value: val}
}

// Hash computes a hash for a KeyValue[K, V].
func (kv KeyValue[K, V]) Hash(seed uintptr) uintptr {
	return hash.Interface(kv.Key, seed)
}

// String returns a string representation of a KeyValue[K, V].
func (kv KeyValue[K, V]) String() string {
	return fmt.Sprintf("%s", kv)
}

// String returns a string representation of a KeyValue[K, V].
func (kv KeyValue[K, V]) Format(f fmt.State, verb rune) {
	fu.Format(kv.Key, f, verb)
	f.Write([]byte{':'})
	fu.Format(kv.Value, f, verb)
}

func KeyEqual[K, V any](a, b KeyValue[K, V]) bool {
	return Equal(a, b)
}

func KeyValueEqual[K, V any](a, b KeyValue[K, V]) bool {
	return value.Equal(a.Key, b.Key) && value.Equal(a.Value, b.Value)
}

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable[K, V any] interface {
	Equal(KeyValue[K, V]) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key interface {
	Equatable
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equatable, that is used
// to perform the test.
func Equal[K, V any](a, b KeyValue[K, V]) bool {
	return value.Equal(a.Key, b.Key)
}