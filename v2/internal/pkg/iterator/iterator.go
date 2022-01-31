package iterator

// Iterator provides for iterating over a Set.
type Iterator[T comparable] interface {
	Next() bool
	Value() T
}
