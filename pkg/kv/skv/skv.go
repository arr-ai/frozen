package skv

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/fu"
	"github.com/arr-ai/frozen/internal/value"
)

// KeyValue represents a key-value pair for insertion into a Map.
type KeyValue struct {
	Key   string
	Value interface{}
}

// KV creates a KeyValue.
func KV(key string, val interface{}) KeyValue {
	return KeyValue{Key: key, Value: val}
}

// Hash computes a hash for a KeyValue.
func (kv KeyValue) Hash(seed uintptr) uintptr {
	return hash.String(kv.Key, seed)
}

// String returns a string representation of a KeyValue.
func (kv KeyValue) String() string {
	return fmt.Sprintf("%s", kv)
}

// String returns a string representation of a KeyValue.
func (kv KeyValue) Format(f fmt.State, verb rune) {
	fu.Format(kv.Key, f, verb)
	f.Write([]byte{':'})
	fu.Format(kv.Value, f, verb)
}

func KeyEqual(a, b KeyValue) bool {
	return a.Key == b.Key
}

func KeyValueEqual(a, b KeyValue) bool {
	return a.Key == b.Key && value.Equal(a.Value, b.Value)
}

// Equatable represents a type that can be compared for equality with another
// value.
type Equatable interface {
	Equal(KeyValue) bool
}

// Key represents a type that can be used as a key in a Map or a Set.
type Key interface {
	Equatable
	hash.Hashable
}

// Equal returns true iff a == b. If a or b implements Equatable, that is used
// to perform the test.
func Equal(a, b KeyValue) bool {
	return a.Key == b.Key
}
