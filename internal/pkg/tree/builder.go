package tree

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

func DefaultNPKeyEqArgs[T any]() *EqArgs[T] {
	return NewDefaultKeyEqArgs[T](depth.NonParallel)
}

var defaultKeyCombineArgsCache sync.Map

func DefaultNPKeyCombineArgs[T any]() *CombineArgs[T] {
	var t T
	rt := reflect.TypeOf(&t)
	if f, ok := defaultKeyCombineArgsCache.Load(rt); ok {
		return f.(*CombineArgs[T]) //nolint:forcetypeassert
	}
	args := NewCombineArgs(DefaultNPKeyEqArgs[T](), UseRHS[T])
	defaultKeyCombineArgsCache.Store(rt, args)
	return args
}

func NewDefaultKeyEqArgs[T any](gauge depth.Gauge) *EqArgs[T] {
	return NewEqArgs(gauge, value.EqualFuncFor[T](), getHashFunc[T]())
}

// Builder[T] provides a more efficient way to build nodes incrementally.
type Builder[T any] struct {
	t    Tree[T]
	args *CombineArgs[T]
}

func NewBuilder[T any](int) *Builder[T] {
	return &Builder[T]{args: DefaultNPKeyCombineArgs[T]()}
}

func (b *Builder[T]) Count() int {
	return b.t.count
}

func (b *Builder[T]) add(args *CombineArgs[T], v T) {
	if b.t.root == nil {
		b.t.root = newLeaf1(v)
		b.t.count = 1
	} else {
		h := newHasherWith(v, 0, args.hash)
		if vetting {
			backup := b.clone()
			defer vet[T](func() { backup.add(args, v) }, &b.t)(nil)
		}
		var matches int
		b.t.root, matches = b.t.root.Add(args, v, 0, h)
		b.t.count += 1 - matches
	}
}

func (b *Builder[T]) Add(v T) {
	if b.args == nil {
		b.args = DefaultNPKeyCombineArgs[T]()
	}
	b.add(b.args, v)
}

func (b *Builder[T]) Remove(v T) {
	if b.t.root != nil {
		if b.args == nil {
			b.args = DefaultNPKeyCombineArgs[T]()
		}
		h := newHasherWith(v, 0, b.args.hash)
		if vetting {
			backup := b.clone()
			defer vet[T](func() { backup.Remove(v) }, &b.t)(nil)
		}
		var matches int
		b.t.root, matches = b.t.root.Remove(v, 0, h)
		b.t.count -= matches
	}
}

func (b *Builder[T]) Get(el T) *T {
	return b.t.Get(el)
}

func (b *Builder[T]) Finish() Tree[T] {
	t := b.Borrow()
	b.t = Tree[T]{}
	return t
}

func (b *Builder[T]) Borrow() Tree[T] {
	return b.t
}

func (b Builder[T]) String() string {
	return b.Borrow().String()
}

func (b Builder[T]) Format(state fmt.State, verb rune) {
	b.Borrow().Format(state, verb)
}

func (b *Builder[T]) clone() *Builder[T] {
	return &Builder[T]{t: b.t.clone()}
}
