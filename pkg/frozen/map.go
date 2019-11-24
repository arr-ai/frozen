package frozen

import (
	"fmt"
)

type KeyValue struct {
	Key, Value interface{}
}

func KV(key, value interface{}) KeyValue {
	return KeyValue{Key: key, Value: value}
}

func (kv KeyValue) Hash() uint64 {
	return hash(kv.Key)
}

func (kv KeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(KeyValue); ok {
		return equal(kv.Key, kv2.Key)
	}
	return false
}

// Map maps keys to values. The zero value is the empty Map.
type Map struct {
	t     hamt
	count int
}

func NewMap(kvs ...KeyValue) Map {
	return Map{}.WithKVs(kvs...)
}

func (m Map) hamt() hamt {
	if m.t == nil {
		return empty{}
	}
	return m.t
}

func (m Map) IsEmpty() bool {
	return m.hamt().isEmpty()
}

func (m Map) Count() int {
	return m.count
}

// Put returns a new Map with key associated with value and all other keys
// retained from m.
func (m Map) With(key, value interface{}) Map {
	result, old := m.hamt().put(KV(key, value), newBuffer(m.count))
	count := m.count
	if old == nil {
		count++
	}
	return Map{t: result, count: count}
}

// Put returns a new Map with key associated with value and all other keys
// retained from m.
func (m Map) WithKVs(kvs ...KeyValue) Map {
	for _, kv := range kvs {
		m = m.With(kv.Key, kv.Value)
	}
	return m
}

// Put returns a new Map with all keys retained from m except key.
func (m Map) Without(keys Set) Map {
	result := m.hamt()
	count := m.count
	for k := keys.Range(); k.Next(); {
		var old element
		result, old = result.delete(KV(k.Value(), nil), newBuffer(m.count-1))
		if old != nil {
			count--
		}
	}
	if result == (empty{}) {
		result = nil
	}
	return Map{t: result, count: count}
}

// Get returns the value associated with key in m and true iff the key was
// found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	if kv, has := m.hamt().get(KV(key, nil)); has {
		return kv.(KeyValue).Value, true
	}
	return nil, false
}

func (m Map) MustGet(key interface{}) interface{} {
	if value, has := m.Get(key); has {
		return value
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

func (m Map) ValueElse(key interface{}, deflt interface{}) interface{} {
	if value, has := m.Get(key); has {
		return value
	}
	return deflt
}

func (m Map) ValueElseFunc(key interface{}, deflt func() interface{}) interface{} {
	if value, has := m.Get(key); has {
		return value
	}
	return deflt()
}

func (m Map) Keys() Set {
	return m.Reduce(func(acc, key, _ interface{}) interface{} {
		return acc.(Set).With(key)
	}, Set{}).(Set)
}

func (m Map) Values() Set {
	return m.Reduce(func(acc, _, value interface{}) interface{} {
		return acc.(Set).With(value)
	}, Set{}).(Set)
}

func (m Map) Project(keys Set) Map {
	return m.Where(func(key, value interface{}) bool {
		return keys.Has(key)
	})
}

func (m Map) Where(pred func(key, value interface{}) bool) Map {
	return m.Reduce(func(acc, key, value interface{}) interface{} {
		if pred(key, value) {
			return acc.(Map).With(key, value)
		}
		return acc
	}, NewMap()).(Map)
}

func (m Map) Map(f func(key, value interface{}) interface{}) Map {
	return m.Reduce(func(acc, key, value interface{}) interface{} {
		return acc.(Map).With(key, f(key, value))
	}, NewMap()).(Map)
}

func (m Map) Reduce(f func(acc, key, value interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

func (m Map) Update(n Map) Map {
	return n.Reduce(func(acc, key, value interface{}) interface{} {
		return acc.(Map).With(key, value)
	}, m).(Map)
}

// Hash computes a hash value for s.
func (m Map) Hash() uint64 {
	var h uint64 = 3167960924819262823
	for i := m.Range(); i.Next(); {
		h ^= 12012876008735959943*hash(i.Key()) + hash(i.Value())
	}
	return h
}

func (m Map) Equal(i interface{}) bool {
	if n, ok := i.(Map); ok {
		if m.Hash() != n.Hash() {
			return false
		}
		for i := m.Range(); i.Next(); {
			if value, has := n.Get(i.Key()); has {
				if !equal(value, i.Value()) {
					return false
				}
			} else {
				return false
			}
		}
		for i := n.Range(); i.Next(); {
			if _, has := n.Get(i.Key()); !has {
				return false
			}
		}
		return true
	}
	return false
}

func (m Map) String() string {
	return fmt.Sprintf("%v", m)
}

func (m Map) Format(f fmt.State, _ rune) {
	f.Write([]byte("{"))
	for i := m.Range(); i.Next(); {
		if i.Index() > 0 {
			f.Write([]byte(", "))
		}
		fmt.Fprintf(f, "%v: %v", i.Key(), i.Value())
	}
	f.Write([]byte("}"))
}

func (m Map) Range() *MapIter {
	return &MapIter{i: m.hamt().iterator()}
}

type MapIter struct {
	i  *hamtIter
	kv KeyValue
}

func (i *MapIter) Index() int {
	return i.i.i
}

func (i *MapIter) Next() bool {
	if i.i.next() {
		i.kv = i.i.e.elem.(KeyValue)
		return true
	}
	return false
}

func (i *MapIter) Key() interface{} {
	return i.kv.Key
}

func (i *MapIter) Value() interface{} {
	return i.kv.Value
}
