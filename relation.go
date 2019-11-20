package frozen

func NewRelation(header []interface{}, rows ...[]interface{}) Set {
	r := EmptySet()
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

func (s Set) Unnest(attrs ...interface{}) Set {
	for _, attr := range attrs {
		s = s.Reduce(func(acc, i interface{}) interface{} {
			t := i.(Map)
			key := t.Without(attr)
			return acc.(Set).Union(t.MustGet(attr).(Set).Reduce(func(acc, i interface{}) interface{} {
				return acc.(Set).With(key.Update(i.(Map)))
			}, EmptySet()).(Set))
		}, EmptySet()).(Set)
	}
	return s
}
