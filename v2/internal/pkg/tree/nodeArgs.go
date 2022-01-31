package tree

import (
	"github.com/arr-ai/frozen/v2/internal/pkg/depth"
)

// DefaultNPEqArgs provides default equality with non-parallel behaviour.
func DefaultNPEqArgs[T comparable]() *EqArgs[T] {
	return NewDefaultEqArgs[T](depth.NonParallel)
}

// DefaultNPCombineArgs provides default combiner with non-parallel
// behaviour.
func DefaultNPCombineArgs[T comparable]() *CombineArgs[T] {
	return NewCombineArgs[T](DefaultNPEqArgs[T](), UseRHS[T])
}

type NodeArgs struct {
	depth.Gauge
}

func NewNodeArgs(gauge depth.Gauge) NodeArgs {
	return NodeArgs{
		Gauge: gauge,
	}
}

type CombineArgs[T comparable] struct {
	*EqArgs[T]

	f func(a, b T) T

	flipped *CombineArgs[T]
}

func NewCombineArgs[T comparable](ea *EqArgs[T], combine func(a, b T) T) *CombineArgs[T] {
	return &CombineArgs[T]{EqArgs: ea, f: combine}
}

func (a *CombineArgs[T]) Flip() *CombineArgs[T] {
	if a.flipped == nil {
		f := a.f
		a.flipped = &CombineArgs[T]{
			EqArgs:  a.EqArgs.Flip(),
			f:       func(a, b T) T { return f(b, a) },
			flipped: a,
		}
	}
	return a.flipped
}

type EqArgs[T comparable] struct {
	NodeArgs

	eq func(a, b T) bool
	// TODO
	lhash, rhash func(a T, seed uintptr) uintptr

	flipped *EqArgs[T]
}

func NewEqArgs[T comparable](
	gauge depth.Gauge,
	eq func(a, b T) bool,
	lhash, rhash func(a T, seed uintptr) uintptr,
) *EqArgs[T] {
	na := NewNodeArgs(gauge)
	return &EqArgs[T]{
		NodeArgs: na,
		eq:       eq,
		lhash:    lhash,
		rhash:    rhash,
	}
}

func NewDefaultEqArgs[T comparable](gauge depth.Gauge) *EqArgs[T] {
	return NewEqArgs[T](gauge, elementEqual[T], hashValue[T], hashValue[T])
}

func (a *EqArgs[T]) Flip() *EqArgs[T] {
	if a.flipped == nil {
		eq := a.eq
		a.flipped = &EqArgs[T]{
			NodeArgs: a.NodeArgs,
			eq:       func(a, b T) bool { return eq(b, a) },
			lhash:    a.rhash,
			rhash:    a.lhash,
			flipped:  a,
		}
	}
	return a.flipped
}

type WhereArgs[T comparable] struct {
	NodeArgs

	Pred func(elem T) bool
}
