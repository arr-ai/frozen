package tree

import "github.com/arr-ai/frozen/errors"

type unDefroster struct {
	n *node
}

var _ unNode = unDefroster{}

func (d unDefroster) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Add(args, v, depth, h, matches)
}

func (d unDefroster) appendTo([]elementT) []elementT {
	panic(errors.Unimplemented)
}

func (d unDefroster) Empty() bool {
	return d.n.Empty()
}

func (d unDefroster) Freeze() *node {
	return d.n
}

func (d unDefroster) Get(args *EqArgs, v elementT, h hasher) *elementT {
	return d.n.Get(args, v, h)
}

func (d unDefroster) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Remove(args, v, depth, h, matches)
}
