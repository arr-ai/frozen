package frozen

import "sync"

var unLeafPool = sync.Pool{
	New: func() interface{} {
		var buf [maxLeafLen]interface{}
		return &unLeaf{buf: &buf}
	},
}

type unLeaf struct {
	data []interface{}

	// Use a pointer to decouple the memory lifecycle from the parent's.
	buf *[maxLeafLen]interface{}
}

var _ unNode = &unLeaf{}

func newUnLeaf() *unLeaf {
	var l *unLeaf
	if usePools {
		l = unLeafPool.Get().(*unLeaf)
	} else {
		l = unLeafPool.New().(*unLeaf)
	}
	l.data = l.buf[:0]
	return l
}

func (l *unLeaf) free() {
	if usePools {
		unLeafPool.Put(l)
	}
}

func (l *unLeaf) Add(args *combineArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	if i := l.find(args.eqArgs, v); i != -1 {
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
		to.Add(defaultNPCombineArgs, e, 0, 0, nil)
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

func (l *unLeaf) Get(args *eqArgs, v interface{}, h hasher) *interface{} {
	if i := l.find(args, v); i != -1 {
		return &l.data[i]
	}
	return nil
}

func (l *unLeaf) Remove(args *eqArgs, v interface{}, depth int, h hasher, matches *int) unNode {
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

func (l *unLeaf) find(args *eqArgs, v interface{}) int {
	for i, e := range l.data {
		if args.eq(e, v) {
			return i
		}
	}
	return -1
}
