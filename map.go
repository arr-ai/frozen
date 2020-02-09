package frozen

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

// Map maps keys to values. The zero value is the empty Map.
type Map struct {
	root  *node
	count int
}

var _ Key = Map{}

// NewMap creates a new Map with kvs as keys and values.
func NewMap(kvs ...KeyValue) Map {
	var b MapBuilder
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewMapFromKeys creates a new Map in which values are computed from keys.
func NewMapFromKeys(keys Set, f func(key interface{}) interface{}) Map {
	var b MapBuilder
	for i := keys.Range(); i.Next(); {
		val := i.Value()
		b.Put(val, f(val))
	}
	return b.Finish()
}

// NewMapFromGoMap takes a map[interface{}]interface{} and returns a frozen Map from it.
func NewMapFromGoMap(m map[interface{}]interface{}) Map {
	mb := NewMapBuilder(len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the Map has no entries.
func (m Map) IsEmpty() bool {
	return m.root == nil
}

// Count returns the number of entries in the Map.
func (m Map) Count() int {
	return m.count
}

// Any returns an arbitrary entry from the Map.
func (m Map) Any() (key, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty map")
}

// With returns a new Map with key associated with val and all other keys
// retained from m.
func (m Map) With(key, val interface{}) Map {
	kv := KV(key, val)
	matches := 0
	var prepared *node
	root := m.root.with(kv, useRHS, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	return Map{root: root, count: m.Count() + 1 - matches}
}

// Without returns a new Map with all keys retained from m except the elements
// of keys.
func (m Map) Without(keys Set) Map {
	// TODO: O(m+n)
	root := m.root
	matches := 0
	var prepared *node
	for k := keys.Range(); k.Next(); {
		kv := KV(k.Value(), nil)
		root = root.without(kv, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	}
	return Map{root: root, count: m.Count() - matches}
}

// Has returns true iff the key exists in the map.
func (m Map) Has(key interface{}) bool {
	return m.root.get(KV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	if kv := m.root.get(KV(key, nil)); kv != nil {
		return kv.(KeyValue).Value, true
	}
	return nil, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m Map) MustGet(key interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m Map) GetElse(key, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m Map) GetElseFunc(key interface{}, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the Map.
func (m Map) Keys() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Key())
	}
	return b.Finish()
}

// Values returns a Set with all the Values in the Map.
func (m Map) Values() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Value())
	}
	return b.Finish()
}

// Project returns a Map with only keys included from this Map.
func (m Map) Project(keys Set) Map {
	return m.Where(func(key, val interface{}) bool {
		return keys.Has(key)
	})
}

// Where returns a Map with only key-value pairs satisfying pred.
func (m Map) Where(pred func(key, val interface{}) bool) Map {
	var b MapBuilder
	for i := m.Range(); i.Next(); {
		if key, val := i.Entry(); pred(key, val) {
			b.Put(key, val)
		}
	}
	return b.Finish()
}

// Map returns a Map with keys from this Map, but the values replaced by the
// result of calling f.
func (m Map) Map(f func(key, val interface{}) interface{}) Map {
	var b MapBuilder
	for i := m.Range(); i.Next(); {
		key, val := i.Entry()
		b.Put(key, f(key, val))
	}
	return b.Finish()
}

// Reduce returns the result of applying f to each key-value pair on the Map.
// The result of each call is used as the acc argument for the next element.
func (m Map) Reduce(f func(acc, key, val interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m Map) Merge(n Map, resolve func(key, a, b interface{}) interface{}) Map {
	if m.IsEmpty() {
		return n
	}
	matches := 0
	extractAndResolve := func(a, b interface{}) interface{} {
		i := a.(KeyValue)
		j := b.(KeyValue)
		return KV(i.Key, resolve(i.Key, i.Value, j.Value))
	}
	root := m.root.union(n.root, extractAndResolve, 0, &matches, theCopier)
	return Map{root: root, count: m.Count() + n.Count() - matches}
}

// Update returns a Map with key-value pairs from n added or replacing existing
// keys.
func (m Map) Update(n Map) Map {
	f := useRHS
	if m.Count() >= n.Count() {
		m, n = n, m
		f = useLHS
	}
	matches := 0
	root := m.root.union(n.root, f, 0, &matches, theCopier)
	return Map{root: root, count: m.Count() + n.Count() - matches}
}

// Hash computes a hash val for s.
func (m Map) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(uintptr(3167960924819262823&uint64(^uintptr(0))), seed)
	for i := m.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), hash.Interface(i.Key(), seed))
	}
	return h
}

// Equal returns true iff i is a Map with all the same key-value pairs as this
// Map.
func (m Map) Equal(i interface{}) bool {
	if n, ok := i.(Map); ok {
		c := newCloner(false, m.Count())
		equalAsync := c.noneFalse()
		equal := m.root.equal(n.root, func(a, b interface{}) bool {
			kva := a.(KeyValue)
			kvb := b.(KeyValue)
			return Equal(kva.Key, kvb.Key) && Equal(kva.Value, kvb.Value)
		}, 0, c)
		return equal && equalAsync()
	}
	return false
}

// String returns a string representatio of the Map.
func (m Map) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map into state.
func (m Map) Format(state fmt.State, _ rune) {
	state.Write([]byte("("))
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			state.Write([]byte(", "))
		}
		fmt.Fprintf(state, "%v: %v", i.Key(), i.Value())
	}
	state.Write([]byte(")"))
}

// Range returns a MapIterator over the Map.
func (m Map) Range() *MapIterator {
	return &MapIterator{i: m.root.iterator(m.count)}
}

// MapIterator provides for iterating over a Map.
type MapIterator struct {
	i  Iterator
	kv KeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *MapIterator) Next() bool {
	if i.i.Next() {
		var ok bool
		i.kv, ok = i.i.Value().(KeyValue)
		if !ok {
			panic(fmt.Sprintf("Unexpected type: %T", i.i.Value()))
		}
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *MapIterator) Key() interface{} {
	return i.kv.Key
}

// Value returns the value for the current entry.
func (i *MapIterator) Value() interface{} {
	return i.kv.Value
}

// Entry returns the current key-value pair as two return values.
func (i *MapIterator) Entry() (key, value interface{}) {
	return i.kv.Key, i.kv.Value
}
