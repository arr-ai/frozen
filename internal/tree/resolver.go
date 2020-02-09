package tree

type Resolve func(a, b interface{}) interface{}

type Resolver struct {
	resolve Resolve
	flipped *Resolver
}

func NewResolver(resolve Resolve) *Resolver {
	return &Resolver{resolve: resolve}
}

func (r *Resolver) Resolve(a, b interface{}) interface{} {
	return r.resolve(a, b)
}

func (r *Resolver) Flip() *Resolver {
	if r.flipped == nil {
		r.flipped = &Resolver{
			resolve: func(a, b interface{}) interface{} { return r.Resolve(b, a) },
			flipped: r,
		}
	}
	return r.flipped
}
