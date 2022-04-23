package frozen

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/pkg/value"
)

type mapEntry[K, V any] struct {
	KeyValue[K, V]
}

func newMapEntry[K, V any](k K, v V) mapEntry[K, V] {
	return mapEntry[K, V]{KeyValue: KeyValue[K, V]{Key: k, Value: v}}
}

func newMapKey[K, V any](k K) mapEntry[K, V] {
	var v V
	return mapEntry[K, V]{KeyValue: KeyValue[K, V]{Key: k, Value: v}}
}

func (e mapEntry[K, V]) Equal(e2 mapEntry[K, V]) bool {
	return value.Equal(e.Key, e2.Key)
}

// Hash computes a hash for a mapEntry[K, V].
func (e mapEntry[K, V]) Hash(seed uintptr) uintptr {
	return hash.Interface(e.Key, seed)
}

// mapEntryHash hashes using the KeyValue's own key.
func mapEntryEqual[K, V any](a, b mapEntry[K, V]) bool {
	return value.Equal(a.Key, b.Key) && value.Equal(a.Value, b.Value)
}

// mapEntryHash hashes using the KeyValue's own key.
func mapEntryKeyEqual[K, V any](a, b mapEntry[K, V]) bool {
	return value.Equal(a.Key, b.Key)
}

// mapEntryHash hashes using the KeyValue's own key.
func mapEntryHash[K, V any](e mapEntry[K, V], seed uintptr) uintptr {
	return e.Hash(seed)
}

func mapEntryKeyHash[K, V any](v mapEntry[K, V], seed uintptr) uintptr {
	return hash.Interface(v.Key, seed)
}
