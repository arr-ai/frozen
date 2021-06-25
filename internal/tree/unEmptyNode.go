package tree

type unEmptyNode struct{}

var _ unNode = unEmptyNode{}

func (e unEmptyNode) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) unNode {
	l := newUnLeaf()
	return l.Add(args, v, depth, h, matches)
}

func (unEmptyNode) appendTo(dest []elementT) []elementT {
	return dest
}

func (unEmptyNode) Freeze() *node {
	return newLeaf().Node()
}

func (e unEmptyNode) Get(args *EqArgs, v elementT, h hasher) *elementT {
	return nil
}

func (e unEmptyNode) Remove(_ *EqArgs, _ elementT, _ int, _ hasher, _ *int) unNode {
	return e
}
