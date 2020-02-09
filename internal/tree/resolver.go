package tree

import (
	"fmt"
	"sync"
)

type Resolve func(a, b interface{}) interface{}

var UseLHS = NewResolver("lhs", func(a, _ interface{}) interface{} { return a }).Flip().Flip()
var UseRHS = NewResolver("rhs", func(_, b interface{}) interface{} { return b }).Flip().Flip()

var resolversMux sync.Mutex
var resolvers = map[string]*Resolver{}

type Resolver struct {
	name    string
	resolve Resolve
	flipped *Resolver
}

func NewResolver(name string, resolve Resolve) *Resolver {
	r := &Resolver{name: name, resolve: resolve}
	resolversMux.Lock()
	defer resolversMux.Unlock()
	if _, has := resolvers[name]; has {
		panic(fmt.Errorf("resolver %q added twice", name))
	}
	resolvers[name] = r
	return r
}

func ResolverByName(name string) *Resolver {
	return resolvers[name]
}

func (r *Resolver) Name() string {
	return r.name
}

func (r *Resolver) Resolve(a, b interface{}) interface{} {
	return r.resolve(a, b)
}

func (r *Resolver) Flip() *Resolver {
	if r.flipped == nil {
		resolve := func(a, b interface{}) interface{} { return r.Resolve(b, a) }
		r.flipped = NewResolver("~"+r.name, resolve)
		r.flipped.flipped = r
	}
	return r.flipped
}
