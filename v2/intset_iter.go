package frozen

import (
	"golang.org/x/exp/constraints"

	"github.com/arr-ai/frozen/v2/internal/pkg/iterator"
)

type intSetIterator[T constraints.Integer] struct {
	blockIter      MapIterator[T, *cellBlock]
	block          []cellMask
	firstIntInCell T
}

func (i *intSetIterator[T]) Next() bool {
	if len(i.block) > 0 {
		if i.block[0] != 0 {
			i.block[0] &= i.block[0] - 1
		}

		for i.block[0] == 0 {
			i.firstIntInCell += cellBits
			i.block = i.block[1:]
			if len(i.block) == 0 {
				break
			}
		}
	}

	if len(i.block) == 0 {
		if !i.blockIter.Next() {
			return false
		}
		i.firstIntInCell = i.blockIter.Key() << blockShift
		block := *i.blockIter.Value()
		for i.block = block[:]; i.block[0] == 0; i.block = i.block[1:] {
			i.firstIntInCell += cellBits
		}
	}
	return len(i.block) > 0
}

func (i *intSetIterator[T]) Value() T {
	return i.firstIntInCell + T(iterator.BitIterator(i.block[0]).Index())
}
