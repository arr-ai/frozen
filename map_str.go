package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"
)

var (
	defaultNPStringKeyEqArgs      = newDefaultStringKeyEqArgs(nonParallel)
	defaultNPStringKeyCombineArgs = newCombineArgs(defaultNPStringKeyEqArgs, useRHS)

	stringKeyHash = stringKeyHasher(hash.String)
)

func stringKeyHasher(hash func(v string, seed uintptr) uintptr) func(v interface{}, seed uintptr) uintptr {
	return func(v interface{}, seed uintptr) uintptr {
		return hash(v.(StringKeyValue).Key, seed)
	}
}

func newDefaultStringKeyEqArgs(gauge parallelDepthGauge) *eqArgs {
	return newEqArgs(gauge, StringKeyEqual, stringKeyHash, stringKeyHash)
}

// StringKeyValue represents a key-value pair for insertion into a StringMap.
type StringKeyValue struct {
	Key   string
	Value interface{}
}

// StringKV creates a StringKeyValue.
func StringKV(key string, val interface{}) StringKeyValue {
	return StringKeyValue{Key: key, Value: val}
}

// Hash computes a hash for a StringKeyValue.
func (kv StringKeyValue) Hash(seed uintptr) uintptr {
	return hash.String(kv.Key, seed)
}

func StringKeyValueEqual(a, b interface{}) bool {
	i := a.(StringKeyValue)
	j := b.(StringKeyValue)
	return Equal(i.Key, j.Key) && Equal(i.Value, j.Value)
}

// StringMap returns a string representation of a StringKeyValue.
func (kv StringKeyValue) StringMap() string {
	return fmt.Sprintf("%#v:%#v", kv.Key, kv.Value)
}

func StringKeyEqual(a, b interface{}) bool {
	return Equal(a.(StringKeyValue).Key, b.(StringKeyValue).Key)
}

// StringMap maps keys to values. The zero value is the empty StringMap.
type StringMap struct {
	root tree
}

var _ Key = StringMap{}

func newStringMap(root tree) StringMap {
	return StringMap{root: root}
}

// NewStringMap creates a new StringMap with kvs as keys and values.
func NewStringMap(kvs ...StringKeyValue) StringMap {
	var b StringMapBuilder
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewStringMapFromKeys creates a new StringMap in which values are computed from keys.
func NewStringMapFromKeys(keys Set, f func(key string) interface{}) StringMap {
	var b StringMapBuilder
	for i := keys.Range(); i.Next(); {
		val := i.Value().(string)
		b.Put(val, f(val))
	}
	return b.Finish()
}

// NewStringMapFromGoMap takes a map[string]interface{} and returns a frozen StringMap from it.
func NewStringMapFromGoMap(m map[string]interface{}) StringMap {
	mb := NewStringMapBuilder(len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the StringMap has no entries.
func (m StringMap) IsEmpty() bool {
	return m.root.count == 0
}

// Count returns the number of entries in the StringMap.
func (m StringMap) Count() int {
	return m.root.count
}

// Any returns an arbitrary entry from the StringMap.
func (m StringMap) Any() (key string, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty map")
}

// With returns a new StringMap with key associated with val and all other keys
// retained from m.
func (m StringMap) With(key string, val interface{}) StringMap {
	kv := StringKV(key, val)
	return newStringMap(m.root.With(defaultNPStringKeyCombineArgs, kv))
}

// Without returns a new StringMap with all keys retained from m except the elements
// of keys.
func (m StringMap) Without(keys Set) StringMap {
	args := newEqArgs(
		m.root.Gauge(),
		func(a, b interface{}) bool {
			return Equal(a.(StringKeyValue).Key, b)
		},
		stringKeyHash,
		hash.Interface)
	return newStringMap(m.root.Difference(args, keys.root))
}

// Without2 shoves keys into a Set and calls m.Without.
func (m StringMap) Without2(keys ...string) StringMap {
	var sb SetBuilder
	for _, key := range keys {
		sb.Add(key)
	}
	return m.Without(sb.Finish())
}

// Has returns true iff the key exists in the map.
func (m StringMap) Has(key string) bool {
	return m.root.Get(defaultNPStringKeyEqArgs, StringKV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m StringMap) Get(key string) (interface{}, bool) {
	if kv := m.root.Get(defaultNPStringKeyEqArgs, StringKV(key, nil)); kv != nil {
		return (*kv).(StringKeyValue).Value, true
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

type StringMapReduceFunc func(acc interface{}, key string, val interface{}) interface{}

// Reduce returns the result of applying f to each key-value pair on the StringMap.
// The result of each call is used as the acc argument for the next element.
func (m StringMap) Reduce(f StringMapReduceFunc, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

func (m StringMap) eqArgs() *eqArgs {
	return newEqArgs(
		newParallelDepthGauge(m.Count()),
		StringKeyEqual,
		stringKeyHash,
		stringKeyHash,
	)
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m StringMap) Merge(n StringMap, resolve func(key string, a, b interface{}) interface{}) StringMap {
	extractAndResolve := func(a, b interface{}) interface{} {
		i := a.(StringKeyValue)
		j := b.(StringKeyValue)
		return StringKV(i.Key, resolve(i.Key, i.Value, j.Value))
	}
	args := newCombineArgs(m.eqArgs(), extractAndResolve)
	return newStringMap(m.root.Combine(args, n.root))
}

// Update returns a StringMap with key-value pairs from n added or replacing existing
// keys.
func (m StringMap) Update(n StringMap) StringMap {
	f := useRHS
	if m.Count() > n.Count() {
		m, n = n, m
		f = useLHS
	}
	return newStringMap(m.root.Combine(newCombineArgs(m.eqArgs(), f), n.root))
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
		args := newEqArgs(
			newParallelDepthGauge(m.Count()),
			StringKeyValueEqual,
			hash.Interface,
			hash.Interface,
		)
		return m.root.Equal(args, n.root)
	}
	return false
}

// StringMap returns a string representatio of the StringMap.
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
	return &StringMapIterator{i: m.root.Iterator()}
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
