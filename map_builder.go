package frozen

// MapBuilder provides a more efficient way to build Maps incrementally.
type MapBuilder struct {
	nb nodeBuilder
}

func NewMapBuilder(capacity int) *MapBuilder {
	return &MapBuilder{nb: *newNodeBuilder(capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *MapBuilder) Count() int {
	return b.nb.Count()
}

// Put adds or changes an entry into the Map under construction.
func (b *MapBuilder) Put(key, value interface{}) {
	b.nb.Add(defaultNPKeyCombineArgs, KV(key, value))
}

// Remove removes an entry from the Map under construction.
func (b *MapBuilder) Remove(key interface{}) {
	b.nb.Remove(defaultNPKeyEqArgs, KV(key, nil))
}

func (b *MapBuilder) Has(v interface{}) bool {
	_, has := b.Get(v)
	return has
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *MapBuilder) Get(key interface{}) (interface{}, bool) {
	if entry := b.nb.Get(defaultNPKeyEqArgs, KV(key, nil)); entry != nil {
		return (*entry).(KeyValue).Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the MapBuilder was
// initialised or the last call to Finish.
func (b *MapBuilder) Finish() Map {
	return newMap(b.nb.Finish())
}
