package tree

import (
	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/value"
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

	eq       func(a, b T) bool
	hash     func(a T) H128
	fullHash bool // H128 match implies equality; skip element comparison
}

// NewEqArgs creates EqArgs from an old-style seeded hash function.
// The seeded function is wrapped internally to produce H128 (calls with seeds 0 and 1).
func NewEqArgs[T any](
	gauge depth.Gauge,
	eq func(a, b T) bool,
	hash func(a T, seed uintptr) uintptr,
	fullHash bool,
) *EqArgs[T] {
	return &EqArgs[T]{
		NodeArgs: NewNodeArgs(gauge),
		eq:       eq,
		hash:     func(a T) H128 { return H128{hash(a, 0), hash(a, 1)} },
		fullHash: fullHash,
	}
}

// NewDefaultEqArgs creates EqArgs using the resolved H128 hash function directly,
// avoiding the seeded-function wrapper overhead.
func NewDefaultEqArgs[T any](gauge depth.Gauge) *EqArgs[T] {
	return &EqArgs[T]{
		NodeArgs: NewNodeArgs(gauge),
		eq:       value.EqualFuncFor[T](),
		hash:     getHashFunc[T](),
		fullHash: true,
	}
}

type WhereArgs[T any] struct {
	NodeArgs

	Pred func(elem T) bool
}
