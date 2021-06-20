package kvt

import (
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/pkg/kv"
)

type unTree struct {
	root  unNode
	count int
}

func (t *unTree) Add(args *CombineArgs, v kv.KeyValue) {
	count := -(t.count + 1)
	t.root = t.Root().Add(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t *unTree) Count() int {
	return t.count
}

func (t *unTree) Gauge() depth.Gauge {
	return depth.NewGauge(t.count)
}

func (t *unTree) Get(args *EqArgs, v kv.KeyValue) *kv.KeyValue {
	return t.Root().Get(args, v, newHasher(v, 0))
}

func (t *unTree) Remove(args *EqArgs, v kv.KeyValue) {
	count := -t.count
	t.root = t.Root().Remove(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t *unTree) Root() unNode {
	if t.count == 0 {
		return unEmptyNode{}
	}
	return t.root
}
