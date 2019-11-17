package frozen

// Set holds a set of values. The zero value is the empty Set.
type Set struct {
	t hamt
}

var emptySet = Set{t: empty{}}

func EmptySet() Set {
	return emptySet
}

// Add returns a new Set containing value and all other values retained from m.
func (s Set) Add(value interface{}) Set {
	return Set{t: s.t.put(value, struct{}{})}
}

// Put returns a new Set with all values retained from Set except value.
func (s Set) Delete(value interface{}) Set {
	return Set{t: s.t.delete(value)}
}

// Has returns the value associated with key and true iff the key was found.
func (s Set) Has(value interface{}) bool {
	_, has := s.t.get(value)
	return has
}

// Hash computes a hash value for s.
func (s Set) Hash() uint64 {
	// go run github.com/marcelocantos/primal/cmd/random_primes 1
	const random64BitPrime = 10538386443025343807

	var h uint64 = random64BitPrime
	for i := s.t.iterator(); i.next(); {
		h += hash(i.e.key)
	}
	return h
}

func (s Set) Range() *SetIter {
	return &SetIter{i: s.t.iterator()}
}

type SetIter struct {
	i *hamtIter
}

func (i *SetIter) Next() bool {
	return i.i.next()
}

func (i *SetIter) Value() interface{} {
	return i.i.e.key
}
