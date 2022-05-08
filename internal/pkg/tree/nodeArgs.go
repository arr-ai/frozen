package tree

import (
	"github.com/arr-ai/frozen/internal/pkg/depth"
)

// DefaultNPEqArgs provides default equality with non-parallel behaviour.
func DefaultNPEqArgs[T any]() *EqArgs[T] {
	return NewDefaultEqArgs[T](depth.NonParallel)
}

// DefaultNPCombineArgs provides default combiner with non-parallel
// behaviour.
func DefaultNPCombineArgs[T any]() *CombineArgs[T] {
	return NewCombineArgs(DefaultNPEqArgs[T](), UseRHS[T])
}

type NodeArgs struct {
	depth.Gauge
}

func NewNodeArgs(gauge depth.Gauge) NodeArgs {
	return NodeArgs{
		Gauge: gauge,
	}
}

type CombineArgs[T any] struct {
	*EqArgs[T]

	f func(a, b T) T

	flipped *CombineArgs[T]
}

func NewCombineArgs[T any](ea *EqArgs[T], combine func(a, b T) T) *CombineArgs[T] {
	return &CombineArgs[T]{EqArgs: ea, f: combine}
}

func (a *CombineArgs[T]) Flip() *CombineArgs[T] {
	if a.flipped == nil {
		f := a.f
		a.flipped = &CombineArgs[T]{
			EqArgs:  a.EqArgs,
			f:       func(a, b T) T { return f(b, a) },
			flipped: a,
		}
	}
	return a.flipped
}

type EqArgs[T any] struct {
	NodeArgs

	eq func(a, b T) bool
	// TODO
	hash func(a T, seed uintptr) uintptr
}

func NewEqArgs[T any](
	gauge depth.Gauge,
	eq func(a, b T) bool,
	hash func(a T, seed uintptr) uintptr,
) *EqArgs[T] {
	na := NewNodeArgs(gauge)
	return &EqArgs[T]{
		NodeArgs: na,
		eq:       eq,
		hash:     hash,
	}
}

func NewDefaultEqArgs[T any](gauge depth.Gauge) *EqArgs[T] {
	return NewEqArgs(gauge, elementEqual[T], hashValue[T])
}

type WhereArgs[T any] struct {
	NodeArgs

	Pred func(elem T) bool
}
