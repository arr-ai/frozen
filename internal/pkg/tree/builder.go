package tree

import (
	"fmt"
	"sync"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

func DefaultNPKeyEqArgs[T any]() *EqArgs[T] {
	return NewDefaultKeyEqArgs[T](depth.NonParallel)
}

var defaultKeyCombineArgsCache sync.Map

func DefaultNPKeyCombineArgs[T any]() *CombineArgs[T] {
	key := typeKey[T]()
	if f, ok := defaultKeyCombineArgsCache.Load(key); ok {
		return f.(*CombineArgs[T]) //nolint:forcetypeassert
	}
	args := NewCombineArgs(DefaultNPKeyEqArgs[T](), UseRHS[T])
	defaultKeyCombineArgsCache.Store(key, args)
	return args
}

func NewDefaultKeyEqArgs[T any](gauge depth.Gauge) *EqArgs[T] {
	return &EqArgs[T]{
		NodeArgs: NewNodeArgs(gauge),
		eq:       value.EqualFuncFor[T](),
		hash:     getHashFunc[T](),
		fullHash: true,
	}
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
		b.t.root, matches = b.t.root.Remove(b.args.EqArgs, v, 0, h)
		b.t.count -= matches
	}
}

func (b *Builder[T]) Get(el T) *T {
	if b.t.root == nil {
		return nil
	}
	if b.args == nil {
		b.args = DefaultNPKeyCombineArgs[T]()
	}
	h := newHasherWith(el, 0, b.args.hash)
	return b.t.root.Get(b.args.EqArgs, el, h, 0)
}

func (b *Builder[T]) Finish() Tree[T] {
	t := b.Borrow()
	if b.args == nil {
		b.args = DefaultNPKeyCombineArgs[T]()
	}
	hf := b.args.hash
	if t.root != nil {
		computeH0(t.root, hf)
	}
	t.hf = hf
	b.t = Tree[T]{}
	return t
}

// computeH0 recursively fills in h0 for all nodes, bottom-up.
func computeH0[T any](n node[T], hf func(T) H128) {
	switch n := n.(type) {
	case *leaf1[T]:
		n.h0 = newElemH128(n.data, hf)
	case *leaf2[T]:
		n.ha = newElemH128(n.data[0], hf)
		n.h0 = n.ha.xor(newElemH128(n.data[1], hf))
	case *leaf[T]:
		n.h0 = H128{}
		for _, e := range n.data {
			n.h0 = n.h0.xor(newElemH128(e, hf))
		}
	case *branch[T]:
		n.h0 = H128{}
		for m := n.p.mask; m != 0; m = m.Next() {
			child := n.p.data[m.FirstIndex()]
			computeH0(child, hf)
			n.h0 = n.h0.xor(child.H0())
		}
	}
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
