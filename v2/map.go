package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/pkg/debug"
	"github.com/arr-ai/frozen/v2/internal/pkg/depth"
	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
	"github.com/arr-ai/frozen/v2/internal/pkg/tree"
	"github.com/arr-ai/frozen/v2/internal/pkg/value"
	"github.com/arr-ai/frozen/v2/pkg/kv"
)

func defaultMapNPKeyEqArgs[K, V any]() *tree.EqArgs[kv.KeyValue[K, V]] {
	return newDefaultMapKeyEqArgs[K, V](depth.NonParallel)
}

func defaultMapNPKeyCombineArgs[K, V any]() *tree.CombineArgs[kv.KeyValue[K, V]] {
	return tree.NewCombineArgs(defaultMapNPKeyEqArgs[K, V](), tree.UseRHS[kv.KeyValue[K, V]])
}

func newDefaultMapKeyEqArgs[K, V any](gauge depth.Gauge) *tree.EqArgs[kv.KeyValue[K, V]] {
	return tree.NewEqArgs(gauge, mapElementEqual[K, V], mapHashValue[K, V], mapHashValue[K, V])
}

func mapElementEqual[K, V any](a, b kv.KeyValue[K, V]) bool {
	return value.Equal(a.Key, b.Key)
}

func mapHashValue[K, V any](v kv.KeyValue[K, V], seed uintptr) uintptr {
	return hash.Interface(v.Key, seed)
}

// Map maps keys to values. The zero value is the empty Map.
type Map[K any, V any] struct {
	tree tree.Tree[kv.KeyValue[K, V]]
}

func newMap[K any, V any](tree tree.Tree[kv.KeyValue[K, V]]) Map[K, V] {
	return Map[K, V]{tree: tree}
}

