package tree

const maxLeafLen = 8

type packer [fanout]noderef

func (p *packer) EqualPacker(q *packer) bool {
	for i, a := range p {
		b := q[i]
		if (a == nil) != (b == nil) || a != nil && !a.Equal(DefaultNPEqArgs, b, 0) {
			return false
		}
	}
	return true
}

func (p *packer) Get(i int) noderef {
	if n := p[i]; n != nil {
		return n
	}
	return theEmptyNode
}

func (p *packer) With(i int, n noderef) *packer {
	ret := *p
	if n.Empty() {
		ret[i] = nil
	} else {
		ret[i] = n
	}
	return &ret
}

func (p *packer) Iterator(buf [][]noderef) Iterator {
	return newPackerIterator(buf, p)
}
