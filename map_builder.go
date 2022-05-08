package frozen

import (
	"github.com/arr-ai/frozen/internal/pkg/tree"
)

// MapBuilder[K, V] provides a more efficient way to build Maps incrementally.
type MapBuilder[K any, V any] struct {
	tb tree.Builder[mapEntry[K, V]]
}

func NewMapBuilder[K any, V any](capacity int) *MapBuilder[K, V] {
	return &MapBuilder[K, V]{tb: *tree.NewBuilder[mapEntry[K, V]](capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *MapBuilder[K, V]) Count() int {
	return b.tb.Count()
}

// Put adds or changes an entry into the Map under construction.
func (b *MapBuilder[K, V]) Put(key K, value V) {
	b.tb.Add(newMapEntry(key, value))
}

// Remove removes an entry from the Map under construction.
func (b *MapBuilder[K, V]) Remove(key K) {
	b.tb.Remove(newMapKey[K, V](key))
}

func (b *MapBuilder[K, V]) Has(key K) bool {
	_, has := b.Get(key)
	return has
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *MapBuilder[K, V]) Get(key K) (V, bool) {
	if entry := b.tb.Get(newMapKey[K, V](key)); entry != nil {
		return entry.Value, true
	}
	var v V
	return v, false
}

// Finish returns a Map containing all entries added since the MapBuilder[K, V] was
// initialised or the last call to Finish.
func (b *MapBuilder[K, V]) Finish() Map[K, V] {
	return newMap(b.tb.Finish())
}
