package frozen

import (
	"fmt"
	"math/bits"

	"github.com/marcelocantos/hash"
)

// Set holds a set of values. The zero value is the empty Set.
type Set struct {
	root  *node
	count int
}

var _ Key = Set{}

// Iterator provides for iterating over a Set.
type Iterator interface {
	Next() bool
	Value() interface{}
}

// NewSet creates a new Set with values as elements.
func NewSet(values ...interface{}) Set {
	var b SetBuilder
	for _, value := range values {
		b.Add(value)
	}
	return b.Finish()
}

// IsEmpty returns true iff the Set has no elements.
func (s Set) IsEmpty() bool {
	return s.root == nil
}

// Count returns the number of elements in the Set.
func (s Set) Count() int {
	return s.count
}

// Range returns an Iterator over the Set.
func (s Set) Range() Iterator {
	return &setIter{i: s.root.iterator()}
}

// Any returns an arbitrary element from the Set.
func (s Set) Any() interface{} {
	for i := s.Range(); i.Next(); {
		return i.Value()
	}
	panic("empty set")
}

// String returns a string representation of the Set.
func (s Set) String() string {
	return fmt.Sprintf("%v", s)
}

// Format writes a string representation of the Set into state.
func (s Set) Format(state fmt.State, _ rune) {
	state.Write([]byte("{"))
	for i, n := s.Range(), 0; i.Next(); n++ {
		if n > 0 {
			state.Write([]byte(", "))
		}
		fmt.Fprintf(state, "%v", i.Value())
	}
	state.Write([]byte("}"))
}

// RangeBy returns a SetIterator for the Set that iterates over the elements in
// a specified order.
func (s Set) RangeBy(less func(a, b interface{}) bool) Iterator {
	return s.root.orderedIterator(less, s.Count())
}

// Hash computes a hash value for s.
func (s Set) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(10538386443025343807, seed)
	for i := s.Range(); i.Next(); {
		h = hash.Interface(i.Value(), h)
	}
	return h
}

// Equal implements Equatable.
func (s Set) Equal(t interface{}) bool {
	if set, ok := t.(Set); ok {
		return s.EqualSet(set)
	}
	return false
}

// EqualSet returns true iff s and set have all the same elements.
func (s Set) EqualSet(t Set) bool {
	if s.root == nil || t.root == nil {
		return s.root == nil && t.root == nil
	}
	return s.root.equal(t.root, Equal)
}

func (s Set) IsSubsetOf(t Set) bool {
	return isSubsetOf(s.root, t.root, 0)
}

func isSubsetOf(a, b *node, depth int) bool {
	switch {
	case a == nil:
		return true
	case b == nil:
		return false
	case a.isLeaf() && b.isLeaf():
		return Equal(a.elem, b.elem)
	case a.isLeaf():
		return b.getImpl(a.elem, newHasher(a.elem, depth)) != nil
	case b.isLeaf():
		return false
	default:
		for mask := a.mask; mask != 0; mask &= mask - 1 {
			i := bits.TrailingZeros64(uint64(mask))
			if !isSubsetOf(a.children[i], b.children[i], depth+1) {
				return false
			}
		}
		return true
	}
}

// Has returns the value associated with key and true iff the key was found.
func (s Set) Has(val interface{}) bool {
	return s.root.get(val) != nil
}

// With returns a new Set retaining all the elements of the Set as well as values.
func (s Set) With(values ...interface{}) Set {
	return s.Union(NewSet(values...))
}

// Without returns a new Set with all values retained from Set except values.
func (s Set) Without(values ...interface{}) Set {
	return s.Minus(NewSet(values...))
}

// Where returns a Set with all elements that are in s and satisfy pred.
func (s Set) Where(pred func(el interface{}) bool) Set {
	return s.Reduce(func(r, i interface{}) interface{} {
		if pred(i) {
			return r.(Set).With(i)
		}
		return r
	}, NewSet()).(Set)
}

// Map returns a Set with all the results of applying f to all elements in s.
func (s Set) Map(f func(el interface{}) interface{}) Set {
	return s.Reduce(func(r, i interface{}) interface{} {
		return r.(Set).With(f(i))
	}, NewSet()).(Set)
}

// Reduce returns the result of applying f to each element of s. The result of
// each call is used as the acc argument for the next element.
func (s Set) Reduce(f func(acc, el interface{}) interface{}, acc interface{}) interface{} {
	for i := s.Range(); i.Next(); {
		acc = f(acc, i.Value())
	}
	return acc
}

// Intersection returns a Set with all elements that are in both s and t.
func (s Set) Intersection(t Set) Set {
	return s.merge(t, newIntersectionComposer())
}

// Union returns a Set with all elements that are in either s or t.
func (s Set) Union(t Set) Set {
	return s.merge(t, newUnionComposer(s.Count()+t.Count()))
}

// SymmetricDifference returns a Set with all elements that are s or t, but not
// both.
func (s Set) SymmetricDifference(t Set) Set {
	return s.merge(t, newSymmetricDifferenceComposer(s.Count()+t.Count()))
}

// Minus returns a Set with all elements that are s but not in t.
func (s Set) Minus(t Set) Set {
	return s.merge(t, newMinusComposer(s.Count()))
}

// GroupBy returns a Map that groups elements in the Set by their key.
func (s Set) GroupBy(key func(el interface{}) interface{}) Map {
	var builders MapBuilder
	for i := s.Range(); i.Next(); {
		v := i.Value()
		k := key(v)
		var b *SetBuilder
		if builder, has := builders.Get(k); has {
			b = builder.(*SetBuilder)
		} else {
			b = &SetBuilder{}
			builders.Put(k, b)
		}
		b.Add(v)
	}
	var result MapBuilder
	for i := builders.Finish().Range(); i.Next(); {
		result.Put(i.Key(), i.Value().(*SetBuilder).Finish())
	}
	return result.Finish()
}

func (s Set) merge(t Set, c *composer) Set {
	n := s.root.merge(t.root, c)
	return Set{root: n, count: c.count()}
}

type setIter struct {
	i *nodeIter
}

func (i *setIter) Next() bool {
	return i.i.next()
}

func (i *setIter) Value() interface{} {
	return i.i.elem
}