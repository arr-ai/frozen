package iterator

type SliceIterator[T any] struct {
	slice []T
	index int
}

func NewSliceIterator[T any](slice []T) *SliceIterator[T] {
	return &SliceIterator[T]{slice: slice, index: -1}
}

func (i *SliceIterator[T]) Next() bool {
	i.index++
	return i.index < len(i.slice)
}

func (i *SliceIterator[T]) Value() T {
	return i.slice[i.index]
}
