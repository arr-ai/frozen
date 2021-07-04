package frozen

import (
	"github.com/arr-ai/frozen/internal/tree/kvt"
)

// MapBuilder provides a more efficient way to build Maps incrementally.
type MapBuilder struct {
	tb kvt.Builder
}

func NewMapBuilder(capacity int) *MapBuilder {
	return &MapBuilder{tb: *kvt.NewBuilder(capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *MapBuilder) Count() int {
	return b.tb.Count()
}

// Put adds or changes an entry into the Map under construction.
func (b *MapBuilder) Put(key, value interface{}) {
	b.tb.Add(kvt.DefaultNPKeyCombineArgs, KV(key, value))
}

// Remove removes an entry from the Map under construction.
func (b *MapBuilder) Remove(key interface{}) {
	b.tb.Remove(kvt.DefaultNPKeyEqArgs, KV(key, nil))
}

func (b *MapBuilder) Has(v interface{}) bool {
	_, has := b.Get(v)
	return has
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *MapBuilder) Get(key interface{}) (interface{}, bool) {
	if entry := b.tb.Get(kvt.DefaultNPKeyEqArgs, KV(key, nil)); entry != nil {
		return entry.Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the MapBuilder was
// initialised or the last call to Finish.
func (b *MapBuilder) Finish() Map {
	return newMap(b.tb.Finish())
}
