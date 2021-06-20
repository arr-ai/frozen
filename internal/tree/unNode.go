package tree

type unNode interface {
	Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) unNode
	Freeze() node
	Get(args *EqArgs, v interface{}, h hasher) *interface{}
	Remove(args *EqArgs, v interface{}, depth int, h hasher, matches *int) unNode

	// For internal use by unNode implementations.
	copyTo(n *unLeaf)
	countUpTo(max int) int
}
