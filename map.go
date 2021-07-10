package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
	"github.com/arr-ai/frozen/internal/iterator/kvi"
	"github.com/arr-ai/frozen/internal/pkg/debug"
	"github.com/arr-ai/frozen/internal/tree/kvt"
	"github.com/arr-ai/frozen/internal/value"
	"github.com/arr-ai/frozen/pkg/kv"
)

func KeyEqual(a, b interface{}) bool {
	return value.Equal(a.(KeyValue).Key, b.(KeyValue).Key)
}

// Map maps keys to values. The zero value is the empty Map.
type Map struct {
	tree kvt.Tree
}

var _ value.Key = Map{}

func newMap(tree kvt.Tree) Map {
	return Map{tree: tree}
}

// NewMap creates a new Map with kvs as keys and values.
func NewMap(kvs ...kv.KeyValue) Map {
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
	return m.tree.Count() == 0
}

// Count returns the number of entries in the Map.
func (m Map) Count() int {
	return m.tree.Count()
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
	return newMap(m.tree.With(kvt.DefaultNPKeyCombineArgs, kv))
}

// Without returns a new Map with all keys retained from m except the elements
// of keys.
func (m Map) Without(keys Set) Map {
	args := kvt.NewEqArgs(
		m.tree.Gauge(), kvt.KeyEqual, kvt.KeyHash, kvt.KeyHash)
	for i := keys.Range(); i.Next(); {
		m.tree = m.tree.Without(args, KV(i.Value(), nil))
	}
	return m
	// TODO: Reinstate parallelisable implementation below.
	// return newMap(m.tree.Difference(args, keys.tree))
}

// Without2 shoves keys into a Set and calls m.Without.
func (m Map) Without2(keys ...interface{}) Map {
	var sb SetBuilder
	for _, key := range keys {
		sb.Add(key)
	}
	return m.Without(sb.Finish())
}

// Has returns true iff the key exists in the map.
func (m Map) Has(key interface{}) bool {
	return m.tree.Get(kvt.DefaultNPKeyEqArgs, KV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	if kv := m.tree.Get(kvt.DefaultNPKeyEqArgs, KV(key, nil)); kv != nil {
		return kv.Value, true
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

func (m Map) EqArgs() *kvt.EqArgs {
	return kvt.NewEqArgs(
		depth.NewGauge(m.Count()),
		kvt.KeyEqual,
		kvt.KeyHash,
		kvt.KeyHash,
	)
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m Map) Merge(n Map, resolve func(key, a, b interface{}) interface{}) Map {
	extractAndResolve := func(a, b KeyValue) KeyValue {
		return KV(a.Key, resolve(a.Key, a.Value, b.Value))
	}
	args := kvt.NewCombineArgs(m.EqArgs(), extractAndResolve)
	return newMap(m.tree.Combine(args, n.tree))
}

// Update returns a Map with key-value pairs from n added or replacing existing
// keys.
func (m Map) Update(n Map) Map {
	f := kvt.UseRHS
	if m.Count() > n.Count() {
		m, n = n, m
		f = kvt.UseLHS
	}
	args := kvt.NewCombineArgs(m.EqArgs(), f)
	return newMap(m.tree.Combine(args, n.tree))
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
		args := kvt.NewEqArgs(
			depth.NewGauge(m.Count()),
			kv.KeyValueEqual,
			kvt.KeyHash,
			kvt.KeyHash,
		)
		return m.tree.Equal(args, n.tree)
	}
	return false
}

// String returns a string representatio of the Map.
func (m Map) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map into state.
func (m Map) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			fu.WriteString(f, ", ")
		}
		fu.Format(i.Key(), f, verb)
		fu.WriteString(f, ": ")
		fu.Format(i.Value(), f, verb)
	}
	fu.WriteString(f, ")")
}

// Range returns a MapIterator over the Map.
func (m Map) Range() *MapIterator {
	return &MapIterator{i: m.tree.Iterator()}
}

// DebugReport is for internal use.
func (m Map) DebugReport(debug.Tag) string {
	return m.tree.String()
}

// MapIterator provides for iterating over a Map.
type MapIterator struct {
	i  kvi.Iterator
	kv KeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *MapIterator) Next() bool {
	if i.i.Next() {
		i.kv = i.i.Value()
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
