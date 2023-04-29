package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/pkg/debug"
	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/tree"
)

// Map maps keys to values. The zero value is the empty Map.
type Map[K any, V any] struct {
	tree tree.Tree[mapEntry[K, V]]
}

func newMap[K any, V any](tree tree.Tree[mapEntry[K, V]]) Map[K, V] {
	return Map[K, V]{tree: tree}
}

// NewMap creates a new Map with kvs as keys and values.
func NewMap[K any, V any](kvs ...KeyValue[K, V]) Map[K, V] {
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
	kval := newMapEntry(key, val)
	return newMap(m.tree.With(kval))
}

// Without returns a new Map with all keys retained from m except the elements
// of keys.
func (m Map[K, V]) Without(key K) Map[K, V] {
	m.tree = m.tree.Without(newMapKey[K, V](key))
	return m
}

// Has returns true iff the key exists in the map.
func (m Map[K, V]) Has(key K) bool {
	_, has := m.Get(key)
	return has
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map[K, V]) Get(key K) (_ V, _ bool) {
	if kv := m.tree.Get(newMapKey[K, V](key)); kv != nil {
		return kv.Value, true
	}
	return
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
func (m Map[K, V]) Project(keys ...K) Map[K, V] {
	var mb MapBuilder[K, V]
	for _, k := range keys {
		if v, has := m.Get(k); has {
			mb.Put(k, v)
		}
	}
	return mb.Finish()
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

func (m Map[K, V]) EqArgs() *tree.EqArgs[mapEntry[K, V]] {
	return tree.NewEqArgs(
		depth.NewGauge(m.Count()),
		mapEntryEqual[K, V],
		mapEntryHash[K, V],
	)
}

func (m Map[K, V]) EqKeyArgs() *tree.EqArgs[mapEntry[K, V]] {
	return tree.NewEqArgs(
		depth.NewGauge(m.Count()),
		mapEntryKeyEqual[K, V],
		mapEntryHash[K, V],
	)
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m Map[K, V]) Merge(n Map[K, V], resolve func(key K, a, b V) V) Map[K, V] {
	extractAndResolve := func(a, b mapEntry[K, V]) mapEntry[K, V] {
		return newMapEntry(a.Key, resolve(a.Key, a.Value, b.Value))
	}
	args := tree.NewCombineArgs(m.EqKeyArgs(), extractAndResolve)
	return newMap(m.tree.Combine(args, n.tree))
}

// Update returns a Map with key-value pairs from n added or replacing existing
// keys.
func (m Map[K, V]) Update(n Map[K, V]) Map[K, V] {
	f := tree.UseRHS[mapEntry[K, V]]
	if m.Count() > n.Count() {
		m, n = n, m
		f = tree.UseLHS[mapEntry[K, V]]
	}
	args := tree.NewCombineArgs(m.EqKeyArgs(), f)
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
		mapEntryEqual[K, V],
		mapEntryHash[K, V],
	)
	return m.tree.Equal(args, n.tree)
}

// Same returns true iff a is a Map and m and n have all the same key-values.
func (m Map[K, V]) Same(a any) bool {
	n, is := a.(Map[K, V])
	return is && m.Equal(n)
}

// String returns a string representatio of the Map.
func (m Map[K, V]) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map into state.
func (m Map[K, V]) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		f.Write([]byte{'M'})
		m.tree.Format(f, verb)
		return
	}

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
	i  Iterator[mapEntry[K, V]]
	kv mapEntry[K, V]
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

// ToMap transforms a Map[K, V] to a map[K]V. K must be comparable.
func ToMap[K comparable, V any](m Map[K, V]) map[K]V {
	result := make(map[K]V, m.Count())
	for r := m.Range(); r.Next(); {
		k, v := r.Entry()
		result[k] = v
	}
	return result
}
