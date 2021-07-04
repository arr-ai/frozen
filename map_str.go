package frozen

import (
	"encoding/json"
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fu"
	"github.com/arr-ai/frozen/internal/iterator/skvi"
	"github.com/arr-ai/frozen/internal/tree/skvt"
	"github.com/arr-ai/frozen/internal/value"
	"github.com/arr-ai/frozen/pkg/kv/skv"
)

// StringMapBuilder provides a more efficient way to build Maps incrementally.
type StringMapBuilder struct {
	tb skvt.Builder
}

func NewStringMapBuilder(capacity int) *StringMapBuilder {
	return &StringMapBuilder{tb: *skvt.NewBuilder(capacity)}
}

// Count returns the number of entries in the StringMap under construction.
func (b *StringMapBuilder) Count() int {
	return b.tb.Count()
}

// Put adds or changes an entry into the StringMap under construction.
func (b *StringMapBuilder) Put(key string, value interface{}) {
	b.tb.Add(skvt.DefaultNPKeyCombineArgs, skv.KV(key, value))
}

// Remove removes an entry from the StringMap under construction.
func (b *StringMapBuilder) Remove(key string) {
	b.tb.Remove(skvt.DefaultNPKeyEqArgs, skv.KV(key, nil))
}

func (b *StringMapBuilder) Has(v string) bool {
	_, has := b.Get(v)
	return has
}

// Get returns the value for key from the StringMap under construction or false if
// not found.
func (b *StringMapBuilder) Get(key string) (interface{}, bool) {
	if entry := b.tb.Get(skvt.DefaultNPKeyEqArgs, skv.KV(key, nil)); entry != nil {
		return entry.Value, true
	}
	return nil, false
}

// Finish returns a StringMap containing all entries added since the StringMapBuilder was
// initialised or the last call to Finish.
func (b *StringMapBuilder) Finish() StringMap {
	return newStringMap(b.tb.Finish())
}

func StringKeyEqual(a, b interface{}) bool {
	return value.Equal(a.(KeyValue).Key, b.(KeyValue).Key)
}

// StringMap maps keys to values. The zero value is the empty StringMap.
type StringMap struct {
	tree skvt.Tree
}

var _ value.Key = StringMap{}

func newStringMap(tree skvt.Tree) StringMap {
	return StringMap{tree: tree}
}

// NewStringMap creates a new StringMap with kvs as keys and values.
func NewStringMap(kvs ...skv.KeyValue) StringMap {
	var b StringMapBuilder
	for _, skv := range kvs {
		b.Put(skv.Key, skv.Value)
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

// NewStringMapFromGoMap takes a map[interface{}]interface{} and returns a frozen StringMap from it.
func NewStringMapFromGoMap(m map[string]interface{}) StringMap {
	mb := NewStringMapBuilder(len(m))
	for k, v := range m {
		mb.Put(k, v)
	}
	return mb.Finish()
}

// IsEmpty returns true if the StringMap has no entries.
func (m StringMap) IsEmpty() bool {
	return m.tree.Count() == 0
}

// Count returns the number of entries in the StringMap.
func (m StringMap) Count() int {
	return m.tree.Count()
}

// Any returns an arbitrary entry from the StringMap.
func (m StringMap) Any() (key, value interface{}) {
	for i := m.Range(); i.Next(); {
		return i.Entry()
	}
	panic("empty map")
}

// With returns a new StringMap with key associated with val and all other keys
// retained from m.
func (m StringMap) With(key string, val interface{}) StringMap {
	kv := skv.KV(key, val)
	return newStringMap(m.tree.With(skvt.DefaultNPKeyCombineArgs, kv))
}

// Without returns a new StringMap with all keys retained from m except the elements
// of keys.
func (m StringMap) Without(keys Set) StringMap {
	args := skvt.NewEqArgs(
		m.tree.Gauge(), skvt.KeyEqual, skvt.KeyHash, skvt.KeyHash)
	for i := keys.Range(); i.Next(); {
		m.tree = m.tree.Without(args, skv.KV(i.Value().(string), nil))
	}
	return m
	// TODO: Reinstate parallelisable implementation below.
	// return newStringMap(m.tree.Difference(args, keys.tree))
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
	return m.tree.Get(skvt.DefaultNPKeyEqArgs, skv.KV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m StringMap) Get(key string) (interface{}, bool) {
	if skv := m.tree.Get(skvt.DefaultNPKeyEqArgs, skv.KV(key, nil)); skv != nil {
		return skv.Value, true
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
func (m StringMap) StringMap(f func(key, val interface{}) interface{}) StringMap {
	var b StringMapBuilder
	for i := m.Range(); i.Next(); {
		key, val := i.Entry()
		b.Put(key, f(key, val))
	}
	return b.Finish()
}

// Reduce returns the result of applying f to each key-value pair on the StringMap.
// The result of each call is used as the acc argument for the next element.
func (m StringMap) Reduce(f func(acc, key, val interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

func (m StringMap) EqArgs() *skvt.EqArgs {
	return skvt.NewEqArgs(depth.NewGauge(m.Count()), skvt.KeyEqual, skvt.KeyHash, skvt.KeyHash)
}

// Merge returns a map from the merging between two maps, should there be a key overlap,
// the value that corresponds to key will be replaced by the value resulted from the
// provided resolve function.
func (m StringMap) Merge(n StringMap, resolve func(key string, a, b interface{}) interface{}) StringMap {
	extractAndResolve := func(a, b skv.KeyValue) skv.KeyValue {
		return skv.KV(a.Key, resolve(a.Key, a.Value, b.Value))
	}
	args := skvt.NewCombineArgs(m.EqArgs(), extractAndResolve)
	return newStringMap(m.tree.Combine(args, n.tree))
}

// Update returns a StringMap with key-value pairs from n added or replacing existing
// keys.
func (m StringMap) Update(n StringMap) StringMap {
	f := skvt.UseRHS
	if m.Count() > n.Count() {
		m, n = n, m
		f = skvt.UseLHS
	}
	args := skvt.NewCombineArgs(m.EqArgs(), f)
	return newStringMap(m.tree.Combine(args, n.tree))
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
		args := skvt.NewEqArgs(
			depth.NewGauge(m.Count()),
			skv.KeyValueEqual,
			skvt.KeyHash,
			skvt.KeyHash,
		)
		return m.tree.Equal(args, n.tree)
	}
	return false
}

// String returns a string representatio of the StringMap.
func (m StringMap) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the StringMap into state.
func (m StringMap) Format(f fmt.State, verb rune) {
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

// Range returns a StringMapIterator over the StringMap.
func (m StringMap) Range() *StringMapIterator {
	return &StringMapIterator{i: m.tree.Iterator()}
}

// MarshalJSON implements json.Marshaler.
func (m StringMap) MarshalJSON() ([]byte, error) {
	proxy := map[string]interface{}{}
	for i := m.Range(); i.Next(); {
		proxy[i.Key()] = i.Value()
	}
	data, err := json.Marshal(proxy)
	return data, errors.Wrap(err, 0)
}

// StringMapIterator provides for iterating over a StringMap.
type StringMapIterator struct {
	i   skvi.Iterator
	skv skv.KeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *StringMapIterator) Next() bool {
	if i.i.Next() {
		i.skv = i.i.Value()
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *StringMapIterator) Key() string {
	return i.skv.Key
}

// Value returns the value for the current entry.
func (i *StringMapIterator) Value() interface{} {
	return i.skv.Value
}

// Entry returns the current key-value pair as two return values.
func (i *StringMapIterator) Entry() (key string, value interface{}) {
	return i.skv.Key, i.skv.Value
}
