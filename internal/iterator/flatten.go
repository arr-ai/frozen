package iterator

type flattener struct {
	ii Iterator
	i  Iterator
}

func Flatten(ii Iterator) Iterator {
	return &flattener{ii: ii, i: Empty}
}

func (i *flattener) Next() bool {
	if i.i.Next() {
		return true
	}
	if i.ii.Next() {
		i.i = i.ii.Value().(Iterator)
		return i.i.Next()
	}
	return false
}

func (i *flattener) Value() elementT {
	return i.i.Value()
}
