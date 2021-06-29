package tree

import "fmt"

type node interface {
	fmt.Formatter
	fmt.Stringer

	Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node
	AppendTo(dest []elementT) []elementT
	Canonical(depth int) node
	Combine(args *CombineArgs, n2 node, depth int, matches *int) node
	Difference(args *EqArgs, n2 node, depth int, removed *int) node
	Empty() bool
	Equal(args *EqArgs, n2 node, depth int) bool
	Get(args *EqArgs, v elementT, h hasher) *elementT
	Intersection(args *EqArgs, n2 node, depth int, matches *int) node
	Iterator(buf [][]node) Iterator
	Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT
	SubsetOf(args *EqArgs, n2 node, depth int) bool
	Map(args *CombineArgs, depth int, count *int, f func(v elementT) elementT) node
	Vet()
	Where(args *WhereArgs, depth int, matches *int) node
	With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node
	Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node
	Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) node
}
