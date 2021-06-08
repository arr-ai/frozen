package frozen

import (
	"math/bits"
	"os"
	"strconv"
	"strings"
)

type cloner struct {
	parallelDepth int
	mutate        bool
}

var (
	theCopier  = &cloner{mutate: false, parallelDepth: -1}
	theMutator = &cloner{mutate: true, parallelDepth: -1}
)

func newCloner(mutate bool, capacity int) *cloner {
	frozenConcurrency := os.Getenv("FROZEN_CONCURRENCY")
	var maxConcurrency int
	var err error
	if strings.ToLower(frozenConcurrency) == "off" {
		maxConcurrency = 1<<(bits.UintSize-1) - 1
	} else {
		maxConcurrency, err = strconv.Atoi(frozenConcurrency)
		if err != nil {
			maxConcurrency = 15
		}
	}
	return &cloner{
		mutate: mutate,

		// Give parallel workers O(32k) elements each to process. If
		// parallelDepth < 0, it won't parallelise.
		parallelDepth: (bits.Len64(uint64(capacity)) - maxConcurrency) / 3,
	}
}

func (c *cloner) node(n *node, prepared **node) *node {
	switch {
	case c.mutate:
		return n
	case *prepared != nil:
		return *prepared
	default:
		result := *n
		*prepared = &result
		return &result
	}
}

func (c *cloner) leaf(l *leaf) *leaf {
	switch {
	case c.mutate:
		return l
	default:
		result := *l
		return &result
	}
}

func (c *cloner) extras(l *leaf, capacityIncrease int) extraLeafElems {
	if c.mutate {
		return l.extras()
	}
	x := l.extras()
	x = append(make([]interface{}, 0, len(x)+capacityIncrease), x...)
	return x
}
