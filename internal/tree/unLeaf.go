package tree

type unLeaf struct {
	data  map[hasher][]interface{}
	count int
}

var _ unNode = &unLeaf{}

func newUnLeaf() *unLeaf {
	return &unLeaf{data: map[hasher][]interface{}{}}
}

func (l *unLeaf) Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) (ret unNode) {
	if vetting {
		defer vetUnNode(l)(&ret)
	}
	bucket := l.data[h]
	for i, e := range bucket {
		if args.eq(e, v) {
			*matches++
			bucket[i] = args.f(e, v)
			return l
		}
	}
	if l.count < maxLeafLen || depth >= maxTreeDepth {
		l.count++
		l.data[h] = append(bucket, v)
		return l
	}

	b := newUnBranch()
	for _, bucket := range l.data {
		for _, e := range bucket {
			b.Add(args, e, depth, newHasher(e, depth), matches)
		}
	}
	b.Add(args, v, depth, h, matches)

	return b
}

func (l *unLeaf) copyTo(to *unLeaf, depth int) {
	if vetting {
		defer vetUnNode(l)(&to)
	}
	for _, bucket := range l.data {
		for _, e := range bucket {
			h := newHasher(e, depth)
			bucket := to.data[h]
			to.data[h] = append(bucket, e)
		}
	}
	to.count += l.count
}

func (l *unLeaf) countUpTo(max int) int {
	if vetting {
		defer vetUnNode(l)()
	}
	return l.count
}

func (l *unLeaf) Freeze() node {
	if vetting {
		defer vetUnNode(l)()
	}
	ret := make(leaf, 0, l.count)
	for _, bucket := range l.data {
		ret = append(ret, bucket...)
	}
	return ret
}

func (l *unLeaf) Get(args *EqArgs, v interface{}, h hasher) *interface{} {
	if vetting {
		defer vetUnNode(l)()
	}
	bucket := l.data[h]
	for i, e := range bucket {
		if args.eq(e, v) {
			return &bucket[i]
		}
	}
	return nil
}

func (l *unLeaf) Remove(args *EqArgs, v interface{}, depth int, h hasher, matches *int) (ret unNode) {
	if vetting {
		defer vetUnNode(l)(&ret)
	}
	bucket := l.data[h]
	for i, e := range bucket {
		if args.eq(e, v) {
			*matches++
			if l.count--; l.count == 0 {
				return unEmptyNode{}
			}
			if last := len(bucket) - 1; last > 0 {
				if i < last {
					bucket[i] = bucket[last]
				}
				l.data[h] = bucket[:last]
			} else {
				delete(l.data, h)
			}
			return l
		}
	}
	return l
}
