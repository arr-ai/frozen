package frozen

// IntLess dictates the order of two elements.
type IntLess func(a, b int) bool

type IntIterator interface {
	Next() bool
	Value() int
}
