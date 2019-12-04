package frozen

import (
	"fmt"

	"github.com/marcelocantos/frozen/pkg/value"
)

type KeyValue struct {
	Key, Value interface{}
}

func KV(key, val interface{}) KeyValue {
	return KeyValue{Key: key, Value: val}
}

func (kv KeyValue) Hash() uint64 {
	return value.Hash(kv.Key)
}

func (kv KeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(KeyValue); ok {
		return value.Equal(kv.Key, kv2.Key)
	}
	return false
}

// Map maps keys to values. The zero value is the empty Map.
type Map struct {
	n     *node
	count int
}

var _ value.Key = Set{}

func NewMap(kvs ...KeyValue) Map {
	return Map{}.WithKVs(kvs...)
}

func (m Map) IsEmpty() bool {
	return m.n == nil
}

func (m Map) Count() int {
	return m.count
}

// Put returns a new Map with key associated with val and all other keys
// retained from m.
func (m Map) With(key, val interface{}) Map {
	n, old := m.n.put(KV(key, val))
	count := m.count
	if old == nil {
		count++
	}
	return Map{n: n, count: count}
}

// Put returns a new Map with key associated with val and all other keys
// retained from m.
func (m Map) WithKVs(kvs ...KeyValue) Map {
	for _, kv := range kvs {
		m = m.With(kv.Key, kv.Value)
	}
	return m
}

// Put returns a new Map with all keys retained from m except key.
func (m Map) Without(keys Set) Map {
	n := m.n
	count := m.count
	for k := keys.Range(); k.Next(); {
		var old interface{}
		n, old = n.delete(KV(k.Value(), nil))
		if old != nil {
			count--
		}
	}
	return Map{n: n, count: count}
}

// Get returns the value associated with key in m and true iff the key was
// found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	if kv := m.n.get(KV(key, nil)); kv != nil {
		return kv.(KeyValue).Value, true
	}
	return nil, false
}

func (m Map) MustGet(key interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

func (m Map) ValueElse(key, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

func (m Map) ValueElseFunc(key interface{}, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

func (m Map) Keys() Set {
	return m.Reduce(func(acc, key, _ interface{}) interface{} {
		return acc.(Set).With(key)
	}, Set{}).(Set)
}

func (m Map) Values() Set {
	return m.Reduce(func(acc, _, val interface{}) interface{} {
		return acc.(Set).With(val)
	}, Set{}).(Set)
}

func (m Map) Project(keys Set) Map {
	return m.Where(func(key, val interface{}) bool {
		return keys.Has(key)
	})
}

func (m Map) Where(pred func(key, val interface{}) bool) Map {
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		if pred(key, val) {
			return acc.(Map).With(key, val)
		}
		return acc
	}, NewMap()).(Map)
}

func (m Map) Map(f func(key, val interface{}) interface{}) Map {
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		return acc.(Map).With(key, f(key, val))
	}, NewMap()).(Map)
}

func (m Map) Reduce(f func(acc, key, val interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

func (m Map) Update(n Map) Map {
	return n.Reduce(func(acc, key, val interface{}) interface{} {
		return acc.(Map).With(key, val)
	}, m).(Map)
}

// Hash computes a hash val for s.
func (m Map) Hash() uint64 {
	var h uint64 = 3167960924819262823
	for i := m.Range(); i.Next(); {
		h ^= 12012876008735959943*value.Hash(i.Key()) + value.Hash(i.Value())
	}
	return h
}

func (m Map) Equal(i interface{}) bool {
	if n, ok := i.(Map); ok {
		if m.Hash() != n.Hash() {
			return false
		}
		for i := m.Range(); i.Next(); {
			if val, has := n.Get(i.Key()); has {
				if !value.Equal(val, i.Value()) {
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
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			f.Write([]byte(", "))
		}
		fmt.Fprintf(f, "%v: %v", i.Key(), i.Value())
	}
	f.Write([]byte("}"))
}

func (m Map) Range() *MapIter {
	return &MapIter{i: m.n.iterator()}
}

type MapIter struct {
	i  *nodeIter
	kv KeyValue
}

func (i *MapIter) Next() bool {
	if i.i.next() {
		var ok bool
		i.kv, ok = i.i.elem.(KeyValue)
		if !ok {
			panic(fmt.Sprintf("Unexpected type: %T", i.i.elem))
		}
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
