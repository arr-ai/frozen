package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/fu"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

// collision is a fallback node for elements with identical hash codes across
// all rounds (i.e. pathological custom hash functions). It stores elements in
// a flat slice and uses linear scanning. Under normal hash functions (hash.Any),
// this node type is never created.
type collision[T any] struct {
	data []T
}

func newCollision[T any](data ...T) *collision[T] {
	return &collision[T]{data: data}
}

// fmt.Formatter

func (c *collision[T]) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	for i, e := range c.data {
		if i > 0 {
			fu.WriteString(f, ",")
		}
		fu.Format(e, f, verb)
	}
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (c *collision[T]) String() string {
	return fmt.Sprintf("%s", c)
}

// node[T]

func (c *collision[T]) Add(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if args.eq(e, v) {
			c.data[i] = args.f(e, v)
			return c, 1
		}
	}
	c.data = append(c.data, v)
	return c, 0
}

func (c *collision[T]) AddFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if value.Equal(e, v) {
			c.data[i] = v
			return c, 1
		}
	}
	c.data = append(c.data, v)
	return c, 0
}

func (c *collision[T]) AppendTo(dest []T) []T {
	if len(dest)+len(c.data) > cap(dest) {
		return nil
	}
	return append(dest, c.data...)
}

func (c *collision[T]) Combine(args *CombineArgs[T], n node[T], _ int) (_ node[T], matches int) {
	switch n := n.(type) {
	case *collision[T]:
		ret := &collision[T]{data: append([]T(nil), c.data...)}
		for _, e := range n.data {
			found := false
			for j, f := range ret.data {
				if args.eq(f, e) {
					ret.data[j] = args.f(f, e)
					matches++
					found = true
					break
				}
			}
			if !found {
				ret.data = append(ret.data, e)
			}
		}
		return ret, matches
	case *leaf1[T]:
		for i, e := range c.data {
			if args.eq(e, n.data) {
				ret := &collision[T]{data: append([]T(nil), c.data...)}
				ret.data[i] = args.f(e, n.data)
				return ret, 1
			}
		}
		ret := &collision[T]{data: append(append([]T(nil), c.data...), n.data)}
		return ret, 0
	case *branch[T]:
		return n.Combine(args.Flip(), c, 0)
	default:
		panic("unexpected node type in collision.Combine")
	}
}

func (c *collision[T]) Difference(_ depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	var ret []T
	for _, e := range c.data {
		h := newHasher(e, depth)
		if n.Get(e, h, depth) != nil {
			matches++
		} else {
			ret = append(ret, e)
		}
	}
	return collisionCanonical[T](ret), matches
}

func (c *collision[T]) Empty() bool {
	return len(c.data) == 0
}

func (c *collision[T]) Equal(args *EqArgs[T], n node[T], _ int) bool {
	n2, ok := n.(*collision[T])
	if !ok || len(c.data) != len(n2.data) {
		return false
	}
outer:
	for _, e := range c.data {
		for _, f := range n2.data {
			if args.eq(e, f) {
				continue outer
			}
		}
		return false
	}
	return true
}

func (c *collision[T]) Get(v T, _ hasher, _ int) *T {
	for i, e := range c.data {
		if value.Equal(e, v) {
			return &c.data[i]
		}
	}
	return nil
}

func (c *collision[T]) Intersection(_ depth.Gauge, n node[T], depth int) (_ node[T], matches int) {
	var ret []T
	for _, e := range c.data {
		h := newHasher(e, depth)
		if n.Get(e, h, depth) != nil {
			ret = append(ret, e)
			matches++
		}
	}
	return collisionCanonical[T](ret), matches
}

func (c *collision[T]) Iterator([][]node[T]) Iterator[T] {
	return newSliceIterator(c.data)
}

func (c *collision[T]) Map(args *CombineArgs[T], _ int, f func(e T) T) (_ node[T], matches int) {
	var b Builder[T]
	for _, e := range c.data {
		b.add(args, f(e))
	}
	t := b.Finish()
	return t.root, t.count
}

func (c *collision[T]) Reduce(_ NodeArgs, _ int, r func(values ...T) T) T {
	return r(c.data...)
}

func (c *collision[T]) Remove(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if value.Equal(e, v) {
			last := len(c.data) - 1
			if last == 0 {
				return nil, 1
			}
			if i < last {
				c.data[i] = c.data[last]
			}
			c.data = c.data[:last]
			return collisionCanonical[T](c.data), 1
		}
	}
	return c, 0
}

func (c *collision[T]) SubsetOf(_ depth.Gauge, n node[T], depth int) bool {
	for _, e := range c.data {
		h := newHasher(e, depth)
		if n.Get(e, h, depth) == nil {
			return false
		}
	}
	return true
}

func (c *collision[T]) Vet() int {
	if len(c.data) < 2 {
		panic(fmt.Errorf("collision too small (%d)", len(c.data)))
	}
	return len(c.data)
}

func (c *collision[T]) Where(args *WhereArgs[T], _ int) (_ node[T], matches int) {
	var ret []T
	for _, e := range c.data {
		if args.Pred(e) {
			ret = append(ret, e)
			matches++
		}
	}
	return collisionCanonical[T](ret), matches
}

func (c *collision[T]) With(args *CombineArgs[T], v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if args.eq(e, v) {
			ret := &collision[T]{data: append([]T(nil), c.data...)}
			ret.data[i] = args.f(e, v)
			return ret, 1
		}
	}
	ret := &collision[T]{data: append(append([]T(nil), c.data...), v)}
	return ret, 0
}

func (c *collision[T]) WithFast(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if value.Equal(e, v) {
			ret := &collision[T]{data: append([]T(nil), c.data...)}
			ret.data[i] = v
			return ret, 1
		}
	}
	ret := &collision[T]{data: append(append([]T(nil), c.data...), v)}
	return ret, 0
}

func (c *collision[T]) Without(v T, _ int, _ hasher) (_ node[T], matches int) {
	for i, e := range c.data {
		if value.Equal(e, v) {
			ret := make([]T, len(c.data)-1)
			copy(ret, c.data[:i])
			copy(ret[i:], c.data[i+1:])
			return collisionCanonical[T](ret), 1
		}
	}
	return c, 0
}

func (c *collision[T]) clone() node[T] {
	return &collision[T]{data: append([]T(nil), c.data...)}
}

// collisionCanonical returns the simplest node for the given data.
func collisionCanonical[T any](data []T) node[T] {
	switch len(data) {
	case 0:
		return nil
	case 1:
		return newLeaf1(data[0])
	default:
		return &collision[T]{data: data}
	}
}
