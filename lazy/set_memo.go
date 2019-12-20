package lazy

import "github.com/marcelocantos/frozen"

type memoSet struct {
	Set
}

func memo(src Set) *memoSet {
	if m, ok := src.(*memoSet); ok {
		return m
	}
	return &memoSet{Set: src}
}

func (s *memoSet) Range() SetIterator {
	return &cachingIterator{iter: s.Set.Range(), memo: &s.Set}
}

type cachingIterator struct {
	iter SetIterator
	memo *Set
	seen frozen.Set
}

func (i *cachingIterator) Next() bool {
	for i.iter.Next() {
		if val := i.iter.Value(); !i.seen.Has(val) {
			i.seen = i.seen.With(val)
			return true
		}
	}
	*i.memo = Frozen(i.seen)
	return false
}

func (i *cachingIterator) Value() interface{} {
	return i.iter.Value()
}
