package frozen

import "math/bits"

type buffer []full

func newBuffer(count int) *buffer {
	depth := (bits.Len64(uint64(count))+(hamtBits-1))/hamtBits + 6
	pool := make(buffer, 0, depth)
	return &pool
}

func (b *buffer) copy(f *full) *full {
	*b = append(*b, *f)
	f = &(*b)[0]
	*b = (*b)[1:]
	return f
}
