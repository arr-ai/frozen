package tree

import (
	"github.com/arr-ai/frozen/internal/pkg/masker"
)

const maxLeafLen = 2

type packer[T any] struct {
	mask masker.Masker
	data [fanout]node[T]
}

func (p *packer[T]) EqualPacker(q *packer[T]) bool {
	for i, a := range p.data {
		b := q.data[i]
		if (a == nil) != (b == nil) || a != nil && !a.Equal(DefaultNPEqArgs[T](), b, 0) {
			return false
		}
	}
	return true
}

func (p *packer[T]) GetChild(m masker.Masker) node[T] {
	return p.data[m.FirstIndex()]
}

func (p *packer[T]) SetChild(i int, n node[T]) {
	m := masker.NewMasker(i)
	if n == nil {
		p.mask &^= m
	} else {
		p.mask |= m
	}
	p.data[i] = n
}

func (p *packer[T]) SetNonNilChild(i int, n node[T]) {
	p.mask |= masker.NewMasker(i)
	p.data[i] = n
}

func (p *packer[T]) WithChild(i int, n node[T]) *packer[T] {
	ret := *p
	ret.SetChild(i, n)
	return &ret
}

func (p *packer[T]) Iterator(buf [][]node[T]) Iterator[T] {
	return newPackerIterator(buf, p)
}

func (p *packer[T]) updateMask() {
	var mask masker.Masker
	for i, n := range p.data {
		if n != nil {
			mask |= masker.NewMasker(i)
		}
	}
	p.mask = mask
}

func (p *packer[T]) updateMaskBit(m masker.Masker) {
	p.mask |= m
}
