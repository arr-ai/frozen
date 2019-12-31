package frozen

import (
	"math/bits"
	"sync"
	"sync/atomic"
)

type cloner struct {
	parallelDepth int
	wg            sync.WaitGroup
	mutate        bool
	update        func(interface{})
}

var (
	theCopier  = &cloner{mutate: false, parallelDepth: -1}
	theMutator = &cloner{mutate: true, parallelDepth: -1}
)

func newCloner(mutate bool, capacity int) *cloner {
	return &cloner{
		mutate: mutate,

		// Give parallel workers O(32k) elements each to process. If
		// parallelDepth < 0, it won't parallelise.
		parallelDepth: (bits.Len64(uint64(capacity)) - 15) / 3,
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

func (c *cloner) run(f func()) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		f()
	}()
}

func (c *cloner) wait() {
	c.wg.Wait()
}

func (c *cloner) chain(update func(interface{})) {
	if next := c.update; next != nil {
		c.update = func(arg interface{}) {
			update(arg)
			next(arg)
		}
	} else {
		c.update = update
	}
}

func (c *cloner) counter() func() int {
	n := uintptr(0)
	c.chain(func(arg interface{}) {
		if i, ok := arg.(int); ok {
			atomic.AddUintptr(&n, uintptr(i))
		}
	})
	return func() int {
		c.wait()
		return int(n)
	}
}

func (c *cloner) noneFalse() func() bool {
	someFalse := uintptr(0)
	c.chain(func(arg interface{}) {
		if b, ok := arg.(bool); ok {
			if !b {
				atomic.StoreUintptr(&someFalse, 1)
			}
		}
	})
	return func() bool {
		c.wait()
		return someFalse == 0
	}
}
