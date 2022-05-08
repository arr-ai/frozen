package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/iterator"
)

type node[T any] interface {
	fmt.Formatter
	fmt.Stringer

	Add(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int)
	AddFast(v T, depth int, h hasher) (_ node[T], matches int)
	AppendTo(dest []T) []T
	Canonical(depth int) node[T]
	Combine(args *CombineArgs[T], n2 node[T], depth int) (_ node[T], matches int)
	Difference(gauge depth.Gauge, n2 node[T], depth int) (_ node[T], matches int)
	Empty() bool
	Equal(args *EqArgs[T], n2 node[T], depth int) bool
	Get(v T, h hasher) *T
	Intersection(gauge depth.Gauge, n2 node[T], depth int) (_ node[T], matches int)
	Iterator(buf [][]node[T]) iterator.Iterator[T]
	Reduce(args NodeArgs, depth int, r func(values ...T) T) T
	SubsetOf(gauge depth.Gauge, n2 node[T], depth int) bool
	Map(args *CombineArgs[T], depth int, f func(v T) T) (_ node[T], matches int)
	Vet() int
	Where(args *WhereArgs[T], depth int) (_ node[T], matches int)
	With(args *CombineArgs[T], v T, depth int, h hasher) (_ node[T], matches int)
	WithFast(v T, depth int, h hasher) (_ node[T], matches int)
	Without(v T, depth int, h hasher) (_ node[T], matches int)
	Remove(v T, depth int, h hasher) (_ node[T], matches int)
	clone() node[T]
}
