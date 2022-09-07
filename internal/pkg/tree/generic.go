package tree

import (
	"github.com/arr-ai/hash"

	internalIterator "github.com/arr-ai/frozen/internal/pkg/iterator"
	"github.com/arr-ai/frozen/internal/pkg/value"
	"github.com/arr-ai/frozen/pkg/iterator"
)

func elementEqual[T any](a, b T) bool {
	return value.Equal(a, b)
}

func hashValue[T any](i T, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}

func newSliceIterator[T any](slice []T) iterator.Iterator[T] {
	return internalIterator.NewSliceIterator(slice)
}
