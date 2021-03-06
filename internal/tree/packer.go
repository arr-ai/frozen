package tree

import "github.com/arr-ai/frozen/internal/pkg/masker"

const maxLeafLen = 2

type packer struct {
	mask masker.Masker
	data [fanout]node
}

func (p *packer) EqualPacker(q *packer) bool {
	for i, a := range p.data {
		b := q.data[i]
		if (a == nil) != (b == nil) || a != nil && !a.Equal(DefaultNPEqArgs, b, 0) {
			return false
		}
	}
	return true
}

func (p *packer) GetChild(m masker.Masker) node {
	return p.data[m.FirstIndex()]
}

func (p *packer) SetChild(i int, n node) {
	m := masker.NewMasker(i)
	if n == nil {
		p.mask &^= m
	} else {
		p.mask |= m
	}
	p.data[i] = n
}

func (p *packer) SetNonNilChild(i int, n node) {
	p.mask |= masker.NewMasker(i)
	p.data[i] = n
}

func (p *packer) WithChild(i int, n node) *packer {
	ret := *p
	ret.SetChild(i, n)
	return &ret
}

func (p *packer) Iterator(buf [][]node) Iterator {
	return newPackerIterator(buf, p)
}

func (p *packer) updateMask() {
	var mask masker.Masker
	for i, n := range p.data {
		if n != nil {
			mask |= masker.NewMasker(i)
		}
	}
	p.mask = mask
}

func (p *packer) updateMaskBit(m masker.Masker) {
	p.mask |= m
}
