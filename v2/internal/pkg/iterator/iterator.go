package iterator

// Iterator provides for iterating over a Set.
type Iterator[T any] interface {
	Next() bool
	Value() T
}
