//nolint:dupl
package frozen

import (
	"encoding/json"
	"fmt"

	"github.com/arr-ai/hash"
)

// StringKeyValue represents a key-value pair for insertion into a StringMap.
type StringKeyValue struct {
	Key   string
	Value interface{}
}

// StringKV creates a StrKeyValue.
func StringKV(key string, val interface{}) StringKeyValue {
	return StringKeyValue{Key: key, Value: val}
}

// Hash computes a hash for a StrKeyValue.
func (kv StringKeyValue) Hash(seed uintptr) uintptr {
	return hash.Interface(kv.Key, seed)
}

// Equal returns true iff i is a StrKeyValue whose key equals this StrKeyValue's key.
func (kv StringKeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(StringKeyValue); ok {
		return Equal(kv.Key, kv2.Key)
	}
	return false
}

// String returns a string representation of a StrKeyValue.
func (kv StringKeyValue) String() string {
	return fmt.Sprintf("%#v:%#v", kv.Key, kv.Value)
}

// StringMapBuilder provides a more efficient way to build Maps incrementally.
type StringMapBuilder struct {
	root          *branch
	prepared      *branch
	redundantPuts int
	removals      int
	attemptedAdds int
	cloner        *cloner
}

func NewStringMapBuilder(capacity int) *StringMapBuilder {
	return &StringMapBuilder{cloner: newCloner(true, capacity)}
}

// Count returns the number of entries in the Map under construction.
func (b *StringMapBuilder) Count() int {
	return b.attemptedAdds - b.redundantPuts - b.removals
}

// Put adds or changes an entry into the Map under construction.
func (b *StringMapBuilder) Put(key string, value interface{}) {
	kv := StringKV(key, value)
	b.root = b.root.with(kv, useRHS, 0, newHasher(kv, 0), &b.redundantPuts, theMutator, &b.prepared)
	b.attemptedAdds++
}

// Remove removes an entry from the Map under construction.
func (b *StringMapBuilder) Remove(key string) {
	kv := StringKV(key, nil)
	b.root = b.root.without(kv, 0, newHasher(kv, 0), &b.removals, theMutator, &b.prepared)
}

// Get returns the value for key from the Map under construction or false if
// not found.
func (b *StringMapBuilder) Get(key string) (interface{}, bool) {
	if entry := b.root.get(StringKV(key, nil)); entry != nil {
		return entry.(StringKeyValue).Value, true
	}
	return nil, false
}

// Finish returns a Map containing all entries added since the StringMapBuilder was
// initialised or the last call to Finish.
func (b *StringMapBuilder) Finish() StringMap {
	count := b.Count()
	if count == 0 {
		return StringMap{}
	}
	root := b.root
	*b = StringMapBuilder{}
	return StringMap{root: root, count: count}
}

// StringMap StringMaps keys to values. The zero value is the empty StringMap.
type StringMap struct {
	root  *branch
	count int
}

var _ Key = StringMap{}

// NewStringMap creates a new StringMap with kvs as keys and values.
func NewStringMap(kvs ...StringKeyValue) StringMap {
	var b StringMapBuilder
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewStringMapFromGoStringMap takes a StringMap[interface{}]interface{} and returns a frozen StringMap from it.
func NewStringMapFromGoStringMap(m map[string]interface{}) StringMap {
	mb := NewStringMapBuilder(len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the StringMap has no entries.
func (m StringMap) IsEmpty() bool {
	return m.root == nil
}

// Count returns the number of entries in the StringMap.
func (m StringMap) Count() int {
	return m.count
}

// Any returns an arbitrary entry from the StringMap.
func (m StringMap) Any() (key string, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty StringMap")
}

// With returns a new StringMap with key associated with val and all other keys
// retained from m.
func (m StringMap) With(key string, val interface{}) StringMap {
	kv := StringKV(key, val)
	matches := 0
	var prepared *branch
	root := m.root.with(kv, useRHS, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	return StringMap{root: root, count: m.Count() + 1 - matches}
}

// Without returns a new StringMap with all keys retained from m except the elements
// of keys.
func (m StringMap) Without(keys Set) StringMap {
	// TODO: O(m+n)
	root := m.root
	matches := 0
	var prepared *branch
	for k := keys.Range(); k.Next(); {
		kv := StringKV(k.Value().(string), nil)
		root = root.without(kv, 0, newHasher(kv, 0), &matches, theCopier, &prepared)
	}
	return StringMap{root: root, count: m.Count() - matches}
}

// Has returns true iff the key exists in the StringMap.
func (m StringMap) Has(key string) bool {
	return m.root.get(StringKV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m StringMap) Get(key string) (interface{}, bool) {
	if kv := m.root.get(StringKV(key, nil)); kv != nil {
		return kv.(StringKeyValue).Value, true
	}
	return nil, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m StringMap) MustGet(key string) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m StringMap) GetElse(key string, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m StringMap) GetElseFunc(key string, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the StringMap.
func (m StringMap) Keys() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Key())
	}
	return b.Finish()
}

// Values returns a Set with all the Values in the StringMap.
func (m StringMap) Values() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Value())
	}
	return b.Finish()
}

