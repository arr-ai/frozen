package tree

import (
	"github.com/arr-ai/hash"

	"github.com/arr-ai/frozen/internal/iterator"
	"github.com/arr-ai/frozen/internal/value"
)

type (
	elementT = interface{}
	Iterator = iterator.Iterator
)

var emptyIterator = iterator.Empty

func elementEqual(a, b interface{}) bool {
	return value.Equal(a, b)
}

func hashValue(i elementT, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}

func isBlank(i interface{}) bool {
	return i == nil
}

func interfaceAsElement(i interface{}) elementT {
	return i
}

func newSliceIterator(slice []interface{}) Iterator {
	return iterator.NewSliceIterator(slice)
}
