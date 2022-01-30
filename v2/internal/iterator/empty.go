package iterator

import (
	"github.com/arr-ai/frozen/v2/errors"
)

// Empty is the empty iterator.
func Empty[T comparable]() Iterator[T] { return empty[T]{} }

type empty[T comparable] struct{}

func (empty[T]) Next() bool {
	return false
}

func (empty[T]) Value() T {
	panic(errors.WTF)
}
