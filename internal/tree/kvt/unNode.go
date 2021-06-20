package kvt

import (
	"github.com/arr-ai/frozen/pkg/kv"
)

type unNode interface {
	Add(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode
	Freeze() node
	Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue
	Remove(args *EqArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode

	// For internal use by unNode implementations.
	copyTo(n *unLeaf)
	countUpTo(max int) int
}
