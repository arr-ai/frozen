package lazy

import (
	"github.com/arr-ai/frozen"
	"github.com/marcelocantos/hash"
)

type emptySet struct{}

func (emptySet) IsEmpty() bool {
	return true
}

func (emptySet) FastIsEmpty() (empty, ok bool) {
	return true, true
}

func (emptySet) Count() int {
	return 0
}

func (emptySet) FastCount() (count int, ok bool) {
	return 0, true
}

func (emptySet) CountUpTo(limit int) int {
	return 0
}

func (emptySet) FastCountUpTo(limit int) (count int, ok bool) {
	return 0, true
}

func (emptySet) Freeze() Set {
	return Frozen(frozen.Set{})
}

func (emptySet) Equal(set interface{}) bool {
	if set, ok := set.(Set); ok {
		return set.EqualSet(set)
	}
	return false
}

func (emptySet) EqualSet(set Set) bool {
	return set.IsEmpty()
}

func (emptySet) Hash(seed uintptr) uintptr {
	return hash.Uintptr(hashSeed, seed)
}

func (emptySet) Has(el interface{}) bool {
	return false
}

func (emptySet) FastHas(el interface{}) (has, ok bool) {
	return false, true
}

func (emptySet) IsSubsetOf(set Set) bool {
	return true
}

func (emptySet) Range() SetIterator {
	return emptySetIterator{}
}

func (emptySet) Where(pred Predicate) Set {
	return emptySet{}
}

func (emptySet) With(els ...interface{}) Set {
	return Frozen(frozen.NewSet(els...))
}

func (emptySet) Without(els ...interface{}) Set {
	return emptySet{}
}

func (emptySet) Map(_ Mapper) Set {
	return emptySet{}
}

func (emptySet) Union(s Set) Set {
	return s
}

func (emptySet) Intersection(_ Set) Set {
	return emptySet{}
}

func (emptySet) Difference(_ Set) Set {
	return emptySet{}
}

func (emptySet) SymmetricDifference(s Set) Set {
	return s
}

func (emptySet) Powerset() Set {
	return Frozen(frozen.NewSet(emptySet{}))
}

type emptySetIterator struct{}

func (emptySetIterator) Next() bool {
	return false
}

func (emptySetIterator) Value() interface{} {
	panic("empty set")
}
