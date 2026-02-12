package tree

import (
	internalIterator "github.com/arr-ai/frozen/internal/pkg/iterator"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

func elementEqual[T any](a, b T) bool {
	return value.Equal(a, b)
}

func hashValue[T any](key T, seed uintptr) uintptr {
	return getHashFunc[T]()(key, seed)
}

func newSliceIterator[T any](slice []T) Iterator[T] {
	return internalIterator.NewSliceIterator(slice)
}
