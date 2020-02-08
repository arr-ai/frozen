package frozen

import (
	"fmt"
	"math/bits"

	"github.com/arr-ai/frozen/internal/tree"
	"github.com/arr-ai/frozen/types"
	"github.com/arr-ai/hash"
)

// Set holds a set of values. The zero value is the empty Set.
type Set struct {
	root  *tree.Node
	count int
}

var _ types.Key = Set{}

// NewSet creates a new Set with values as elements.
func NewSet(values ...interface{}) Set {
	var b SetBuilder
	for _, value := range values {
		b.Add(value)
	}
	return b.Finish()
}

// NewSetFromStrings creates a new Set with values as elements.
func NewSetFromStrings(values ...string) Set {
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
func (s Set) Range() types.Iterator {
	return s.root.Iterator(s.count)
}

func (s Set) Elements() []interface{} {
	result := make([]interface{}, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// OrderedElements takes elements in a defined order.
func (s Set) OrderedElements(less types.Less) []interface{} {
	result := make([]interface{}, 0, s.Count())
	for i := s.OrderedRange(less); i.Next(); {
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

// AnyN returns a set of N arbitrary elements from the Set.
func (s Set) AnyN(n int) Set {
	count := 0
	var setBuilder SetBuilder
	for i := s.Range(); i.Next() && count < n; count++ {
		setBuilder.Add(i.Value())
	}
	return setBuilder.Finish()
}

// OrderedFirstN returns a list of elements in a defined order.
func (s Set) OrderedFirstN(n int, less types.Less) []interface{} {
	result := make([]interface{}, 0, n)
	currentLength := 0
	for i := s.root.OrderedIterator(less, n); i.Next() && currentLength < n; currentLength++ {
		result = append(result, i.Value())
	}
	return result
}

// First returns the first element in a defined order.
func (s Set) First(less types.Less) interface{} {
	for _, i := range s.OrderedFirstN(1, less) {
		return i
	}
	panic("empty set")
}

// FirstN returns a set of the first n elements in a defined order.
func (s Set) FirstN(n int, less types.Less) Set {
	return NewSet(s.OrderedFirstN(n, less)...)
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

// OrderedRange returns a SetIterator for the Set that iterates over the elements in
// a specified order.
func (s Set) OrderedRange(less types.Less) Iterator {
	return s.root.OrderedIterator(less, s.Count())
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
	c := tree.NewCloner(false, s.Count())
	equalAsync := c.NoneFalse()
	equal := s.root.Equal(t.root, types.Equal, 0, c)
	return equal && equalAsync()
}

// IsSubsetOf returns true iff no element in s is not in t.
func (s Set) IsSubsetOf(t Set) bool {
	c := tree.NewCloner(false, s.Count())
	isSubsetAsync := c.NoneFalse()
	c.Update(s.root.IsSubsetOf(t.root, 0, c))
	return isSubsetAsync()
}

// Has returns the value associated with key and true iff the key was found.
func (s Set) Has(val interface{}) bool {
	return s.root.Get(val) != nil
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
func (s Set) Where(pred func(elem interface{}) bool) Set {
	c := tree.NewCloner(false, s.Count())
	matches := 0
	matchesAsync := c.Counter()
	var root *tree.Node
	s.root.Where(pred, 0, &matches, c, &root)
	matches += matchesAsync()
	return Set{root: root, count: matches}
}

// Map returns a Set with all the results of applying f to all elements in s.
func (s Set) Map(f func(elem interface{}) interface{}) Set {
	sbs := []*SetBuilder{}
	var spawn func() *tree.ForEacher
	spawn = func() *tree.ForEacher {
		var sb SetBuilder
		sbs = append(sbs, &sb)
		return tree.NewForEacher(func(elem interface{}) { sb.Add(f(elem)) }, spawn)
	}
	c := tree.NewCloner(false, s.Count())
	s.root.ForEach(spawn(), 0, c)
	c.Wait()

	sets := make([]Set, 0, len(sbs))
	for _, sb := range sbs {
		sets = append(sets, sb.Finish())
	}

	return Union(sets...)
}

// Reduce returns the result of applying `reduce` to the elements of `s` or
// `nil` if `s.IsEmpty()`. The result of each call is used as the acc argument
// for the next element.
//
// The `reduce` function must have the following properties:
//
//   - commutative: `reduce(a, b, c) == reduce(c, a, b)`
//   - associative: `reduce(reduce(a, b), c) == reduce(a, reduce(b, c))`
//
// By implication, `reduce` must accept its own output as input.
//
// 'elems` will never be empty.
func (s Set) Reduce(reduce func(elems ...interface{}) interface{}) interface{} {
	if s.Count() == 0 {
		return nil
	}

	pointers := make([]*[]interface{}, 0, bits.Len(uint(s.Count()))>>15)
	var spawn func() *tree.ForBatcher
	spawn = func() *tree.ForBatcher {
		var value []interface{}
		pointers = append(pointers, &value)
		return tree.NewForBatcher(
			func(elems ...interface{}) {
				value = append(value, reduce(elems...))
			},
			spawn,
		)
	}
	c := tree.NewCloner(false, s.Count())
	s.root.ForBatches(spawn(), 0, c)
	c.Wait()

	values := make([]interface{}, 0, len(pointers))
	// In case there are no elements above the parallelisation waterline.
	if *pointers[0] != nil {
		values = append(values, reduce(*pointers[0]...))
	}
	for _, p := range pointers[1:] {
		values = append(values, reduce(*p...))
	}
	return reduce(values...)
}

// Reduce2 is a convenience wrapper for `Reduce`, allowing the caller to
// implement a simpler, albeit less efficient, binary `reduce` function instead
// of an n-adic one.
func (s Set) Reduce2(reduce func(a, b interface{}) interface{}) interface{} {
	return s.Reduce(func(elems ...interface{}) interface{} {
		acc := elems[0]
		for _, elem := range elems[1:] {
			acc = reduce(acc, elem)
		}
		return acc
	})
}

// Intersection returns a Set with all elements that are in both s and t.
func (s Set) Intersection(t Set) Set {
	if s.Count() > t.Count() {
		s, t = t, s
	}
	c := tree.NewCloner(false, (s.Count()+t.Count())/2)
	countAsync := c.Counter()
	count := 0
	var root *tree.Node
	s.root.Intersection(t.root, 0, &count, c, &root)
	count += countAsync()
	return Set{root: root, count: count}
}

// Union returns a Set with all elements that are in either s or t.
func (s Set) Union(t Set) Set {
	c := tree.NewCloner(false, s.Count()+t.Count())
	matchesAsync := c.Counter()
	matches := 0
	root := s.root.Union(t.root, useRHS, 0, &matches, c)
	matches += matchesAsync()
	return Set{root: root, count: s.Count() + t.Count() - matches}
}

// Difference returns a Set with all elements that are s but not in t.
func (s Set) Difference(t Set) Set {
	c := tree.NewCloner(false, s.Count())
	matchesAsync := c.Counter()
	matches := 0
	var root *tree.Node
	s.root.Difference(t.root, 0, &matches, c, &root)
	matches += matchesAsync()
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
	for i := types.BitIterator(1); i < 1<<n; i++ {
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
