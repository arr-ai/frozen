package frozen

// IntLess dictates the order of two elements.
type IntLess func(a, b int) bool

type IntIterator interface {
	Next() bool
	Value() int
}

type intSetIterator struct {
	blockIter      *MapIterator[int, *cellBlock]
	block          []cellMask
	firstIntInCell int
}

func (i *intSetIterator) Next() bool {
	if len(i.block) > 0 && i.block[0] != 0 {
		i.block[0] &= i.block[0] - 1
	}

	if len(i.block) > 0 && i.block[0] == 0 {
		for ; len(i.block) != 0 && i.block[0] == 0; i.block = i.block[1:] {
			i.firstIntInCell += cellBits
		}
	}

	if len(i.block) == 0 {
		if !i.blockIter.Next() {
			return false
		}
		i.firstIntInCell = i.blockIter.Key() * blockBits
		block := i.blockIter.Value()
		for i.block = block[:]; i.block[0] == 0; i.block = i.block[1:] {
			i.firstIntInCell += cellBits
		}
	}
	return len(i.block) > 0
}

func (i *intSetIterator) Value() int {
	return i.firstIntInCell + BitIterator(i.block[0]).Index()
}
