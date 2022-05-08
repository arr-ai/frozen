package lazy

// Predicate represents a function that returns true iff el satisfies some
// condition. The function must be pure. That is: a == b => p(a) == p(b).
type Predicate func(el any) bool

// Mapper represents a function that transforms el. The function must be
// pure. That is: a == b => f(a) == f(b).
type Mapper func(el any) any

// Set represents a set of elements.
type Set interface {
	// IsEmpty returns true iff there are no elements in this Set.
	IsEmpty() bool

	// FastIsEmpty returns IsEmpty() in O(1) time. Otherwise, ok=false.
	FastIsEmpty() (empty, ok bool)

	// Count returns the number of elements in this Set.
	Count() int

	// FastCount returns Count() in O(1) time. Otherwise, ok=false.
	FastCount() (count int, ok bool)

	// CountUpTo returns the cardinality of this set, up to limit. This avoids
	// the problem of counting intractable sets.
	CountUpTo(limit int) int

	// FastCountUpTo returns CountUpTo() in O(1) time. Otherwise, ok=false.
	FastCountUpTo(limit int) (count int, ok bool)

	// Freeze returns a frozen.Set with all the elements in this Set.
	Freeze() Set

	// Range returns an iterator over this Set. Traversal order is indeterminate
	// and may differ from one invocation of Range to the next.
	Range() SetIterator

	// Hash returns a hash derived from the elements of the set.
	Hash(seed uintptr) uintptr

	// Equal implements value.Equaler, returning true iff this Set and set
	// have all the same elements.
	Equal(set any) bool

	// EqualSet returns true iff this Set and set have all the same elements.
	EqualSet(set Set) bool

	// IsSubset returns true iff every element of this Set is in set.
	IsSubsetOf(set Set) bool

	// Has returns true iff el is in this Set.
	Has(el any) bool

	// FastHas returns Has(el) in <= O(log n) time. Otherwise, ok=false.
	FastHas(el any) (has, ok bool)

	// With returns a Set containing all the elements from this Set and all the
	// elements of els.
	With(v any) Set

	// With returns a Set containing all the elements from this Set except the
	// elements of els.
	Without(v any) Set

	// Where returns a Set containing all the elements from this Set that
	// satisfy pred.
	Where(pred Predicate) Set

	// Map returns a Set containing the results of calling m for all the
	// elements of this Set. Note that the result Set might have fewer elements
	// than this Set.
	Map(m Mapper) Set

	// Union returns a Set containing all values that are in either this Set or
	// set.
	Union(set Set) Set

	// Union returns a Set containing all values that are in both this Set and
	// set.
	Intersection(set Set) Set

	// Difference returns a Set containing all values that are in this Set but
	// not in set.
	Difference(set Set) Set

	// Difference returns a Set containing all values that are in this Set but
	// not in set.
	SymmetricDifference(set Set) Set

	// Powerset returns the set of all subsets of this Set.
	Powerset() Set
}

type SetIterator interface {
	Next() bool
	Value() any
}
