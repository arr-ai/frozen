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
		return hash(v.(KeyValue).Key, seed)
	}
}

type node interface {
	fmt.Stringer

	Canonical(depth int) node
	Combine(args *combineArgs, n node, depth int, matches *int) node
	CountUpTo(max int) int
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
