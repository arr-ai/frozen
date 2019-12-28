package frozen

// MapBuilder provides a more efficient way to build Maps incrementally.
type MapBuilder struct {
	root          *node
	adder         *composer
	remover       *composer
	attemptedAdds int
}

// Count returns the number of entries in the Map under construction.
func (b *MapBuilder) Count() int {
	return b.attemptedAdds - b.failedAdds() - b.successfulRemoves()
}

// Put adds or changes an entry into the Map under construction.
func (b *MapBuilder) Put(key, value interface{}) {
	if b.adder == nil {
		b.adder = newUnionComposer(0)
		b.adder.mutate = true
	}
	b.root = b.root.apply(b.adder, KV(key, value))
	b.attemptedAdds++
}

// Remove removes an entry from the Map under construction.
func (b *MapBuilder) Remove(key interface{}) {
	if b.remover == nil {
		b.remover = newDifferenceComposer(0)
		b.remover.mutate = true
	}
	b.root = b.root.apply(b.remover, KV(key, nil))
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

func (b *MapBuilder) failedAdds() int {
	if b.adder == nil {
		return 0
	}
	return b.adder.delta.input
}

func (b *MapBuilder) successfulRemoves() int {
	if b.remover == nil {
		return 0
	}
	return b.remover.delta.input
}
