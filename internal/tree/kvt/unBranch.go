package kvt

import (
	"sync"

	"github.com/arr-ai/frozen/internal/pool"
	"github.com/arr-ai/frozen/pkg/kv"
)

var unBranchPool = sync.Pool{
	New: func() interface{} {
		pool.ThePoolStats.New("unBranch")
		return &unBranch{}
	},
}

var unBranchPrototype = func() unBranch {
	var b unBranch
	for i := range b.p {
		b.p[i] = unEmptyNode{}
	}
	return b
}()

type unBranch struct {
	p [fanout]unNode
}

var _ unNode = &unBranch{}

func newUnBranch() *unBranch {
	if !pool.UsePools {
		return unBranchPool.New().(*unBranch)
	}
	b := unBranchPool.Get().(*unBranch)
	pool.ThePoolStats.Get("unBranch")
	*b = unBranchPrototype
	return b
}

func (b *unBranch) free() {
	if pool.UsePools {
		pool.ThePoolStats.Put("unBranch")
		unBranchPool.Put(b)
	}
}

func (b *unBranch) Add(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	i := h.hash()
	n := b.p[i]
	if n == nil {
		n = unEmptyNode{}
	}
	b.p[i] = n.Add(args, v, depth+1, h.next(), matches)
	return b
}

func (b *unBranch) copyTo(n *unLeaf) {
	for _, e := range b.p {
		if e != nil {
			e.copyTo(n)
		}
	}
}

func (b *unBranch) countUpTo(max int) int {
	total := 0
	for _, e := range b.p {
		if e != nil {
			total += e.countUpTo(max)
			if total >= max {
				break
			}
		}
	}
	return total
}

func (b *unBranch) Freeze() node {
	var mask masker
	for i, n := range b.p {
		switch n.(type) {
		case nil, unEmptyNode:
		default:
			mask |= newMasker(i)
		}
	}
	data := make([]node, 0, mask.count())
	for m := mask; m != 0; m = m.next() {
		data = append(data, b.p[m.index()].Freeze())
	}
	b.free()
	return &branch{p: packer{mask: mask, data: data}}
}

func (b *unBranch) Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue {
	if n := b.p[h.hash()]; n != nil {
		return n.Get(args, v, h.next())
	}
	return nil
}

func (b *unBranch) Remove(args *EqArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	i := h.hash()
	if n := b.p[i]; n != nil {
		b.p[i] = b.p[i].Remove(args, v, depth+1, h.next(), matches)
		if n := b.countUpTo(maxLeafLen + 1); n <= maxLeafLen {
			l := newUnLeaf()
			b.copyTo(l)
			b.free()
			return l
		}
	}
	return b
}