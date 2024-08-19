package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

// KeyValue[K, V] represents a key-value pair for insertion into a Map.
type KeyValue[K, V any] struct {
	Key   K
	Value V
}

// KV creates a KeyValue[K, V].
func KV[K, V any](k K, v V) KeyValue[K, V] {
	return KeyValue[K, V]{Key: k, Value: v}
}

// Hash computes a hash for a KeyValue[K, V].
func (kv KeyValue[K, V]) Hash(seed uintptr) uintptr {
	return hash.Any(kv.Key, seed)
}

// String returns a string representation of a KeyValue[K, V].
func (kv KeyValue[K, V]) String() string {
	return fmt.Sprintf("%s", kv)
}

// String returns a string representation of a KeyValue[K, V].
func (kv KeyValue[K, V]) Format(f fmt.State, verb rune) {
	fu.Format(kv.Key, f, verb)
	f.Write([]byte{':'}) //nolint:errcheck
	fu.Format(kv.Value, f, verb)
}

func (kv KeyValue[K, V]) Equal(kv2 KeyValue[K, V]) bool {
	return value.Equal(kv.Key, kv2.Key) && value.Equal(kv.Value, kv2.Value)
}

func (kv KeyValue[K, V]) Same(a any) bool {
	kv2, is := a.(KeyValue[K, V])
	return is && kv.Equal(kv2)
}
