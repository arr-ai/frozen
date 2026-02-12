package tree

import (
	internalIterator "github.com/arr-ai/frozen/internal/pkg/iterator"
)

func newSliceIterator[T any](slice []T) Iterator[T] {
	return internalIterator.NewSliceIterator(slice)
}
