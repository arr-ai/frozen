package frozen

type emptyNode struct{}

func (emptyNode) String() string {
	return "∅"
}

func (e emptyNode) Canonical(_ int) node {
	return e
}

func (e emptyNode) Combine(_ *combineArgs, n node, _ int, _ *int) node {
	return n
}

func (e emptyNode) CountUpTo(_ int) int {
	return 0
}

func (e emptyNode) Defrost() unNode {
	return unEmptyNode{}
}

func (e emptyNode) Difference(_ *eqArgs, _ node, _ int, _ *int) node {
	return e
}

func (e emptyNode) Equal(_ *eqArgs, n node, _ int) bool {
	return e == n
}

func (emptyNode) Get(_ *eqArgs, _ interface{}, _ hasher) *interface{} {
	return nil
}

func (e emptyNode) Intersection(_ *eqArgs, _ node, _ int, _ *int) node {
	return e
}

func (emptyNode) Iterator([]packed) Iterator {
	return emptyIterator{}
}

func (emptyNode) Reduce(_ nodeArgs, _ int, _ func(...interface{}) interface{}) interface{} {
	panic(WTF)
}

func (emptyNode) SubsetOf(_ *eqArgs, _ node, _ int) bool {
	return true
}

func (e emptyNode) Transform(_ *combineArgs, _ int, _ *int, _ func(interface{}) interface{}) node {
	return e
}

func (e emptyNode) Where(_ *whereArgs, _ int, _ *int) node {
	return e
}

func (emptyNode) With(args *combineArgs, v interface{}, depth int, h hasher, matches *int) node {
	return leaf{v}
}

func (e emptyNode) Without(_ *eqArgs, _ interface{}, _ int, _ hasher, _ *int) node {
	return e
}

type emptyIterator struct{}

func (emptyIterator) Next() bool {
	return false
}

func (emptyIterator) Value() interface{} {
	panic(WTF)
}
