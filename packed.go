package frozen

type packed struct {
	mask masker
	data []node
}

func (p packed) new(flipper masker, delta int) packed {
	var result packed
	result.mask = p.mask ^ flipper
	result.data = make([]node, 0, len(p.data)+delta)
	return result
}

func fromNodes(nodes *[fanout]node) packed {
	var p packed
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

// func (p packed) toNodes() *[fanout]node {
// 	var result [fanout]node
// 	for m := masker(1)<<fanout - 1; m != 0; m = m.next() {
// 		result[m.index()] = p.get(m)
// 	}
// 	return &result
// }

// func (p packed) empty() bool {
// 	return p.mask == 0
// }

// func (p packed) count() int {
// 	return p.mask.count()
// }

func (p packed) get(i masker) node {
	if i.firstIsIn(p.mask) {
		return p.data[p.mask.offset(i)]
	}
	return emptyNode{}
}

func (p packed) with(i masker, n node) packed {
	i = i.first()
	index := p.mask.offset(i)
	if existing := i.subsetOf(p.mask); existing {
		switch n := n.(type) {
		case emptyNode:
			result := p.new(i, -1)
			result.data = append(result.data, p.data[:index]...)
			result.data = append(result.data, p.data[index+1:]...)
			return result
		default:
			result := p.new(0, 0)
			result.data = append(result.data, p.data...)
			result.data[index] = n
			return result
		}
	} else {
		switch n := n.(type) {
		case emptyNode:
			return p
		default:
			result := p.new(i, 1)
			result.data = append(result.data, p.data[:index]...)
			result.data = append(result.data, n)
			result.data = append(result.data, p.data[index:]...)
			return result
		}
	}
}

func (p packed) iterator(buf []packed) Iterator {
	return newPackedIterator(buf, p)
}

func (p packed) all(parallel bool, f func(m masker, n node) bool) bool {
	return p.allMask(p.mask, parallel, f)
}

func (p packed) allPair(q packed, mask masker, parallel bool, f func(m masker, a, b node) bool) bool {
	return p.allMask(mask, parallel, func(m masker, n node) bool {
		return f(m, n, q.get(m))
	})
}

func (p packed) allMask(mask masker, parallel bool, f func(m masker, n node) bool) bool {
	if parallel {
		dones := make(chan bool, fanout)
		for m := mask; m != 0; m = m.next() {
			m := m
			go func() {
				dones <- f(m, p.get(m))
			}()
		}
		for m := mask; m != 0; m = m.next() {
			if !<-dones {
				return false
			}
		}
	} else {
		for m := mask; m != 0; m = m.next() {
			if !f(m, p.get(m)) {
				return false
			}
		}
	}
	return true
}

func (p packed) transform(parallel bool, f func(m masker, n node) node) packed {
	var nodes [fanout]node
	p.all(parallel, func(m masker, n node) bool {
		nodes[m.index()] = f(m, n)
		return true
	})
	return fromNodes(&nodes)
}

func (p packed) transformPair(q packed, mask masker, parallel bool, f func(m masker, x, y node) node) packed {
	var nodes [fanout]node
	p.allPair(q, mask, parallel, func(m masker, a, b node) bool {
		nodes[m.index()] = f(m, a, b)
		return true
	})
	return fromNodes(&nodes)
}
