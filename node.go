package frozen

import (
	"fmt"
)

type node interface {
	fmt.Stringer

	Canonical(depth int) node
	Combine(args *combineArgs, n node, depth int, matches *int) node
	CountUpTo(max int) int
	Defrost() unNode
	Difference(args *eqArgs, n node, depth int, removed *int) node
	Equal(args *eqArgs, n node, depth int) bool
	Get(args *eqArgs, v interface{}, h hasher) *interface{}
	Intersection(args *eqArgs, n node, depth int, matches *int) node
	Iterator(buf []packed) Iterator
	Reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{}
	SubsetOf(args *eqArgs, n node, depth int) bool
	Transform(args *combineArgs, depth int, count *int, f func(v interface{}) interface{}) node
	Where(args *whereArgs, depth int, matches *int) node
	With(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node
	Without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) node
}
