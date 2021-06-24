package tree

import (
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/value"
)

var (
	// DefaultNPEqArgs provides default equality with non-parallel behaviour.
	DefaultNPEqArgs = NewDefaultEqArgs(depth.NonParallel)

	// DefaultNPCombineArgs provides default combiner with non-parallel
	// behaviour.
	DefaultNPCombineArgs = NewCombineArgs(DefaultNPEqArgs, UseRHS)
)

type NodeArgs struct {
	depth.Gauge
}

func NewNodeArgs(gauge depth.Gauge) NodeArgs {
	return NodeArgs{
		Gauge: gauge,
	}
}

type CombineArgs struct {
	*EqArgs

	f func(a, b elementT) elementT

	flip *CombineArgs
}

func NewCombineArgs(ea *EqArgs, combine func(a, b elementT) elementT) *CombineArgs {
	flipped := func(a, b elementT) elementT { return combine(b, a) }
	ae := ea.flip
	args := &[2]CombineArgs{
		{EqArgs: ea, f: combine},
		{EqArgs: ae, f: flipped},
	}
	args[0].flip = &args[1]
	args[1].flip = &args[0]
	return &args[0]
}

type EqArgs struct {
	NodeArgs

	eq func(a, b elementT) bool
	// TODO
	lhash, rhash func(a elementT, seed uintptr) uintptr //nolint:structcheck

	flip *EqArgs
}

func NewEqArgs(
	gauge depth.Gauge,
	eq func(a, b elementT) bool,
	lhash, rhash func(a elementT, seed uintptr) uintptr,
) *EqArgs {
	qe := func(a, b elementT) bool { return eq(b, a) }
	na := NewNodeArgs(gauge)
	args := [2]EqArgs{
		{NodeArgs: na, eq: eq, lhash: lhash, rhash: rhash},
		{NodeArgs: na, eq: qe, lhash: rhash, rhash: lhash},
	}
	args[0].flip = &args[1]
	args[1].flip = &args[0]
	return &args[0]
}

func NewDefaultEqArgs(gauge depth.Gauge) *EqArgs {
	return NewEqArgs(gauge, value.Equal, hashValue, hashValue)
}

type WhereArgs struct {
	NodeArgs

	Pred func(elem elementT) bool
}
