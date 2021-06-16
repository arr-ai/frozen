package frozen

type unTree struct {
	root  unNode
	count int
}

func (t *unTree) Add(args *combineArgs, v interface{}) {
	count := -(t.count + 1)
	t.root = t.Root().Add(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t *unTree) Count() int {
	return t.count
}

func (t *unTree) Gauge() parallelDepthGauge {
	return newParallelDepthGauge(t.count)
}

func (t *unTree) Get(args *eqArgs, v interface{}) *interface{} {
	return t.Root().Get(args, v, newHasher(v, 0))
}

func (t *unTree) Remove(args *eqArgs, v interface{}) {
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
