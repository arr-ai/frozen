package tree

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/pkg/iterator"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

func elementEqual[T any](a, b T) bool {
	return value.Equal(a, b)
}

func hashValue[T any](i T, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}

func newSliceIterator[T any](slice []T) iterator.Iterator[T] {
	return iterator.NewSliceIterator(slice)
}
