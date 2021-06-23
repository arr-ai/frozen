package tree

type unNode interface {
	Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) unNode
	Freeze() node
	Get(args *EqArgs, v interface{}, h hasher) *interface{}
	Remove(args *EqArgs, v interface{}, depth int, h hasher, matches *int) unNode

	// For internal use by unNode implementations.

	// copyTo copies all the unNode's elements into dest without triggering a
	// reallocation of the target slice. Returns true iff all elements fit.
	appendTo(dest []interface{}) []interface{}
}