// Project returns a StringMap with only keys included from this StringMap.
func (m StringMap) Project(keys Set) StringMap {
	return m.Where(func(key string, val interface{}) bool {
		return keys.Has(key)
	})
}

// Where returns a StringMap with only key-value pairs satisfying pred.
func (m StringMap) Where(pred func(key string, val interface{}) bool) StringMap {
	var b StringMapBuilder
	for i := m.Range(); i.Next(); {
		if key, val := i.Entry(); pred(key, val) {
			b.Put(key, val)
		}
	}
	return b.Finish()
}

// StringMap returns a StringMap with keys from this StringMap, but the values replaced by the
// result of calling f.
func (m StringMap) StringMap(f func(key string, val interface{}) interface{}) StringMap {
	var b StringMapBuilder
	for i := m.Range(); i.Next(); {
		key, val := i.Entry()
		b.Put(key, f(key, val))
	}
	return b.Finish()
}

// Reduce returns the result of applying f to each key-value pair on the StringMap.
// The result of each call is used as the acc argument for the next element.
func (m StringMap) Reduce(
	f func(acc interface{}, key string, val interface{}) interface{}, acc interface{},
) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

// Merge returns a StringMap from the merging between two StringMaps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m StringMap) Merge(n StringMap, resolve func(key, a, b interface{}) interface{}) StringMap {
	if m.IsEmpty() {
		return n
	}
	matches := 0
	extractAndResolve := func(a, b interface{}) interface{} {
		i := a.(StringKeyValue)
		j := b.(StringKeyValue)
		return StringKV(i.Key, resolve(i.Key, i.Value, j.Value))
	}
	root := m.root.union(n.root, extractAndResolve, 0, &matches, theCopier)
	return StringMap{root: root, count: m.Count() + n.Count() - matches}
}

// Update returns a StringMap with key-value pairs from n added or replacing existing
// keys.
func (m StringMap) Update(n StringMap) StringMap {
	f := useRHS
	if m.Count() >= n.Count() {
		m, n = n, m
		f = useLHS
	}
	matches := 0
	root := m.root.union(n.root, f, 0, &matches, theCopier)
	return StringMap{root: root, count: m.Count() + n.Count() - matches}
}

// Hash computes a hash val for s.
func (m StringMap) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(uintptr(3167960924819262823&uint64(^uintptr(0))), seed)
	for i := m.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), hash.Interface(i.Key(), seed))
	}
	return h
}

// Equal returns true iff i is a StringMap with all the same key-value pairs as this
// StringMap.
func (m StringMap) Equal(i interface{}) bool {
	if n, ok := i.(StringMap); ok {
		c := newCloner(false, m.Count())
		return m.root.equal(n.root, func(a, b interface{}) bool {
			kva := a.(StringKeyValue)
			kvb := b.(StringKeyValue)
			return Equal(kva.Key, kvb.Key) && Equal(kva.Value, kvb.Value)
		}, 0, c)
	}
	return false
}

// String returns a string representatio of the StringMap.
func (m StringMap) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the StringMap into state.
func (m StringMap) Format(state fmt.State, _ rune) {
	state.Write([]byte("("))
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			state.Write([]byte(", "))
		}
		fmt.Fprintf(state, "%v: %v", i.Key(), i.Value())
	}
	state.Write([]byte(")"))
}

// Range returns a StringMapIterator over the StringMap.
func (m StringMap) Range() *StringMapIterator {
	return &StringMapIterator{i: m.root.iterator(m.count)}
}

// MarshalJSON implements json.Marshaler.
func (m StringMap) MarshalJSON() ([]byte, error) {
	proxy := map[string]interface{}{}
	for i := m.Range(); i.Next(); {
		proxy[i.Key()] = i.Value()
	}
	data, err := json.Marshal(proxy)
	return data, errorsWrap(err, 0)
}

// StringMapIterator provides for iterating over a StringMap.
type StringMapIterator struct {
	i  Iterator
	kv StringKeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *StringMapIterator) Next() bool {
	if i.i.Next() {
		var ok bool
		i.kv, ok = i.i.Value().(StringKeyValue)
		if !ok {
			panic(fmt.Sprintf("Unexpected type: %T", i.i.Value()))
		}
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *StringMapIterator) Key() string {
	return i.kv.Key
}

// Value returns the value for the current entry.
func (i *StringMapIterator) Value() interface{} {
	return i.kv.Value
}

// Entry returns the current key-value pair as two return values.
func (i *StringMapIterator) Entry() (key string, value interface{}) {
	return i.kv.Key, i.kv.Value
}
