package kvt

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/iterator/kvi"
	"github.com/arr-ai/frozen/pkg/kv"
)

type node interface {
	fmt.Stringer

	Canonical(depth int) node
	Combine(args *CombineArgs, n node, depth int, matches *int) node
	CountUpTo(max int) int
	Defrost() unNode
	Difference(args *EqArgs, n node, depth int, removed *int) node
	Equal(args *EqArgs, n node, depth int) bool
	Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue
	Intersection(args *EqArgs, n node, depth int, matches *int) node
	Iterator(buf []packer) kvi.Iterator
	Reduce(args NodeArgs, depth int, r func(values ...kv.KeyValue) kv.KeyValue) kv.KeyValue
	SubsetOf(args *EqArgs, n node, depth int) bool
	Transform(args *CombineArgs, depth int, count *int, f func(v kv.KeyValue) kv.KeyValue) node
	Where(args *WhereArgs, depth int, matches *int) node
	With(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) node
	Without(args *EqArgs, v kv.KeyValue, depth int, h hasher, matches *int) node
}
