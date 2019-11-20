package frozen

import (
	"fmt"
	"strings"
)

// Set holds a set of values. The zero value is the empty Set.
type Set struct {
	t     hamt
	count int
	hash  uint64
}

var _ Key = Set{}

var emptySet = Set{t: empty{}}

func EmptySet() Set {
	return emptySet
}

func NewSet(values ...interface{}) Set {
	return EmptySet().With(values...)
}

func (s Set) IsEmpty() bool {
	return s.t.isEmpty()
}

func (s Set) Count() int {
	return s.count
}

// With returns a new Set containing value and all other values retained from m.
func (s Set) With(values ...interface{}) Set {
	t := s.t
	count := s.count
	h := s.hash
	for _, value := range values {
		var old *entry
		t, old = t.put(value, struct{}{})
		h ^= hash(value)
		if old != nil {
			h ^= hash(old.value)
		} else {
			count++
		}
	}
	return Set{t: t, count: count, hash: h}
}

// Put returns a new Set with all values retained from Set except value.
func (s Set) Without(values ...interface{}) Set {
	t := s.t
	count := s.count
	h := s.hash
	for _, value := range values {
		var old *entry
		t, old = t.delete(value)
		if old != nil {
			count--
			h ^= hash(old.value)
		}
	}
	return Set{t: t, count: count, hash: h}
}

// Has returns the value associated with key and true iff the key was found.
func (s Set) Has(value interface{}) bool {
	_, has := s.t.get(value)
	return has
}

// Hash computes a hash value for s.
func (s Set) Hash() uint64 {
	// go run github.com/marcelocantos/primal/cmd/random_primes 1
	return 10538386443025343807 ^ s.hash
}

func (s Set) Equal(i interface{}) bool {
	if t, ok := i.(Set); ok {
		return s.SymmetricDifference(t).IsEmpty()
	}
	return false
}

func (s Set) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i := s.Range(); i.Next(); {
		if i.Index() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", i.Value())
	}
	b.WriteString("]")
	return b.String()
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
		value := i.Value()
		if t.Has(value) {
			r = r.With(i.Value())
		}
	}
	return r
}

func (s Set) SymmetricDifference(t Set) Set {
	r := EmptySet()
	for i := s.Range(); i.Next(); {
		if !t.Has(i.Value()) {
			r = r.With(i.Value())
		}
	}
	for i := t.Range(); i.Next(); {
		if !s.Has(i.Value()) {
			r = r.With(i.Value())
		}
	}
	return r
}

func (s Set) Union(t Set) Set {
	r := EmptySet()
	for i := s.Range(); i.Next(); {
		r = r.With(i.Value())
	}
	for i := t.Range(); i.Next(); {
		r = r.With(i.Value())
	}
	return r
}

func (s Set) Range() *SetIter {
	return &SetIter{i: s.t.iterator()}
}

type SetIter struct {
	i *hamtIter
}

func (i *SetIter) Index() int {
	return i.i.i
}

func (i *SetIter) Next() bool {
	return i.i.next()
}

func (i *SetIter) Value() interface{} {
	return i.i.e.key
}
