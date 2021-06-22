package tree

import "github.com/arr-ai/frozen/internal/iterator"

const maxLeafLen = 8

type packer struct {
	mask masker
	data []node
}

func packerFromNodes(nodes *[fanout]node) packer {
	p := packer{}
	for i, n := range nodes {
		if n != nil && !n.Empty() {
			p.mask |= masker(1) << i
			p.data = append(p.data, n)
		}
	}
	return p
}

func (p packer) Get(i masker) node {
	if i.firstIsIn(p.mask) {
		return p.data[p.mask.offset(i)]
	}
	return theEmptyNode
}

func (p packer) With(i masker, n node) packer {
	i = i.first()
	index := p.mask.offset(i)

	empty := n.Empty()
	if existing := i.subsetOf(p.mask); existing {
		if empty {
			result := p.update(i, index, -1)
			switch index {
			case 0:
				result.data = p.data[1:]
			case len(p.data) - 1:
			default:
				result.data = append(result.data, p.data[index+1:]...)
			}
			return result
		}
		result := p.update(0, len(p.data), 0)
		result.data[index] = n
		return result
	}
	if !empty {
		result := p.update(i, index, 1)
		result.data = append(result.data, n)
		result.data = append(result.data, p.data[index:]...)
		return result
	}
	return p
}

func (p packer) Iterator(buf []packer) iterator.Iterator {
	return newPackerIterator(buf, p)
}

func (p packer) All(parallel bool, f func(m masker, n node) bool) bool {
	return p.AllMask(p.mask, parallel, f)
}

func (p packer) AllPair(q packer, mask masker, parallel bool, f func(m masker, a, b node) bool) bool {
	return p.AllMask(mask, parallel, func(m masker, n node) bool {
		return f(m, n, q.Get(m))
	})
}

func (p packer) AllMask(mask masker, parallel bool, f func(m masker, n node) bool) bool {
	if parallel {
		dones := make(chan bool, fanout)
		for m := mask; m != 0; m = m.next() {
			m := m
			go func() {
				dones <- f(m, p.Get(m))
			}()
		}
		for m := mask; m != 0; m = m.next() {
			if !<-dones {
				return false
			}
		}
	} else {
		for m := mask; m != 0; m = m.next() {
			if !f(m, p.Get(m)) {
				return false
			}
		}
	}
	return true
}

func (p packer) Transform(parallel bool, f func(m masker, n node) node) packer {
	var nodes [fanout]node
	p.All(parallel, func(m masker, n node) bool {
		nodes[m.index()] = f(m, n)
		return true
	})
	return packerFromNodes(&nodes)
}

func (p packer) TransformPair(q packer, mask masker, parallel bool, f func(m masker, x, y node) node) packer {
	var nodes [fanout]node
	p.AllPair(q, mask, parallel, func(m masker, a, b node) bool {
		nodes[m.index()] = f(m, a, b)
		return true
	})
	return packerFromNodes(&nodes)
}

func (p packer) update(flipper masker, prefix, delta int) packer {
	result := packer{}
	result.mask = p.mask ^ flipper
	result.data = make([]node, 0, len(p.data)+delta)
	result.data = append(result.data, p.data[:prefix]...)
	return result
}
