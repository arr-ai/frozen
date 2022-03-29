package iterator

type MapIterator[T any] struct {
	i Iterator[T]
	m func(v T) T
}

func Map[T any](i Iterator[T], m func(v T) T) MapIterator[T] {
	return MapIterator[T]{i: i, m: m}
}

func (m MapIterator[T]) Next() bool {
	return m.i.Next()
}

func (m MapIterator[T]) Value() T {
	return m.m(m.i.Value())
}
