package frozen

type unNode interface {
	Add(args *combineArgs, v interface{}, depth int, h hasher, matches *int) unNode
	Freeze() node
	Get(args *eqArgs, v interface{}, h hasher) *interface{}
	Remove(args *eqArgs, v interface{}, depth int, h hasher, matches *int) unNode

	// For internal use by unNode implementations.
	copyTo(n *unLeaf)
	countUpTo(max int) int
}
