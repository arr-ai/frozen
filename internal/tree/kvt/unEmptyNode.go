package kvt

import (
	"github.com/arr-ai/frozen/pkg/kv"
)

type unEmptyNode struct{}

var _ unNode = unEmptyNode{}

func (e unEmptyNode) Add(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	return newUnLeaf().Add(args, v, depth, h, matches)
}

func (unEmptyNode) copyTo(*unLeaf) {}

func (unEmptyNode) countUpTo(max int) int {
	return 0
}

func (unEmptyNode) Freeze() node {
	return leaf(nil)
}

func (e unEmptyNode) Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue {
	return nil
}

func (e unEmptyNode) Remove(_ *EqArgs, _ kv.KeyValue, _ int, _ hasher, _ *int) unNode {
	return e
}