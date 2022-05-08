package frozen

import (
	"golang.org/x/exp/constraints"
)

type intSetIterator[T constraints.Integer] struct {
	cellIter       MapIterator[T, cellMask]
	cell           cellMask
	firstIntInCell T
}

func (i *intSetIterator[T]) Next() bool {
	if i.cell = i.cell.next(); i.cell == 0 {
		if !i.cellIter.Next() {
			return false
		}
		i.cell = i.cellIter.Value()
		i.firstIntInCell = i.cellIter.Key()
	}
	return true
}

func (i *intSetIterator[T]) Value() T {
	return i.firstIntInCell<<cellShift + T(i.cell.index())
}
