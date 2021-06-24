package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/iterator"
)

type node interface {
	fmt.Stringer

	Canonical(depth int) node
	Combine(args *CombineArgs, n node, depth int, matches *int) node
	CopyTo(dest []interface{}) []interface{}
	Defrost() unNode
	Difference(args *EqArgs, n node, depth int, removed *int) node
	Empty() bool
	Equal(args *EqArgs, n node, depth int) bool
	Get(args *EqArgs, v interface{}, h hasher) *interface{}
	Intersection(args *EqArgs, n node, depth int, matches *int) node
	Iterator(buf [][]node) iterator.Iterator
	Reduce(args NodeArgs, depth int, r func(values ...interface{}) interface{}) interface{}
	SubsetOf(args *EqArgs, n node, depth int) bool
	Transform(args *CombineArgs, depth int, count *int, f func(v interface{}) interface{}) node
	Where(args *WhereArgs, depth int, matches *int) node
	With(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) node
	Without(args *EqArgs, v interface{}, depth int, h hasher, matches *int) node
}
