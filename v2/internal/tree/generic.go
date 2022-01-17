package tree

import (
	"reflect"

	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/v2/internal/iterator"
	"github.com/arr-ai/frozen/v2/internal/value"
)

// TODO: Can we avoid reflect.DeepEqual?
func isZero[T any](t T) bool {
	var zero T
	return reflect.DeepEqual(t, zero)
}

func elementEqual[T any](a, b T) bool {
	return value.Equal(a, b)
}

func hashValue[T any](i T, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}

func newSliceIterator[T any](slice []T) iterator.Iterator[T] {
	return iterator.NewSliceIterator(slice)
}
