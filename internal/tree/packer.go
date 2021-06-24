package tree

import (
	"sync"

	"github.com/arr-ai/frozen/internal/iterator"
)

const maxLeafLen = 8

type packer [fanout]node

func (p *packer) Get(i int) node {
	if n := p[i]; n != nil {
		return n
	}
	return theEmptyNode
}

func (p packer) With(i int, n node) packer {
	ret := p
	if n.Empty() {
		ret[i] = nil
	} else {
		ret[i] = n
	}
	return ret
}

func (p *packer) Iterator(buf [][]node) iterator.Iterator {
	return newPackerIterator(buf, p)
}

func (p *packer) TransformPair(q *packer, parallel bool, f func(i int, x, y node) node) packer {
	var nodes [fanout]node
	if parallel {
		var wg sync.WaitGroup
		for i := range p {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				nodes[i] = f(i, p.Get(i), q.Get(i))
			}()
		}
		wg.Wait()
	} else {
		for i := range p {
			nodes[i] = f(i, p.Get(i), q.Get(i))
		}
	}
	return nodes
}
