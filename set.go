package frozen

import (
	"fmt"

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

func (s Set) Elements() []interface{} {
	result := make([]interface{}, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
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
	h := hash.Uintptr(uintptr(10538386443025343807&uint64(^uintptr(0))), seed)
	for i := s.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), seed)
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
		return Equal(a.leaf().elems[0], b.leaf().elems[0])
	case a.isLeaf():
		return b.getImpl(a.leaf().elems[0], newHasher(a.leaf().elems[0], depth)) != nil
	case b.isLeaf():
		return false
	default:
		for mask := a.mask; mask != 0; mask = mask.Next() {
			i := mask.Index()
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
	return s.Difference(NewSet(values...))
}

// Where returns a Set with all elements that are in s and satisfy pred.
func (s Set) Where(pred func(el interface{}) bool) Set {
	var b SetBuilder
	for i := s.Range(); i.Next(); {
		v := i.Value()
		if pred(v) {
			b.Add(v)
		}
	}
	return b.Finish()
}

// Map returns a Set with all the results of applying f to all elements in s.
func (s Set) Map(f func(el interface{}) interface{}) Set {
	var b SetBuilder
	for i := s.Range(); i.Next(); {
		b.Add(f(i.Value()))
	}
	return b.Finish()
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
	count := 0
	root := s.root.intersection(t.root, 0, &count)
	return Set{root: root, count: count}
}

// Union returns a Set with all elements that are in either s or t.
func (s Set) Union(t Set) Set {
	matches := 0
	root := s.root.union(t.root, false, true, 0, &matches)
	return Set{root: root, count: s.Count() + t.Count() - matches}
}

// Difference returns a Set with all elements that are s but not in t.
func (s Set) Difference(t Set) Set {
	matches := 0
	root := s.root.difference(t.root, false, 0, &matches)
	return Set{root: root, count: s.Count() - matches}
}

// SymmetricDifference returns a Set with all elements that are s or t, but not
// both.
func (s Set) SymmetricDifference(t Set) Set {
	return s.Difference(t).Union(t.Difference(s))
}

func (s Set) Powerset() Set {
	n := s.Count()
	if n > 63 {
		panic("set too large")
	}
	elems := s.Elements()
	subset := Set{}
	result := NewSet(subset)
	for i := BitIterator(1); i < 1<<n; i++ {
		// Use a special counting order that flips a single bit at at time. The
		// bit to flip is the same as the lowest-order 1-bit in the normal
		// counting order, denoted by `(1)`. The flipped bit's new value is the
		// complement of the bit to the left of the `(1)`, denoted by `^-` and
		// `^1`.
		//
		//   ---------- plain ----------  |  ------- highlighted -------
		//      normal       1-bit flips  |     normal       1-bit flips
		//   ------------    -----------  |  ------------    -----------
		//    -  -  -  -     -  -  -  -   |   -  -  -  -     -  -  -  -
		//    -  -  -  1     -  -  -  1   |   -  - ^- (1)    -  -  - [1]
		//    -  -  1  -     -  -  1  1   |   - ^- (1) -     -  - [1] 1
		//    -  -  1  1     -  -  1  -   |   -  - ^1 (1)    -  -  1 [-]
		//    -  1  -  -     -  1  1  -   |  ^- (1) -  -     - [1] 1  -
		//    -  1  -  1     -  1  1  1   |   -  1 ^- (1)    -  1  1 [1]
		//    -  1  1  -     -  1  -  1   |   - ^1 (1) -     -  1 [-] 1
		//    -  1  1  1     -  1  -  -   |   -  1 ^1 (1)    -  1  - [-]
		//    1  -  -  -     1  1  -  -   |  (1) -  -  -    [1] 1  -  -
		//    1  -  -  1     1  1  -  1   |   1  - ^- (1)    1  1  - [1]
		//    1  -  1  -     1  1  1  1   |   1 ^- (1) -     1  1 [1] 1
		//    1  -  1  1     1  1  1  -   |   1  - ^1 (1)    1  1  1 [-]
		//    1  1  -  -     1  -  1  -   |  ^1 (1) -  -     1 [-] 1  -
		//    1  1  -  1     1  -  1  1   |   1  1 ^- (1)    1  -  1 [1]
		//    1  1  1  -     1  -  -  1   |   1 ^1 (1) -     1  - [-] 1
		//    1  1  1  1     1  -  -  -   |   1  1 ^1 (1)    1  -  - [-]
		//
		if flip := i.Index(); i.Has(flip + 1) {
			subset = subset.Without(elems[flip])
		} else {
			subset = subset.With(elems[flip])
		}
		result = result.With(subset)
	}
	return result
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

type setIter struct {
	i *nodeIter
}

func (i *setIter) Next() bool {
	return i.i.next()
}

func (i *setIter) Value() interface{} {
	return i.i.elem
}
