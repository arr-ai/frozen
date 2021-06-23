package tree

import "github.com/arr-ai/frozen/errors"

type unDefroster struct {
	n node
}

var _ unNode = unDefroster{}

func (d unDefroster) Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Add(args, v, depth, h, matches)
}

func (d unDefroster) copyTo(to *unLeaf, depth int) {
	panic(errors.Unimplemented)
}

func (d unDefroster) countUpTo(max int) int {
	return d.n.CountUpTo(max)
}

func (d unDefroster) Empty() bool {
	return d.n.Empty()
}

func (d unDefroster) Freeze() node {
	return d.n
}

func (d unDefroster) Get(args *EqArgs, v interface{}, h hasher) *interface{} {
	return d.n.Get(args, v, h)
}

func (d unDefroster) Remove(args *EqArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Remove(args, v, depth, h, matches)
}
