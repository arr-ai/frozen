package lazy

type mapperSet struct {
	baseSet
	src Set
	m   Mapper
}

func mapper(set Set, m Mapper) Set {
	s := &mapperSet{src: set, m: m}
	s.baseSet.set = s
	return memo(s)
}

func (s *mapperSet) Range() SetIterator {
	return &mapperSetIterator{i: s.src.Range(), m: s.m}
}

type mapperSetIterator struct {
	i SetIterator
	m Mapper
}

func (s *mapperSetIterator) Next() bool {
	return s.i.Next()
}

func (s *mapperSetIterator) Value() any {
	return s.m(s.i.Value())
}
