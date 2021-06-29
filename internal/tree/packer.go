package tree

const maxLeafLen = 2

type packer [fanout]node

func (p *packer) EqualPacker(q *packer) bool {
	for i, a := range p {
		b := q[i]
		if (a == nil) != (b == nil) || a != nil && !a.Equal(DefaultNPEqArgs, b, 0) {
			return false
		}
	}
	return true
}

func (p *packer) With(i int, n node) *packer {
	ret := *p
	ret[i] = n
	return &ret
}

func (p *packer) Iterator(buf [][]node) Iterator {
	return newPackerIterator(buf, p)
}
