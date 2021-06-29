package tree

import "fmt"

type node interface {
	fmt.Formatter
	fmt.Stringer

	Add(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int)
	AppendTo(dest []elementT) []elementT
	Canonical(depth int) node
	Combine(args *CombineArgs, n2 node, depth int) (_ node, matches int)
	Difference(args *EqArgs, n2 node, depth int) (_ node, matches int)
	Empty() bool
	Equal(args *EqArgs, n2 node, depth int) bool
	Get(args *EqArgs, v elementT, h hasher) *elementT
	Intersection(args *EqArgs, n2 node, depth int) (_ node, matches int)
	Iterator(buf [][]node) Iterator
	Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT
	SubsetOf(args *EqArgs, n2 node, depth int) bool
	Map(args *CombineArgs, depth int, f func(v elementT) elementT) (_ node, matches int)
	Vet()
	Where(args *WhereArgs, depth int) (_ node, matches int)
	With(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int)
	Without(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int)
	Remove(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int)
}
