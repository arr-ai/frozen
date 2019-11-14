package frozen

// Map maps keys to values. The zero value is the empty Map.
type Map hamt

// Put returns a new Map with key associated with value and all other keys
// retained from m.
func (m Map) Put(key, value interface{}) Map {
	return Map(hamt(m).put(key, value))
}

// Put returns a new Map with all keys retained from m except key.
func (m Map) Delete(key interface{}) Map {
	return Map(hamt(m).delete(key))
}

// Get returns the value associated with key in m and true iff the key was
// found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	return hamt(m).get(key)
}

// Hash computes a hash value for s.
func (m Map) Hash() uint64 {
	// go run github.com/marcelocantos/primal/cmd/random_primes 2
	const random64BitPrime1 = 3167960924819262823
	const random64BitPrime2 = 4256779204343710393

	var h uint64 = random64BitPrime1
	for i := m.Range(); i.Next(); {
		h += random64BitPrime2*hash(i.Key()) + hash(i.Value())
	}
	return h
}

func (m Map) Range() *MapIter {
	return &MapIter{hamt(m).iterator()}
}

type MapIter struct {
	i *hamtIter
}

func (i *MapIter) Next() bool {
	return i.i.next()
}

func (i *MapIter) Key() interface{} {
	return i.i.e.key
}

func (i *MapIter) Value() interface{} {
	return i.i.e.value
}
