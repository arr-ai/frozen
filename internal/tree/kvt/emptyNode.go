package kvt

import (
	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/iterator/kvi"
	"github.com/arr-ai/frozen/pkg/kv"
)

type emptyNode struct{}

func (emptyNode) String() string {
	return "∅"
}

func (e emptyNode) Canonical(_ int) node {
	return e
}

func (e emptyNode) Combine(_ *CombineArgs, n node, _ int, _ *int) node {
	return n
}

func (e emptyNode) CountUpTo(_ int) int {
	return 0
}

func (e emptyNode) Defrost() unNode {
	return unEmptyNode{}
}

func (e emptyNode) Difference(_ *EqArgs, _ node, _ int, _ *int) node {
	return e
}

func (e emptyNode) Equal(_ *EqArgs, n node, _ int) bool {
	return e == n
}

func (emptyNode) Get(_ *EqArgs, _ kv.KeyValue, _ hasher) *kv.KeyValue {
	return nil
}

func (e emptyNode) Intersection(_ *EqArgs, _ node, _ int, _ *int) node {
	return e
}

func (emptyNode) Iterator([]packer) kvi.Iterator {
	return kvi.Empty
}

func (emptyNode) Reduce(_ NodeArgs, _ int, _ func(...kv.KeyValue) kv.KeyValue) kv.KeyValue {
	panic(errors.WTF)
}

func (emptyNode) SubsetOf(_ *EqArgs, _ node, _ int) bool {
	return true
}

func (e emptyNode) Transform(_ *CombineArgs, _ int, _ *int, _ func(kv.KeyValue) kv.KeyValue) node {
	return e
}

func (e emptyNode) Where(_ *WhereArgs, _ int, _ *int) node {
	return e
}

func (emptyNode) With(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) node {
	return leaf{v}
}

func (e emptyNode) Without(_ *EqArgs, _ kv.KeyValue, _ int, _ hasher, _ *int) node {
	return e
}
