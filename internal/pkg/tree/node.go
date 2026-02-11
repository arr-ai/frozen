package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/depth"
)

type node[T any] interface {
	fmt.Formatter
	fmt.Stringer

	Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int)
	AddFast(v T, depth int, h hasher) (_ node[T], matches int)
	AppendTo(dest []T) []T
	Combine(args *CombineArgs[T], n2 node[T], depth int) (_ node[T], matches int)
	Difference(gauge depth.Gauge, n2 node[T], depth int) (_ node[T], matches int)
	Empty() bool
	Equal(args *EqArgs[T], n2 node[T], depth int) bool
	Get(v T, h hasher, depth int) *T
	Intersection(gauge depth.Gauge, n2 node[T], depth int) (_ node[T], matches int)
	Iterator(buf [][]node[T]) Iterator[T]
	Map(args *CombineArgs[T], depth int, f func(v T) T) (_ node[T], matches int)
	Reduce(args NodeArgs, depth int, r func(values ...T) T) T
	SubsetOf(gauge depth.Gauge, n2 node[T], depth int) bool
	Vet() int
	Where(args *WhereArgs[T], depth int) (_ node[T], matches int)
	With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int)
	WithFast(v T, depth int, h hasher) (_ node[T], matches int)
	Without(v T, depth int, h hasher) (_ node[T], matches int)
	Remove(v T, depth int, h hasher) (_ node[T], matches int)
	clone() node[T]
}
