package frozen

import (
	"fmt"

	"github.com/marcelocantos/frozen/pkg/value"
)

// Set holds a set of values. The zero value is the empty Set.
type Set struct {
	n     *node
	count int
}

var _ value.Key = Set{}

func NewSet(values ...interface{}) Set {
	return Set{}.With(values...)
}

func (s Set) IsEmpty() bool {
	return s.n == nil
}

func (s Set) Count() int {
	return s.count
}

// With returns a new Set containing value and all other values retained from m.
func (s Set) With(values ...interface{}) Set {
	t := s.n
	count := s.count
	for _, value := range values {
		var old interface{}
		t, old = t.put(value)
		if old == nil {
			count++
		}
	}
	return Set{
		n:     t,
		count: count,
	}
}

// Put returns a new Set with all values retained from Set except value.
func (s Set) Without(values ...interface{}) Set {
	t := s.n
	count := s.count
	for _, value := range values {
		var old interface{}
		t, old = t.delete(value)
		if old != nil {
			count--
		}
	}
	return Set{
		n:     t,
		count: count,
	}
}

// Has returns the value associated with key and true iff the key was found.
func (s Set) Has(val interface{}) bool {
	return s.n.get(val) != nil
}

func (s Set) Any() interface{} {
	for i := s.Range(); i.Next(); {
		return i.Value()
	}
	panic("empty set")
}

// Hash computes a hash value for s.
func (s Set) Hash() uint64 {
	var h uint64 = 10538386443025343807
	for i := s.Range(); i.Next(); {
		h ^= hash(i.Value())
	}
	return h
}

func (s Set) Equal(i interface{}) bool {
	if t, ok := i.(Set); ok {
		if s.Hash() != t.Hash() {
			return false
		}
		return s.SymmetricDifference(t).IsEmpty()
	}
	return false
}

func (s Set) String() string {
	return fmt.Sprintf("%v", s)
}

func (s Set) Format(f fmt.State, _ rune) {
	f.Write([]byte("["))
	for i, n := s.Range(), 0; i.Next(); n++ {
		if n > 0 {
			f.Write([]byte(", "))
		}
		fmt.Fprintf(f, "%v", i.Value())
	}
	f.Write([]byte("]"))
}

func (s Set) Where(pred func(i interface{}) bool) Set {
	return s.Reduce(func(r, i interface{}) interface{} {
		if pred(i) {
			return r.(Set).With(i)
		}
		return r
	}, NewSet()).(Set)
}

func (s Set) Map(f func(i interface{}) interface{}) Set {
	return s.Reduce(func(r, i interface{}) interface{} {
		return r.(Set).With(f(i))
	}, NewSet()).(Set)
}

func (s Set) Reduce(f func(acc, i interface{}) interface{}, acc interface{}) interface{} {
	for i := s.Range(); i.Next(); {
		acc = f(acc, i.Value())
	}
	return acc
}

func (s Set) Minus(t Set) Set {
	for i := t.Range(); i.Next(); {
		s = s.Without(i.Value())
	}
	return s
}

func (s Set) Intersection(t Set) Set {
	var r Set
	for i := s.Range(); i.Next(); {
		val := i.Value()
		if t.Has(val) {
			r = r.With(i.Value())
		}
	}
	return r
}

func (s Set) SymmetricDifference(t Set) Set {
	for i := t.Range(); i.Next(); {
		if s.Has(i.Value()) {
			s = s.Without(i.Value())
		} else {
			s = s.With(i.Value())
		}
	}
	return s
}

func (s Set) Union(t Set) Set {
	for i := t.Range(); i.Next(); {
		s = s.With(i.Value())
	}
	return s
}

func (s Set) Range() *SetIter {
	return &SetIter{i: s.n.iterator()}
}

type SetIter struct {
	i *hamtIter
}

func (i *SetIter) Next() bool {
	return i.i.next()
}

func (i *SetIter) Value() interface{} {
	return i.i.elem
}
