package iterator

type MapIterator[T comparable] struct {
	i Iterator[T]
	m func(v T) T
}

func Map[T comparable](i Iterator[T], m func(v T) T) *MapIterator[T] {
	return &MapIterator[T]{i: i, m: m}
}

func (m MapIterator[T]) Next() bool {
	return m.i.Next()
}

func (m MapIterator[T]) Value() T {
	return m.m(m.i.Value())
}
