package frozen

import (
	"fmt"

	"github.com/arr-ai/hash"
)

var (
	defaultNPEqArgs      = newDefaultEqArgs(nonParallel)
	defaultNPCombineArgs = newCombineArgs(defaultNPEqArgs, useRHS)

	defaultNPKeyEqArgs      = newDefaultKeyEqArgs(nonParallel)
	defaultNPKeyCombineArgs = newCombineArgs(defaultNPKeyEqArgs, useRHS)

	keyHash = keyHasher(hash.Interface)
)

func keyHasher(hash func(v interface{}, seed uintptr) uintptr) func(v interface{}, seed uintptr) uintptr {
	return func(v interface{}, seed uintptr) uintptr {
		return hash(v.(KeyValue).Value, seed)
	}
}

type node interface {
	fmt.Stringer

	canonical(depth int) node
	combine(args *combineArgs, n node, depth int, matches *int) node
	countUpTo(max int) int
	difference(args *eqArgs, n node, depth int, removed *int) node
	equal(args *eqArgs, n node, depth int) bool
	get(args *eqArgs, v interface{}, h hasher) *interface{}
	intersection(args *eqArgs, n node, depth int, matches *int) node
	isSubsetOf(args *eqArgs, n node, depth int) bool
	iterator(buf []packed) Iterator
	reduce(args nodeArgs, depth int, r func(values ...interface{}) interface{}) interface{}
	transform(args *combineArgs, depth int, count *int, f func(v interface{}) interface{}) node
	vet() node
	where(args *whereArgs, depth int, matches *int) node
	with(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node
	without(args *eqArgs, v interface{}, depth int, h hasher, matches *int) node
}

func vet(n *node) {
	// (*n).vet()
}

type nodeArgs struct {
	parallelDepthGauge
}

func newNodeArgs(gauge parallelDepthGauge) nodeArgs {
	return nodeArgs{
		parallelDepthGauge: gauge,
	}
}

type combineArgs struct {
	*eqArgs

	f func(a, b interface{}) interface{}

	flip *combineArgs
}

func newCombineArgs(ea *eqArgs, combine func(a, b interface{}) interface{}) *combineArgs {
	flipped := func(a, b interface{}) interface{} { return combine(b, a) }
	ae := ea.flip
	args := &[2]combineArgs{
		{eqArgs: ea, f: combine},
		{eqArgs: ae, f: flipped},
	}
	args[0].flip = &args[1]
	args[1].flip = &args[0]
	return &args[0]
}

type eqArgs struct {
	nodeArgs

	eq func(a, b interface{}) bool
	// TODO
	lhash, rhash func(a interface{}, seed uintptr) uintptr

	flip *eqArgs
}

func newEqArgs(
	gauge parallelDepthGauge,
	eq func(a, b interface{}) bool,
	lhash, rhash func(a interface{}, seed uintptr) uintptr,
) *eqArgs {
	qe := func(a, b interface{}) bool { return eq(b, a) }
	na := newNodeArgs(gauge)
	args := [2]eqArgs{
		{nodeArgs: na, eq: eq, lhash: lhash, rhash: rhash},
		{nodeArgs: na, eq: qe, lhash: rhash, rhash: lhash},
	}
	args[0].flip = &args[1]
	args[1].flip = &args[0]
	return &args[0]
}

func newDefaultEqArgs(gauge parallelDepthGauge) *eqArgs {
	return newEqArgs(gauge, Equal, hash.Interface, hash.Interface)
}

func newDefaultKeyEqArgs(gauge parallelDepthGauge) *eqArgs {
	return newEqArgs(gauge, KeyEqual, keyHash, keyHash)
}

type whereArgs struct {
	nodeArgs

	pred func(elem interface{}) bool
}
