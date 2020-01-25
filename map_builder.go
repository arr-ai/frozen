package frozen

// MapBuilder provides a more efficient way to build Maps incrementally.
type MapBuilder struct {
	root          *node
	prepared      *node
	redundantPuts int
	removals      int
	attemptedAdds int
	cloner        *cloner
}

func NewMapBuilder(capacity int) *MapBuilder {
	return &MapBuilder{cloner: newCloner(true, capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *MapBuilder) Count() int {
	return b.attemptedAdds - b.redundantPuts - b.removals
}

// Put adds or changes an entry into the Map under construction.
func (b *MapBuilder) Put(key, value interface{}) {
	kv := KV(key, value)
	b.root = b.root.with(kv, useRHS, 0, newHasher(kv, 0), &b.redundantPuts, theMutator, &b.prepared)
	b.attemptedAdds++
}

// Remove removes an entry from the Map under construction.
func (b *MapBuilder) Remove(key interface{}) {
	kv := KV(key, nil)
	b.root = b.root.without(kv, 0, newHasher(kv, 0), &b.removals, theMutator, &b.prepared)
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *MapBuilder) Get(key interface{}) (interface{}, bool) {
	if entry := b.root.get(KV(key, nil)); entry != nil {
		return entry.(KeyValue).Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the MapBuilder was
// initialised or the last call to Finish.
func (b *MapBuilder) Finish() Map {
	count := b.Count()
	if count == 0 {
		return Map{}
	}
	root := b.root
	*b = MapBuilder{}
	return Map{root: root, count: count}
}
