package iterator

type MapIterator struct {
	i Iterator
	m func(v elementT) elementT
}

func Map(i Iterator, m func(v elementT) elementT) *MapIterator {
	return &MapIterator{i: i, m: m}
}

func (m MapIterator) Next() bool {
	return m.i.Next()
}

func (m MapIterator) Value() elementT {
	return m.m(m.i.Value())
}
