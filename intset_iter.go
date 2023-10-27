package frozen

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type intSetIterator[T integer] struct {
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
