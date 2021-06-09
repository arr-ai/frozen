package lazy

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen"
)

type EmptySet struct{}

func (EmptySet) IsEmpty() bool {
	return true
}

func (EmptySet) FastIsEmpty() (empty, ok bool) {
	return true, true
}

func (EmptySet) Count() int {
	return 0
}

func (EmptySet) FastCount() (count int, ok bool) {
	return 0, true
}

func (EmptySet) CountUpTo(limit int) int {
	return 0
}

func (EmptySet) FastCountUpTo(limit int) (count int, ok bool) {
	return 0, true
}

func (EmptySet) Freeze() Set {
	return Frozen(frozen.Set{})
}

func (EmptySet) Equal(set interface{}) bool {
	if set, ok := set.(Set); ok {
		return set.EqualSet(set)
	}
	return false
}

func (EmptySet) EqualSet(set Set) bool {
	return set.IsEmpty()
}

func (EmptySet) Hash(seed uintptr) uintptr {
	return hash.Uintptr(hashSeed, seed)
}

func (EmptySet) Has(el interface{}) bool {
	return false
}

func (EmptySet) FastHas(el interface{}) (has, ok bool) {
	return false, true
}

func (EmptySet) IsSubsetOf(set Set) bool {
	return true
}

func (EmptySet) Range() SetIterator {
	return emptySetIterator{}
}

func (EmptySet) Where(pred Predicate) Set {
	return EmptySet{}
}

func (EmptySet) With(els ...interface{}) Set {
	return Frozen(frozen.NewSet(els...))
}

func (EmptySet) Without(els ...interface{}) Set {
	return EmptySet{}
}

func (EmptySet) Map(_ Mapper) Set {
	return EmptySet{}
}

func (EmptySet) Union(s Set) Set {
	return s
}

func (EmptySet) Intersection(_ Set) Set {
	return EmptySet{}
}

func (EmptySet) Difference(_ Set) Set {
	return EmptySet{}
}

func (EmptySet) SymmetricDifference(s Set) Set {
	return s
}

func (EmptySet) Powerset() Set {
	return Frozen(frozen.NewSet(EmptySet{}))
}

type emptySetIterator struct{}

func (emptySetIterator) Next() bool {
	return false
}

func (emptySetIterator) Value() interface{} {
	panic("emptySetIterator.Value(): empty set")
}
