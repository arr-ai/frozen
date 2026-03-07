package tree

import (
	"sync"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

var defaultNPEqArgsCache sync.Map

// DefaultNPEqArgs provides default equality with non-parallel behaviour.
// The result is cached per type T since Gauge is a constant (NonParallel = -1)
// and EqArgs is read-only during Get/Without.
func DefaultNPEqArgs[T any]() *EqArgs[T] {
	key := TypeKeyOf[T]()
	if f, ok := defaultNPEqArgsCache.Load(key); ok {
		return f.(*EqArgs[T]) //nolint:forcetypeassert
	}
	args := NewDefaultEqArgs[T](depth.NonParallel)
	defaultNPEqArgsCache.Store(key, args)
	return args
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

// EqHash provides equality and hashing operations for tree elements.
type EqHash[T any] interface {
	Equal(a, b T) bool
	Hash(a T) H128
	FullHash() bool
}

type EqArgs[T any] struct {
	NodeArgs
	EqHash[T]
	hf func(T) H128 // cached from EqHash.Hash; avoids method-value closure in hot paths
}

// NewEqArgs creates EqArgs from an EqHash implementation.
func NewEqArgs[T any](gauge depth.Gauge, ops EqHash[T]) *EqArgs[T] {
	return &EqArgs[T]{
		NodeArgs: NewNodeArgs(gauge),
		EqHash:   ops,
		hf:       ops.Hash,
	}
}

// NewDefaultEqArgs creates EqArgs using the default resolved hash and equality functions.
func NewDefaultEqArgs[T any](gauge depth.Gauge) *EqArgs[T] {
	ops := getDefaultEqHash[T]()
	return &EqArgs[T]{
		NodeArgs: NewNodeArgs(gauge),
		EqHash:   ops,
		hf:       ops.hash,
	}
}

type WhereArgs[T any] struct {
	NodeArgs

	Pred func(elem T) bool
}

// defaultEqHash is the default EqHash implementation for Set (fullHash=true).
type defaultEqHash[T any] struct {
	eq   func(a, b T) bool
	hash func(T) H128
}

func (d *defaultEqHash[T]) Equal(a, b T) bool { return d.eq(a, b) }

func (d *defaultEqHash[T]) Hash(a T) H128 { return d.hash(a) }

func (d *defaultEqHash[T]) FullHash() bool { return true }

var defaultEqHashCache sync.Map

func getDefaultEqHash[T any]() *defaultEqHash[T] {
	key := TypeKeyOf[T]()
	if f, ok := defaultEqHashCache.Load(key); ok {
		return f.(*defaultEqHash[T]) //nolint:forcetypeassert
	}
	ops := &defaultEqHash[T]{
		eq:   value.EqualFuncFor[T](),
		hash: GetHashFunc[T](),
	}
	defaultEqHashCache.Store(key, ops)
	return ops
}
