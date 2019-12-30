package frozen

type cloner struct {
	nodePool   []node
	leafPool   []leaf
	nodeGrowth int
	leafGrowth int
	mutate     bool
}

var (
	theCopier  = &cloner{mutate: false}
	theMutator = &cloner{mutate: true}
)

func newCloner(capacity int) *cloner {
	return &cloner{
		// TODO: Reeducate these guesses.
		nodePool:   make([]node, 1, capacity),
		leafPool:   make([]leaf, 1, capacity/2),
		nodeGrowth: capacity/8 + 1,
		leafGrowth: capacity/16 + 1,
	}
}

func (c cloner) node(n *node, prepared **node) *node {
	switch {
	case c.mutate:
		return n
	case *prepared != nil:
		return *prepared
	case c.nodePool != nil:
		*prepared = &c.nodePool[:1][0]
		if cap(c.nodePool) > 1 {
			c.nodePool = c.nodePool[1:2]
		} else {
			c.nodePool = make([]node, 1, c.nodeGrowth)
		}
		**prepared = *n
		return *prepared
	default:
		result := *n
		*prepared = &result
		return &result
	}
}

func (c cloner) leaf(l *leaf) *leaf {
	switch {
	case c.mutate:
		return l
	case c.leafPool != nil:
		result := &c.leafPool[0]
		if cap(c.leafPool) > 1 {
			c.leafPool = c.leafPool[1:2]
		} else {
			c.leafPool = make([]leaf, 1, c.leafGrowth)
		}
		*result = *l
		return result
	default:
		result := *l
		return &result
	}
}

func (c cloner) extras(l *leaf, capacityIncrease int) extraLeafElems {
	if c.mutate {
		return l.extras()
	}
	x := l.extras()
	x = append(make([]interface{}, 0, len(x)+capacityIncrease), x...)
	return x
}
