package iterator

type SliceIterator struct {
	slice []elementT
	index int
}

func NewSliceIterator(slice []elementT) *SliceIterator {
	return &SliceIterator{slice: slice, index: -1}
}

func (i *SliceIterator) Next() bool {
	i.index++
	return i.index < len(i.slice)
}

func (i *SliceIterator) Value() elementT {
	return i.slice[i.index]
}
