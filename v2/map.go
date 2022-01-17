package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/depth"
	"github.com/arr-ai/frozen/v2/internal/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/debug"
	"github.com/arr-ai/frozen/v2/internal/value"
	"github.com/arr-ai/frozen/v2/pkg/kv"
)

func KeyEqual[K, V any](a, b kv.KeyValue[K, T]) bool {
	return value.Equal(a.Key, b.Key)
}

// Map[K, V] maps keys to values. The zero value is the empty Map[K, V].
type Map[K, V any] struct {
	tree kv.Tree[KeyValue[K, V]]
}

var _ value.Key = Map[K, V]{}

func newMap[K, V any](tree kvt.Tree[KeyValue[K, V]]) Map[K, V] {
	return Map[K, V]{tree: tree}
}

// NewMap creates a new Map[K, V] with kvs as keys and values.
func NewMap[K, V any](kvs ...kv.KeyValue[K, V]) Map[K, V] {
	b := NewMapBuilder[K, V](len(kvs))
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewMapFromKeys creates a new Map[K, V] in which values are computed from keys.
func NewMapFromKeys[K, V any](keys Set[K], f func(key K) V) Map[K, V] {
	b := NewMapBuilder[K, V](keys.Count())
	for i := keys.Range(); i.Next(); {
		val := i.Value()
		b.Put(val, f(val))
	}
	return b.Finish()
}

// NewMapFromGoMap takes a map[K]V and returns a frozen Map[K, V] from it.
func NewMapFromGoMap[K, V any](m map[K]V) Map[K, V] {
	mb := NewMapBuilder[K, V](len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the Map[K, V] has no entries.
func (m Map[K, V]) IsEmpty() bool {
	return m.tree.Count() == 0
}

// Count returns the number of entries in the Map[K, V].
func (m Map[K, V]) Count() int {
	return m.tree.Count()
}

// Any returns an arbitrary entry from the Map[K, V].
func (m Map[K, V]) Any() (key, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty map")
}

// With returns a new Map[K, V] with key associated with val and all other keys
// retained from m.
func (m Map[K, V]) With(key, val interface{}) Map[K, V] {
	kv := KV(key, val)
	return newMap(m.tree.With(kvt.DefaultNPKeyCombineArgs, kv))
}

// Without returns a new Map[K, V] with all keys retained from m except the elements
// of keys.
func (m Map[K, V]) Without(keys Set) Map[K, V] {
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
func (m Map[K, V]) Without2(keys ...interface{}) Map[K, V] {
	var sb SetBuilder
	for _, key := range keys {
		sb.Add(key)
	}
	return m.Without(sb.Finish())
}

// Has returns true iff the key exists in the map.
func (m Map[K, V]) Has(key interface{}) bool {
	return m.tree.Get(kvt.DefaultNPKeyEqArgs, KV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map[K, V]) Get(key interface{}) (interface{}, bool) {
	if kv := m.tree.Get(kvt.DefaultNPKeyEqArgs, KV(key, nil)); kv != nil {
		return kv.Value, true
	}
	return nil, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m Map[K, V]) MustGet(key interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m Map[K, V]) GetElse(key, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m Map[K, V]) GetElseFunc(key interface{}, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the Map[K, V].
func (m Map[K, V]) Keys() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Key())
	}
	return b.Finish()
}

// Values returns a Set with all the Values in the Map[K, V].
func (m Map[K, V]) Values() Set {
	var b SetBuilder
	for i := m.Range(); i.Next(); {
		b.Add(i.Value())
	}
	return b.Finish()
}

// Project returns a Map[K, V] with only keys included from this Map[K, V].
func (m Map[K, V]) Project(keys Set) Map[K, V] {
	return m.Where(func(key, val interface{}) bool {
		return keys.Has(key)
	})
}

// Where returns a Map[K, V] with only key-value pairs satisfying pred.
func (m Map[K, V]) Where(pred func(key, val interface{}) bool) Map[K, V] {
	var b MapBuilder
	for i := m.Range(); i.Next(); {
		if key, val := i.Entry(); pred(key, val) {
			b.Put(key, val)
		}
	}
	return b.Finish()
}

// // Map[K, V] returns a Map[K, V] with keys from this Map[K, V], but the values replaced by the
// // result of calling f.
// func (m Map[K, V]) Map[K, V](f func(key, val interface{}) interface{}) Map[K, V] {
// 	var b MapBuilder
// 	for i := m.Range(); i.Next(); {
// 		key, val := i.Entry()
// 		b.Put(key, f(key, val))
// 	}
// 	return b.Finish()
// }

// // Reduce returns the result of applying f to each key-value pair on the Map[K, V].
// // The result of each call is used as the acc argument for the next element.
// func (m Map[K, V]) Reduce(f func(acc, key, val interface{}) interface{}, acc interface{}) interface{} {
// 	for i := m.Range(); i.Next(); {
// 		acc = f(acc, i.Key(), i.Value())
// 	}
// 	return acc
// }

func (m Map[K, V]) EqArgs() *kvt.EqArgs {
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
func (m Map[K, V]) Merge(n Map[K, V], resolve func(key K, a, b V) V) Map[K, V] {
	extractAndResolve := func(a, b KeyValue) KeyValue {
		return KV(a.Key, resolve(a.Key, a.Value, b.Value))
	}
	args := kv.NewCombineArgs[KeyValue[K, V]](m.EqArgs(), extractAndResolve)
	return newMap[K, V](m.tree.Combine(args, n.tree))
}

// Update returns a Map[K, V] with key-value pairs from n added or replacing existing
// keys.
func (m Map[K, V]) Update(n Map[K, V]) Map[K, V] {
	f := kvt.UseRHS
	if m.Count() > n.Count() {
		m, n = n, m
		f = kvt.UseLHS
	}
	args := kvt.NewCombineArgs(m.EqArgs(), f)
	return newMap(m.tree.Combine(args, n.tree))
}

// Hash computes a hash val for s.
func (m Map[K, V]) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(uintptr(3167960924819262823&uint64(^uintptr(0))), seed)
	for i := m.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), hash.Interface(i.Key(), seed))
	}
	return h
}

// Equal returns true iff i is a Map[K, V] with all the same key-value pairs as this
// Map[K, V].
func (m Map[K, V]) Equal(i interface{}) bool {
	if n, ok := i.(Map[K, V]); ok {
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

// String returns a string representatio of the Map[K, V].
func (m Map[K, V]) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map[K, V] into state.
func (m Map[K, V]) Format(f fmt.State, verb rune) {
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

// Range returns a MapIterator over the Map[K, V].
func (m Map[K, V]) Range() *MapIterator {
	return &MapIterator{i: m.tree.Iterator()}
}

// DebugReport is for internal use.
func (m Map[K, V]) DebugReport(debug.Tag) string {
	return m.tree.String()
}

// MapIterator provides for iterating over a Map[K, V].
type MapIterator[K, V any] struct {
	i  kv.Iterator[KeyValue[K, V]]
	kv KeyValue[K, V]
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