// NewMap creates a new Map with kvs as keys and values.
func NewMap[K any, V any](kvs ...kv.KeyValue[K, V]) Map[K, V] {
	b := NewMapBuilder[K, V](len(kvs))
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// NewMapFromKeys creates a new Map in which values are computed from keys.
func NewMapFromKeys[K any, V any](keys Set[K], f func(key K) V) Map[K, V] {
	b := NewMapBuilder[K, V](keys.Count())
	for i := keys.Range(); i.Next(); {
		val := i.Value()
		b.Put(val, f(val))
	}
	return b.Finish()
}

// NewMapFromGoMap takes a map[K]V and returns a frozen Map from it.
func NewMapFromGoMap[K comparable, V any](m map[K]V) Map[K, V] {
	mb := NewMapBuilder[K, V](len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the Map has no entries.
func (m Map[K, V]) IsEmpty() bool {
	return m.tree.Count() == 0
}

// Count returns the number of entries in the Map.
func (m Map[K, V]) Count() int {
	return m.tree.Count()
}

// Any returns an arbitrary entry from the Map.
func (m Map[K, V]) Any() (key K, value V) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty map")
}

// With returns a new Map with key associated with val and all other keys
// retained from m.
func (m Map[K, V]) With(key K, val V) Map[K, V] {
	kval := kv.KV(key, val)
	return newMap(m.tree.With(
		defaultMapNPKeyCombineArgs[K, V](),
		kval,
	))
}

// keyHash hashes using the KeyValue's own key.
func keyHash[K, V any](kv kv.KeyValue[K, V], seed uintptr) uintptr {
	return kv.Hash(seed)
}

// Without returns a new Map with all keys retained from m except the elements
// of keys.
func (m Map[K, V]) Without(keys Set[K]) Map[K, V] {
	for i := keys.Range(); i.Next(); {
		var zarro V
		m.tree = m.tree.Without(defaultMapNPKeyEqArgs[K, V](), kv.KV(i.Value(), zarro))
	}
	return m
	// TODO: Reinstate parallelisable implementation below.
	// return newMap(m.tree.Difference(args, keys.tree))
}

// Without2 shoves keys into a Set and calls m.Without.
func (m Map[K, V]) Without2(keys ...K) Map[K, V] {
	var sb SetBuilder[K]
	for _, key := range keys {
		sb.Add(key)
	}
	return m.Without(sb.Finish())
}

// Has returns true iff the key exists in the map.
func (m Map[K, V]) Has(key K) bool {
	_, has := m.Get(key)
	return has
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map[K, V]) Get(key K) (V, bool) {
	var zarro V
	if kv := m.tree.Get(defaultMapNPKeyEqArgs[K, V](), kv.KV(key, zarro)); kv != nil {
		return kv.Value, true
	}
	return zarro, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m Map[K, V]) MustGet(key K) V {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m Map[K, V]) GetElse(key K, deflt V) V {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m Map[K, V]) GetElseFunc(key K, deflt func() V) V {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the Map.
func (m Map[K, V]) Keys() Set[K] {
	var b SetBuilder[K]
	for i := m.Range(); i.Next(); {
		b.Add(i.Key())
	}
	return b.Finish()
}

// Values returns a Set with all the Values in the Map.
func (m Map[K, V]) Values() Set[V] {
	var b SetBuilder[V]
	for i := m.Range(); i.Next(); {
		b.Add(i.Value())
	}
	return b.Finish()
}

// Project returns a Map with only keys included from this Map.
func (m Map[K, V]) Project(keys Set[K]) Map[K, V] {
	return m.Where(func(key K, val V) bool {
		return keys.Has(key)
	})
}

// Where returns a Map with only key-value pairs satisfying pred.
func (m Map[K, V]) Where(pred func(key K, val V) bool) Map[K, V] {
	var b MapBuilder[K, V]
	for i := m.Range(); i.Next(); {
		if key, val := i.Entry(); pred(key, val) {
			b.Put(key, val)
		}
	}
	return b.Finish()
}

// Map returns a Map with keys from this Map, but the values replaced by the
// result of calling f.
func MapMap[K, V, U any](m Map[K, V], f func(key K, val V) U) Map[K, U] {
	var b MapBuilder[K, U]
	for i := m.Range(); i.Next(); {
		key, val := i.Entry()
		b.Put(key, f(key, val))
	}
	return b.Finish()
}

func (m Map[K, V]) EqArgs() *tree.EqArgs[kv.KeyValue[K, V]] {
	return tree.NewEqArgs(
		depth.NewGauge(m.Count()),
		kv.KeyEqual[K, V],
		keyHash[K, V],
		keyHash[K, V],
	)
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m Map[K, V]) Merge(n Map[K, V], resolve func(key K, a, b V) V) Map[K, V] {
	extractAndResolve := func(a, b kv.KeyValue[K, V]) kv.KeyValue[K, V] {
		return kv.KV(a.Key, resolve(a.Key, a.Value, b.Value))
	}
	args := tree.NewCombineArgs(m.EqArgs(), extractAndResolve)
	return newMap(m.tree.Combine(args, n.tree))
}

// Update returns a Map with key-value pairs from n added or replacing existing
// keys.
func (m Map[K, V]) Update(n Map[K, V]) Map[K, V] {
	f := tree.UseRHS[kv.KeyValue[K, V]]
	if m.Count() > n.Count() {
		m, n = n, m
		f = tree.UseLHS[kv.KeyValue[K, V]]
	}
	args := tree.NewCombineArgs(m.EqArgs(), f)
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

// Equal returns true iff i is a Map with all the same key-value pairs as this
// Map.
func (m Map[K, V]) Equal(n Map[K, V]) bool {
	args := tree.NewEqArgs(
		depth.NewGauge(m.Count()),
		kv.KeyValueEqual[K, V],
		keyHash[K, V],
		keyHash[K, V],
	)
	return m.tree.Equal(args, n.tree)
}

// String returns a string representatio of the Map.
func (m Map[K, V]) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map into state.
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

// Range returns a MapIterator over the Map.
func (m Map[K, V]) Range() MapIterator[K, V] {
	return MapIterator[K, V]{i: m.tree.Iterator()}
}

// DebugReport is for internal use.
func (m Map[K, V]) DebugReport(debug.Tag) string {
	return m.tree.String()
}

// MapIterator provides for iterating over a Map.
type MapIterator[K any, V any] struct {
	i  iterator.Iterator[kv.KeyValue[K, V]]
	kv kv.KeyValue[K, V]
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *MapIterator[K, V]) Next() bool {
	if i.i.Next() {
		i.kv = i.i.Value()
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *MapIterator[K, V]) Key() K {
	return i.kv.Key
}

// Value returns the value for the current entry.
func (i *MapIterator[K, V]) Value() V {
	return i.kv.Value
}

// Entry returns the current key-value pair as two return values.
func (i *MapIterator[K, V]) Entry() (key K, value V) {
	return i.kv.Key, i.kv.Value
}
