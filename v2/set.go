package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/pkg/depth"
	"github.com/arr-ai/frozen/v2/internal/pkg/fu"
	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
	"github.com/arr-ai/frozen/v2/internal/pkg/tree"
	"github.com/arr-ai/frozen/v2/internal/pkg/value"
)

// Set holds a set of values of type T. The zero value is the empty Set.
type Set[T any] struct {
	tree tree.Tree[T]
}

var _ value.Key[Set[int]] = Set[int]{}

func newSet[T any](tree tree.Tree[T]) Set[T] {
	return Set[T]{tree: tree}
}

// NewSet creates a new Set with values as elements.
func NewSet[T any](values ...T) Set[T] {
	b := NewSetBuilder[T](len(values))
	for _, value := range values {
		b.Add(value)
	}
	return b.Finish()
}

func (s Set[T]) nodeArgs() tree.NodeArgs {
	return tree.NewNodeArgs(depth.NewGauge(s.Count()))
}

func (s Set[T]) eqArgs() *tree.EqArgs[T] {
	return tree.NewDefaultEqArgs[T](depth.NewGauge(s.Count()))
}

// IsEmpty returns true iff the Set has no elements.
func (s Set[T]) IsEmpty() bool {
	return s.tree.Count() == 0
}

// Count returns the number of elements in the Set.
func (s Set[T]) Count() int {
	return s.tree.Count()
}

// Range returns an Iterator over the Set.
func (s Set[T]) Range() iterator.Iterator[T] {
	return s.tree.Iterator()
}

func (s Set[T]) Elements() []T {
	result := make([]T, 0, s.Count())
	for i := s.Range(); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// OrderedElements takes elements in a defined order.
func (s Set[T]) OrderedElements(less tree.Less[T]) []T {
	result := make([]T, 0, s.Count())
	for i := s.OrderedRange(less); i.Next(); {
		result = append(result, i.Value())
	}
	return result
}

// Any returns an arbitrary element from the Set.
func (s Set[T]) Any() T {
	for i := s.Range(); i.Next(); {
		return i.Value()
	}
	panic("Set[T].Any(): empty set")
}

// AnyN returns a set of N arbitrary elements from the Set.
func (s Set[T]) AnyN(n int) Set[T] {
	if n > s.Count() {
		return s
	}
	b := NewSetBuilder[T](n)
	for i := s.Range(); i.Next() && n > 0; n-- {
		b.Add(i.Value())
	}
	return b.Finish()
}

// OrderedFirstN returns a list of elements in a defined order.
func (s Set[T]) OrderedFirstN(n int, less tree.Less[T]) []T {
	result := make([]T, 0, n)
	currentLength := 0
	for i := s.tree.OrderedIterator(less, n); i.Next() && currentLength < n; currentLength++ {
		result = append(result, i.Value())
	}
	return result
}

// First returns the first element in a defined order.
func (s Set[T]) First(less tree.Less[T]) interface{} {
	for _, i := range s.OrderedFirstN(1, less) {
		return i
	}
	panic("Set[T].First(): empty set")
}

// FirstN returns a set of the first n elements in a defined order.
func (s Set[T]) FirstN(n int, less tree.Less[T]) Set[T] {
	return NewSet[T](s.OrderedFirstN(n, less)...)
}

// String returns a string representation of the Set.
func (s Set[T]) String() string {
	return fu.String(s)
}

// Format writes a string representation of the Set into state.
func (s Set[T]) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fmt.Fprint(f, s.tree.String())
		return
	}

	w, set := f.Width()

	fu.WriteString(f, "{")
	for i, n := s.Range(), 0; i.Next(); n++ {
		if set && n == w {
			fu.WriteString(f, fmt.Sprintf(", ...(+%d)...", s.Count()-n))
			break
		}
		if n > 0 {
			fu.WriteString(f, ", ")
		}
		fmt.Fprintf(f, "%v", i.Value())
	}
	fu.WriteString(f, "}")
}

// OrderedRange returns a SetIterator for the Set that iterates over the elements in
// a specified order.
func (s Set[T]) OrderedRange(less tree.Less[T]) iterator.Iterator[T] {
	return s.tree.OrderedIterator(less, s.Count())
}

// Hash computes a hash value for s.
func (s Set[T]) Hash(seed uintptr) uintptr {
	h := hash.Uintptr(uintptr(10538386443025343807&uint64(^uintptr(0))), seed)
	for i := s.Range(); i.Next(); {
		h ^= hash.Interface(i.Value(), seed)
	}
	return h
}

// Equal returns true iff s and set have all the same elements.
func (s Set[T]) Equal(t Set[T]) bool {
	args := s.eqArgs()
	return s.tree.Equal(args, t.tree)
}

