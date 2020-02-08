package types

import (
	"fmt"

	"github.com/arr-ai/hash"
)

// KeyValue represents a key-value pair for insertion into a Map.
type KeyValue struct {
	Key, Value interface{}
}

// KV creates a KeyValue.
func KV(key, val interface{}) KeyValue {
	return KeyValue{Key: key, Value: val}
}

// Hash computes a hash for a KeyValue.
func (kv KeyValue) Hash(seed uintptr) uintptr {
	return hash.Interface(kv.Key, seed)
}

// Equal returns true iff i is a KeyValue whose key equals this KeyValue's key.
func (kv KeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(KeyValue); ok {
		return Equal(kv.Key, kv2.Key)
	}
	return false
}

// String returns a string representation of a KeyValue.
func (kv KeyValue) String() string {
	return fmt.Sprintf("%#v:%#v", kv.Key, kv.Value)
}
