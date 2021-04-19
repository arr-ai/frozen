package frozen

import (
	"encoding/json"
	"fmt"

	"github.com/arr-ai/hash"
)

// KeyValue represents a key-value pair for insertion into a StrMap.
type StrKeyValue struct {
	Key   string
	Value interface{}
}

// KV creates a StrKeyValue.
func StrKV(key string, val interface{}) StrKeyValue {
	return StrKeyValue{Key: key, Value: val}
}

// Hash computes a hash for a StrKeyValue.
func (kv StrKeyValue) Hash(seed uintptr) uintptr {
	return hash.Interface(kv.Key, seed)
}

// Equal returns true iff i is a StrKeyValue whose key equals this StrKeyValue's key.
func (kv StrKeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(StrKeyValue); ok {
		return Equal(kv.Key, kv2.Key)
	}
	return false
}

// String returns a string representation of a StrKeyValue.
func (kv StrKeyValue) String() string {
	return fmt.Sprintf("%#v:%#v", kv.Key, kv.Value)
}

// MapBuilder provides a more efficient way to build Maps incrementally.
type StrMapBuilder struct {
	root          *node
	prepared      *node
	redundantPuts int
	removals      int
	attemptedAdds int
	cloner        *cloner
}

func NewStrMapBuilder(capacity int) *StrMapBuilder {
	return &StrMapBuilder{cloner: newCloner(true, capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *StrMapBuilder) Count() int {
	return b.attemptedAdds - b.redundantPuts - b.removals
}

// Put adds or changes an entry into the Map under construction.
func (b *StrMapBuilder) Put(key string, value interface{}) {
	kv := StrKV(key, value)
	b.root = b.root.with(kv, useRHS, 0, newHasher(kv, 0), &b.redundantPuts, theMutator, &b.prepared)
	b.attemptedAdds++
}

// Remove removes an entry from the Map under construction.
func (b *StrMapBuilder) Remove(key string) {
	kv := StrKV(key, nil)
	b.root = b.root.without(kv, 0, newHasher(kv, 0), &b.removals, theMutator, &b.prepared)
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *StrMapBuilder) Get(key string) (interface{}, bool) {
	if entry := b.root.get(StrKV(key, nil)); entry != nil {
		return entry.(StrKeyValue).Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the StrMapBuilder was
// initialised or the last call to Finish.
func (b *StrMapBuilder) Finish() StrMap {
	count := b.Count()
	if count == 0 {
		return StrMap{}
	}
	root := b.root
	*b = StrMapBuilder{}
	return StrMap{root: root, count: count}
}

// StrMap StrMaps keys to values. The zero value is the empty StrMap.
type StrMap struct {
	root  *node
	count int
}

var _ Key = StrMap{}

// NewStrMap creates a new StrMap with kvs as keys and values.
func NewStrMap(kvs ...StrKeyValue) StrMap {
	var b StrMapBuilder
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewStrMapFromGoStrMap takes a StrMap[interface{}]interface{} and returns a frozen StrMap from it.
func NewStrMapFromGoStrMap(m map[string]interface{}) StrMap {
	mb := NewStrMapBuilder(len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the StrMap has no entries.
func (m StrMap) IsEmpty() bool {
	return m.root == nil
}

// Count returns the number of entries in the StrMap.
func (m StrMap) Count() int {
	return m.count
}

// Any returns an arbitrary entry from the StrMap.
func (m StrMap) Any() (key string, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty StrMap")
}

// With returns a new StrMap with key associated with val and all other keys
// retained from m.
func (m StrMap) With(key string, val interface{}) StrMap {
	kv := StrKV(key, val)
	matches := 0
	var prepared *node
	root := m.root.with(kv, useRHS, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	return StrMap{root: root, count: m.Count() + 1 - matches}
}

// Without returns a new StrMap with all keys retained from m except the elements
// of keys.
func (m StrMap) Without(keys Set) StrMap {
	// TODO: O(m+n)
	root := m.root
	matches := 0
	var prepared *node
	for k := keys.Range(); k.Next(); {
		kv := StrKV(k.Value().(string), nil)
		root = root.without(kv, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	}
	return StrMap{root: root, count: m.Count() - matches}
}

// Has returns true iff the key exists in the StrMap.
func (m StrMap) Has(key string) bool {
	return m.root.get(StrKV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m StrMap) Get(key string) (interface{}, bool) {
	if kv := m.root.get(StrKV(key, nil)); kv != nil {
		return kv.(StrKeyValue).Value, true
	}
	return nil, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m StrMap) MustGet(key string) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m StrMap) GetElse(key string, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m StrMap) GetElseFunc(key string, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the StrMap.
func (m StrMap) Keys() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Key())
	}
	return b.Finish()
}

// Values returns a Set with all the Values in the StrMap.
func (m StrMap) Values() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Value())
	}
	return b.Finish()
}

// Project returns a StrMap with only keys included from this StrMap.
func (m StrMap) Project(keys Set) StrMap {
	return m.Where(func(key string, val interface{}) bool {
		return keys.Has(key)
	})
}

// Where returns a StrMap with only key-value pairs satisfying pred.
func (m StrMap) Where(pred func(key string, val interface{}) bool) StrMap {
	var b StrMapBuilder
	for i := m.Range(); i.Next(); {
		if key, val := i.Entry(); pred(key, val) {
			b.Put(key, val)
		}
	}
	return b.Finish()
}

// StrMap returns a StrMap with keys from this StrMap, but the values replaced by the
// result of calling f.
func (m StrMap) StrMap(f func(key string, val interface{}) interface{}) StrMap {
	var b StrMapBuilder
	for i := m.Range(); i.Next(); {
		key, val := i.Entry()
		b.Put(key, f(key, val))
	}
	return b.Finish()
}

// Reduce returns the result of applying f to each key-value pair on the StrMap.
// The result of each call is used as the acc argument for the next element.
func (m StrMap) Reduce(f func(acc interface{}, key string, val interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

// Merge returns a StrMap from the merging between two StrMaps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m StrMap) Merge(n StrMap, resolve func(key, a, b interface{}) interface{}) StrMap {
	if m.IsEmpty() {
		return n
	}
	matches := 0
	extractAndResolve := func(a, b interface{}) interface{} {
		i := a.(StrKeyValue)
		j := b.(StrKeyValue)
		return StrKV(i.Key, resolve(i.Key, i.Value, j.Value))
	}
	root := m.root.union(n.root, extractAndResolve, 0, &matches, theCopier)
	return StrMap{root: root, count: m.Count() + n.Count() - matches}
}

// Update returns a StrMap with key-value pairs from n added or replacing existing
// keys.
func (m StrMap) Update(n StrMap) StrMap {
	f := useRHS
	if m.Count() >= n.Count() {
		m, n = n, m
		f = useLHS
	}
	matches := 0
	root := m.root.union(n.root, f, 0, &matches, theCopier)
	return StrMap{root: root, count: m.Count() + n.Count() - matches}
}

// Hash computes a hash val for s.
func (m StrMap) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(uintptr(3167960924819262823&uint64(^uintptr(0))), seed)
	for i := m.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), hash.Interface(i.Key(), seed))
	}
	return h
}

// Equal returns true iff i is a StrMap with all the same key-value pairs as this
// StrMap.
func (m StrMap) Equal(i interface{}) bool {
	if n, ok := i.(StrMap); ok {
		c := newCloner(false, m.Count())
		equalAsync := c.noneFalse()
		equal := m.root.equal(n.root, func(a, b interface{}) bool {
			kva := a.(StrKeyValue)
			kvb := b.(StrKeyValue)
			return Equal(kva.Key, kvb.Key) && Equal(kva.Value, kvb.Value)
		}, 0, c)
		return equal && equalAsync()
	}
	return false
}

// String returns a string representatio of the StrMap.
func (m StrMap) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the StrMap into state.
func (m StrMap) Format(state fmt.State, _ rune) {
	state.Write([]byte("("))
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			state.Write([]byte(", "))
		}
		fmt.Fprintf(state, "%v: %v", i.Key(), i.Value())
	}
	state.Write([]byte(")"))
}

// Range returns a StrMapIterator over the StrMap.
func (m StrMap) Range() *StrMapIterator {
	return &StrMapIterator{i: m.root.iterator(m.count)}
}

// MarshalJSON implements json.Marshaler.
func (m StrMap) MarshalJSON() ([]byte, error) {
	proxy := map[string]interface{}{}
	for i := m.Range(); i.Next(); {
		proxy[i.Key()] = i.Value()
	}
	return json.Marshal(proxy)
}

// StrMapIterator provides for iterating over a StrMap.
type StrMapIterator struct {
	i  Iterator
	kv StrKeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *StrMapIterator) Next() bool {
	if i.i.Next() {
		var ok bool
		i.kv, ok = i.i.Value().(StrKeyValue)
		if !ok {
			panic(fmt.Sprintf("Unexpected type: %T", i.i.Value()))
		}
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *StrMapIterator) Key() string {
	return i.kv.Key
}

// Value returns the value for the current entry.
func (i *StrMapIterator) Value() interface{} {
	return i.kv.Value
}

// Entry returns the current key-value pair as two return values.
func (i *StrMapIterator) Entry() (key string, value interface{}) {
	return i.kv.Key, i.kv.Value
}
