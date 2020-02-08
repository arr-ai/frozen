package tree

type NodeAnalysis struct {
	nodes    int
	leaves   int
	elements int
	depths   []int

	// [depth][count(children)]count
	nodeFillses map[int]map[int]int
	leafFillses map[int]map[int]int

	hashers map[Hasher]int
}

func incrementFills(fillses map[int]map[int]int, depth, count int) {
	fills, ok := fillses[depth]
	if !ok {
		fills = map[int]int{}
		fillses[depth] = fills
	}
	fills[count] = fills[count] + 1
}

func (a *NodeAnalysis) node(n *Node, depth int) {
	a.nodes++
	incrementFills(a.nodeFillses, depth, n.mask.Count())
}

func (a *NodeAnalysis) leaf(l *Leaf, depth int) {
	a.leaves++

	count := l.count()
	a.elements += count

	for len(a.depths) <= depth {
		a.depths = append(a.depths, 0)
	}
	a.depths[depth]++

	incrementFills(a.leafFillses, depth, count)
}

func (a *NodeAnalysis) element(elem interface{}) {
	a.elements++

	if a.hashers != nil {
		h := NewHasher(elem, 0)
		a.hashers[h] = a.hashers[h] + 1
	}
}

func (n *Node) Profile(includeHasher bool) *NodeAnalysis {
	result := &NodeAnalysis{
		nodeFillses: map[int]map[int]int{},
		leafFillses: map[int]map[int]int{},
	}
	if includeHasher {
		result.hashers = map[Hasher]int{}
	}
	n.profileImpl(result, 0)
	return result
}

func (n *Node) profileImpl(a *NodeAnalysis, depth int) {
	switch {
	case n == nil:
		return
	case n.isLeaf():
		a.leaf(n.leaf(), depth)
		for i := n.leaf().iterator(); i.Next(); {
			a.element(i.Value())
		}
	default:
		a.node(n, depth)
		for mask := n.mask; mask != 0; mask = mask.Next() {
			n.children[mask.Index()].profileImpl(a, depth+1)
		}
	}
}
