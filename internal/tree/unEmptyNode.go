package tree

type unEmptyNode struct {
	emptyNode
}

var _ unNode = unEmptyNode{}

func (e unEmptyNode) Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	return newUnLeaf().Add(args, v, depth, h, matches)
}

func (unEmptyNode) copyTo(*unLeaf) {}

func (unEmptyNode) countUpTo(max int) int {
	return 0
}

func (unEmptyNode) Freeze() node {
	return emptyNode{}
}

func (e unEmptyNode) Get(args *EqArgs, v interface{}, h hasher) *interface{} {
	return nil
}

func (e unEmptyNode) Remove(_ *EqArgs, _ interface{}, _ int, _ hasher, _ *int) unNode {
	return e
}
