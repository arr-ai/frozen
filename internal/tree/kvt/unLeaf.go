package kvt

import (
	"sync"

	"github.com/arr-ai/frozen/internal/pool"
	"github.com/arr-ai/frozen/pkg/kv"
)

var unLeafPool = sync.Pool{
	New: func() interface{} {
		pool.ThePoolStats.New("unLeaf")
		var buf [maxLeafLen]kv.KeyValue
		return &unLeaf{buf: &buf}
	},
}

var unLeafPool0 = sync.Pool{
	New: func() interface{} {
		return &unLeaf{}
	},
}

type unLeaf struct {
	data []kv.KeyValue

	// Use a pointer to decouple the memory lifecycle from the parent's.
	buf *[maxLeafLen]kv.KeyValue
}

var _ unNode = &unLeaf{}

func newUnLeaf() *unLeaf {
	var l *unLeaf
	if pool.UsePools {
		l = unLeafPool.Get().(*unLeaf)
		pool.ThePoolStats.Get("unLeaf")
	} else {
		l = unLeafPool.New().(*unLeaf)
	}
	l.data = l.buf[:0]
	return l
}

func newUnLeaf0() *unLeaf {
	var l *unLeaf
	if pool.UsePools {
		l = unLeafPool0.Get().(*unLeaf)
	} else {
		l = unLeafPool0.New().(*unLeaf)
	}
	return l
}

func (l *unLeaf) free() {
	if pool.UsePools {
		unLeafPool.Put(l)
		pool.ThePoolStats.Put("unLeaf")
	}
}

func (l *unLeaf) Add(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	if i := l.find(args.EqArgs, v); i != -1 {
		*matches++
		l.data[i] = args.f(l.data[i], v)
		return l
	}
	if len(l.data) <= maxLeafLen-1 || depth >= maxTreeDepth {
		l.data = append(l.data, v)
		return l
	}

	b := newUnBranch()
	for _, e := range l.data {
		b.Add(args, e, depth, newHasher(e, depth), matches)
	}
	b.Add(args, v, depth, newHasher(v, depth), matches)

	l.free()

	return b
}

func (l *unLeaf) copyTo(to *unLeaf) {
	for _, e := range l.data {
		to.Add(DefaultNPCombineArgs, e, 0, 0, nil)
	}
}

func (l *unLeaf) countUpTo(max int) int {
	return len(l.data)
}

func (l *unLeaf) Freeze() node {
	if len(l.data) == maxLeafLen {
		return leaf(l.data)
	}
	result := append(make(leaf, 0, len(l.data)), l.data...)
	l.free()
	return result
}

func (l *unLeaf) Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue {
	if i := l.find(args, v); i != -1 {
		return &l.data[i]
	}
	return nil
}

func (l *unLeaf) Remove(args *EqArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	if i := l.find(args, v); i != -1 {
		*matches++
		last := len(l.data) - 1
		if last == 0 {
			l.free()
			return unEmptyNode{}
		}
		if i < last {
			l.data[i] = l.data[last]
		}
		l.data = l.data[:last]
		return l
	}
	return l
}

func (l *unLeaf) find(args *EqArgs, v kv.KeyValue) int {
	for i, e := range l.data {
		if args.eq(e, v) {
			return i
		}
	}
	return -1
}