// IsSubsetOf returns true iff no element in s is not in t.
func (s Set[T]) IsSubsetOf(t Set[T]) bool {
	args := s.eqArgs()
	return s.tree.SubsetOf(args, t.tree)
}

// Has returns the value associated with key and true iff the key was found.
func (s Set[T]) Has(val T) bool {
	return s.tree.Get(tree.DefaultNPEqArgs[T](), val) != nil
}

// With returns a new Set retaining all the elements of the Set as well as values.
func (s Set[T]) With(values ...T) Set[T] {
	return s.Union(NewSet(values...))
}

// Without returns a new Set with all values retained from Set except values.
func (s Set[T]) Without(values ...T) Set[T] {
	return s.Difference(NewSet(values...))
}

// Where returns a Set with all elements that are in s and satisfy pred.
func (s Set[T]) Where(pred func(elem T) bool) Set[T] {
	args := &tree.WhereArgs[T]{
		NodeArgs: tree.NewNodeArgs(depth.NewGauge(s.Count())),
		Pred:     pred,
	}
	// root = root.postop(c.parallelDepth)
	return Set[T]{tree: s.tree.Where(args)}
}

// // Map returns a Set with all the results of applying f to all elements in s.
// func SetMap[T, U any](s Set, f func(elem T) U) Set[U] {
// 	args := tree.NewCombineArgs(s.eqArgs(), tree.UseRHS)
// 	return Set[U]{tree: s.tree.Map(args, f)}
// }

// // Reduce returns the result of applying `reduce` to the elements of `s` or
// // `nil` if `s.IsEmpty()`. The result of each call is used as the acc argument
// // for the next element.
// //
// // The `reduce` function must have the following properties:
// //
// //   - commutative: `reduce(a, b, c) == reduce(c, a, b)`
// //   - associative: `reduce(reduce(a, b), c) == reduce(a, reduce(b, c))`
// //
// // By implication, `reduce` must accept its own output as input.
// //
// // 'elems` will never be empty.
// func (s Set) Reduce(reduce func(elems ...interface{}) interface{}) interface{} {
// 	return s.tree.Reduce(s.nodeArgs(), reduce)
// }

// // Reduce2 is a convenience wrapper for `Reduce`, allowing the caller to
// // implement a simpler, albeit less efficient, binary `reduce` function instead
// // of an n-adic one.
// func (s Set) Reduce2(reduce func(a, b interface{}) interface{}) interface{} {
// 	return s.Reduce(func(elems ...interface{}) interface{} {
// 		acc := elems[0]
// 		for _, elem := range elems[1:] {
// 			acc = reduce(acc, elem)
// 		}
// 		return acc
// 	})
// }

// Intersection returns a Set with all elements that are in both s and t.
func (s Set[T]) Intersection(t Set[T]) Set[T] {
	return Set[T]{tree: s.tree.Intersection(s.eqArgs(), t.tree)}
}

// Union returns a Set with all elements that are in either s or t.
func (s Set[T]) Union(t Set[T]) Set[T] {
	args := tree.NewCombineArgs(s.eqArgs(), tree.UseRHS[T])
	return Set[T]{tree: s.tree.Combine(args, t.tree)}
}

// Difference returns a Set with all elements that are s but not in t.
func (s Set[T]) Difference(t Set[T]) Set[T] {
	args := s.eqArgs()
	return Set[T]{tree: s.tree.Difference(args, t.tree)}
}

// SymmetricDifference returns a Set with all elements that are s or t, but not
// both.
func (s Set[T]) SymmetricDifference(t Set[T]) Set[T] {
	st := s.Difference(t)
	ts := t.Difference(s)
	return st.Union(ts)
}

func Powerset[T any](s Set[T]) Set[Set[T]] {
	n := s.Count()
	if n > 63 {
		panic("set too large")
	}
	elems := s.Elements()
	subset := Set[T]{}
	result := NewSet(subset)
	for i := iterator.BitIterator(1); i < 1<<uint(n); i++ {
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

// // GroupBy returns a Map that groups elements in the Set by their key.
// func (s Set) GroupBy(key func(el interface{}) interface{}) Map {
// 	var builders MapBuilder
// 	for i := s.Range(); i.Next(); {
// 		v := i.Value()
// 		k := key(v)
// 		var b *SetBuilder
// 		if builder, has := builders.Get(k); has {
// 			b = builder.(*SetBuilder)
// 		} else {
// 			b = &SetBuilder{}
// 			builders.Put(k, b)
// 		}
// 		b.Add(v)
// 	}
// 	var result MapBuilder
// 	fb := builders.Finish()
// 	for i := fb.Range(); i.Next(); {
// 		result.Put(i.Key(), i.Value().(*SetBuilder).Finish())
// 	}
// 	return result.Finish()
// }
