package iterator

import (
	"github.com/arr-ai/frozen/v2/pkg/errors"
)

// Empty is the empty iterator.
func Empty[T any]() Iterator[T] { return empty[T]{} }

type empty[T any] struct{}

func (empty[T]) Next() bool {
	return false
}

func (empty[T]) Value() T {
	panic(errors.WTF)
}
