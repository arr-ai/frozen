package frozen

type packed struct {
	mask masker
	data []node
}

func packedFromNodes(nodes *[fanout]node) packed {
	p := packed{}
	for i, n := range nodes {
		switch n := n.(type) {
		case nil, emptyNode:
		default:
			p.mask |= masker(1) << i
			p.data = append(p.data, n)
		}
	}
	return p
}

func (p packed) Get(i masker) node {
	if i.firstIsIn(p.mask) {
		return p.data[p.mask.offset(i)]
	}
	return emptyNode{}
}

func (p packed) With(i masker, n node) packed {
	i = i.first()
	index := p.mask.offset(i)

	_, empty := n.(emptyNode)
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

func (p packed) Iterator(buf []packed) Iterator {
	return newPackedIterator(buf, p)
}

func (p packed) All(parallel bool, f func(m masker, n node) bool) bool {
	return p.AllMask(p.mask, parallel, f)
}

func (p packed) AllPair(q packed, mask masker, parallel bool, f func(m masker, a, b node) bool) bool {
	return p.AllMask(mask, parallel, func(m masker, n node) bool {
		return f(m, n, q.Get(m))
	})
}

func (p packed) AllMask(mask masker, parallel bool, f func(m masker, n node) bool) bool {
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

func (p packed) Transform(parallel bool, f func(m masker, n node) node) packed {
	var nodes [fanout]node
	p.All(parallel, func(m masker, n node) bool {
		nodes[m.index()] = f(m, n)
		return true
	})
	return packedFromNodes(&nodes)
}

func (p packed) TransformPair(q packed, mask masker, parallel bool, f func(m masker, x, y node) node) packed {
	var nodes [fanout]node
	p.AllPair(q, mask, parallel, func(m masker, a, b node) bool {
		nodes[m.index()] = f(m, a, b)
		return true
	})
	return packedFromNodes(&nodes)
}

func (p packed) update(flipper masker, prefix, delta int) packed {
	result := packed{}
	result.mask = p.mask ^ flipper
	result.data = make([]node, 0, len(p.data)+delta)
	result.data = append(result.data, p.data[:prefix]...)
	return result
}
