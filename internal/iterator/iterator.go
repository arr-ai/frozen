package iterator

// Iterator provides for iterating over a Set.
type Iterator interface {
	Next() bool
	Value() elementT
}
