package frozen

func NewRelation(header []interface{}, rows ...[]interface{}) Set {
	r := Set{}
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

func (s Set) Project(attrs Set) Set {
	return s.Map(func(i interface{}) interface{} {
		return i.(Map).Project(attrs)
	})
}

func (s Set) Nest(attr interface{}, attrs Set) Set {
	m := s.Reduce(func(acc, i interface{}) interface{} {
		t := i.(Map)
		key := t.Without(attrs)
		nested := acc.(Map).ValueElse(key, Set{})
		return acc.(Map).With(key, nested.(Set).With(t.Project(attrs)))
	}, Map{}).(Map)
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		return acc.(Set).With(key.(Map).With(attr, val))
	}, Set{}).(Set)
}

func (s Set) Unnest(attrs Set) Set {
	for a := attrs.Range(); a.Next(); {
		attr := a.Value()
		attrAsSet := NewSet(attr)
		s = s.Reduce(func(acc, i interface{}) interface{} {
			t := i.(Map)
			key := t.Without(attrAsSet)
			return acc.(Set).Union(t.MustGet(attr).(Set).Reduce(func(acc, i interface{}) interface{} {
				return acc.(Set).With(key.Update(i.(Map)))
			}, Set{}).(Set))
		}, Set{}).(Set)
	}
	return s
}
