package tree

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
	"github.com/arr-ai/frozen/v2/internal/pkg/value"
)

func elementEqual[T comparable](a, b T) bool {
	return value.Equal(a, b)
}

func hashValue[T comparable](i T, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}

func newSliceIterator[T comparable](slice []T) iterator.Iterator[T] {
	return iterator.NewSliceIterator(slice)
}
