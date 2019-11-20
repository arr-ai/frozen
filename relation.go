package frozen

func NewRelation(header []interface{}, rows ...[]interface{}) Set {
	r := NewSet()
	for _, row := range rows {
		if len(row) != len(header) {
			panic("header/row size mismatch")
		}
		t := NewMap()
		for i, h := range header {
			t = t.With(h, row[i])
		}
		r = r.With(t)
	}
	return r
}

func (s Set) Project(attrs ...interface{}) Set {
	return s.Map(func(i interface{}) interface{} {
		return i.(Map).Project(attrs)
	})
}

func (s Set) Nest(attr interface{}, attrs ...interface{}) Set {
	m := EmptyMap()
	for i := s.Range(); i.Next(); {
		t := i.Value().(Map)
		key := t.Without(attrs...)
		nested := m.ValueElseFunc(key, func() interface{} { return EmptySet() })
		m = m.With(key, nested.(Set).With(t.Project(attrs...)))
	}
	result := EmptySet()
	for i := m.Range(); i.Next(); {
		result = result.With(i.Key().(Map).With(attr, i.Value()))
	}
	return result
}
