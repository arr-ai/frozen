package tree

// import (
// 	"fmt"

// 	"github.com/arr-ai/frozen/errors"
// 	"github.com/arr-ai/frozen/internal/fu"
// )

// type empty struct{}

// // fmt.Formatter

// func (empty) Format(f fmt.State, verb rune) {
// 	fu.WriteString(f, "∅")
// }

// // fmt.Stringer

// func (empty) String() string {
// 	return "∅"
// }

// // node

// func (empty) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
// 	return newLeaf1(v)
// }

// func (empty) Canonical(depth int) node {
// 	return empty{}
// }

// func (empty) Combine(args *CombineArgs, n node, depth int, matches *int) node {
// 	return n
// }

// func (empty) AppendTo(dest []elementT) []elementT {
// 	return dest
// }

// func (empty) Difference(args *EqArgs, n node, depth int, removed *int) node {
// 	return empty{}
// }

// func (empty) Empty() bool {
// 	return true
// }

// func (empty) Equal(args *EqArgs, n node, depth int) bool {
// 	return n == empty{}
// }

// func (empty) Get(args *EqArgs, v elementT, _ hasher) *elementT {
// 	return nil
// }

// func (empty) Intersection(args *EqArgs, n node, depth int, matches *int) node {
// 	return empty{}
// }

// func (empty) Iterator([][]node) Iterator {
// 	return emptyIterator
// }

// func (empty) Reduce(_ NodeArgs, _ int, r func(values ...elementT) elementT) elementT {
// 	panic(errors.WTF)
// }

// func (empty) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
// 	return empty{}
// }

// func (empty) SubsetOf(args *EqArgs, n node, depth int) bool {
// 	return true
// }

// func (empty) Map(args *CombineArgs, _ int, counts *int, f func(e elementT) elementT) node {
// 	return empty{}
// }

// func (empty) Where(args *WhereArgs, depth int, matches *int) node {
// 	return empty{}
// }

// func (empty) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
// 	return newLeaf1(v)
// }

// func (empty) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
// 	return empty{}
// }
