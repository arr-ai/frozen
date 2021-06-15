package frozen

type emptyNode struct{}

func (emptyNode) String() string {
	return "∅"
}

func (e emptyNode) canonical(_ int) node {
	return e
}

func (e emptyNode) combine(_ *combineArgs, n node, _ int, _ *int) node {
	return n
}

func (e emptyNode) countUpTo(_ int) int {
	return 0
}

func (e emptyNode) difference(_ *eqArgs, _ node, _ int, _ *int) node {
	return e
}

func (e emptyNode) equal(_ *eqArgs, n node, _ int) bool {
	return e == n
}

func (emptyNode) get(_ *eqArgs, _ interface{}, _ hasher) *interface{} {
	return nil
}

func (e emptyNode) intersection(_ *eqArgs, _ node, _ int, _ *int) node {
	return e
}

func (emptyNode) isSubsetOf(_ *eqArgs, _ node, _ int) bool {
	return true
}

func (emptyNode) iterator([]packed) Iterator {
	return emptyIterator{}
}

func (emptyNode) reduce(_ nodeArgs, _ int, _ func(...interface{}) interface{}) interface{} {
	panic(wtf)
}

func (e emptyNode) transform(_ *combineArgs, _ int, _ *int, _ func(interface{}) interface{}) node {
	return e
}

func (e emptyNode) where(_ *whereArgs, _ int, _ *int) node {
	return e
}

func (e emptyNode) vet() node {
	return e
}

func (emptyNode) with(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node {
	return leaf{v}
}

func (e emptyNode) without(_ *eqArgs, _ interface{}, _ int, _ hasher, _ *int) node {
	return e
}

type emptyIterator struct{}

func (emptyIterator) Next() bool {
	return false
}

func (emptyIterator) Value() interface{} {
	panic(wtf)
}
