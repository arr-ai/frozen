package tree

import (
	"context"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/arr-ai/frozen/slave/proto/slave"
)

type Cloner struct {
	parallelDepth int
	wg            sync.WaitGroup
	mutate        bool
	update        func(interface{})
	clients       []slave.SlaveClient
	ctx           context.Context
}

var (
	Copier  = &Cloner{mutate: false, parallelDepth: -1}
	Mutator = &Cloner{mutate: true, parallelDepth: -1}
)

func NewCloner(mutate bool, capacity int) *Cloner {
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
	return &Cloner{
		mutate: mutate,

		// Give parallel workers O(32k) elements each to process. If
		// parallelDepth < 0, it won't parallelise.
		parallelDepth: (bits.Len64(uint64(capacity)) - maxConcurrency) / 3,

		clients: slaveClients(),
		ctx:     context.TODO(),
	}
}

func (c *Cloner) Update(v interface{}) {
	c.update(v)
}

func (c *Cloner) node(n *Node, prepared **Node) *Node {
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

func (c *Cloner) leaf(l *Leaf) *Leaf {
	switch {
	case c.mutate:
		return l
	default:
		result := *l
		return &result
	}
}

func (c *Cloner) extras(l *Leaf, capacityIncrease int) extraLeafElems {
	if c.mutate {
		return l.extras()
	}
	x := l.extras()
	x = append(make([]interface{}, 0, len(x)+capacityIncrease), x...)
	return x
}

func (c *Cloner) run(f func()) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		f()
	}()
}

func (c *Cloner) Wait() {
	c.wg.Wait()
}

func (c *Cloner) chain(update func(interface{})) {
	if next := c.update; next != nil {
		c.update = func(arg interface{}) {
			update(arg)
			next(arg)
		}
	} else {
		c.update = update
	}
}

func (c *Cloner) Counter() func() int {
	n := uintptr(0)
	c.chain(func(arg interface{}) {
		if i, ok := arg.(int); ok {
			atomic.AddUintptr(&n, uintptr(i))
		}
	})
	return func() int {
		c.Wait()
		return int(n)
	}
}

func (c *Cloner) NoneFalse() func() bool {
	someFalse := uintptr(0)
	c.chain(func(arg interface{}) {
		if b, ok := arg.(bool); ok {
			if !b {
				atomic.StoreUintptr(&someFalse, 1)
			}
		}
	})
	return func() bool {
		c.Wait()
		return someFalse == 0
	}
}
