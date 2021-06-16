package frozen

// StringMapBuilder provides a more efficient way to build Maps incrementally.
type StringMapBuilder struct {
	nb nodeBuilder
}

func NewStringMapBuilder(capacity int) *StringMapBuilder {
	return &StringMapBuilder{nb: *newNodeBuilder(capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *StringMapBuilder) Count() int {
	return b.nb.Count()
}

// Put adds or changes an entry into the Map under construction.
func (b *StringMapBuilder) Put(key string, value interface{}) {
	b.nb.Add(defaultNPStringKeyCombineArgs, StringKV(key, value))
}

// Remove removes an entry from the Map under construction.
func (b *StringMapBuilder) Remove(key string) {
	b.nb.Remove(defaultNPStringKeyEqArgs, StringKV(key, nil))
}

func (b *StringMapBuilder) Has(key string) bool {
	_, has := b.Get(key)
	return has
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *StringMapBuilder) Get(key string) (interface{}, bool) {
	if entry := b.nb.Get(defaultNPStringKeyEqArgs, StringKV(key, nil)); entry != nil {
		return (*entry).(StringKeyValue).Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the StringMapBuilder was
// initialised or the last call to Finish.
func (b *StringMapBuilder) Finish() StringMap {
	return newStringMap(b.nb.Finish())
}
