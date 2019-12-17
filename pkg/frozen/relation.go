package frozen

// Join returns the n-ary join of a Set of Sets.
// TODO: Maybe implement directly instead of chaining binary joins.
func Join(relations Set) Set {
	if i := relations.Range(); i.Next() {
		result := i.Value().(Set)
		for i.Next() {
			result = result.Join(i.Value().(Set))
		}
		return result
	}
	panic("Cannot join no sets")
}

// Union returns the n-ary union of a Set of Sets.
// TODO: Maybe implement directly instead of chaining binary unions.
func Union(relations Set) Set {
	var result Set
	for i := relations.Range(); i.Next(); {
		result = result.Union(i.Value().(Set))
	}
	return result
}

// NewRelation returns a new set of maps, each having the same keys.
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

// Project returns a Set with the result of projecting each map.
func (s Set) Project(attrs Set) Set {
	return s.Map(func(i interface{}) interface{} {
		return i.(Map).Project(attrs)
	})
}

// Join returns all {x, y, z} such that s has {x, y} and t has {y, z}.
// x, y and z represent sets of keys:
//   x: keys unique to maps in s
//   y: keys common to maps in both
//   z: keys unique to maps in t
// It is assumed that all maps in s have the same keys and likewise for t.
func (s Set) Join(t Set) Set {
	if s.IsEmpty() || t.IsEmpty() {
		return Set{}
	}
	sAttrs := s.Any().(Map).Keys()
	tAttrs := t.Any().(Map).Keys()
	commonAttrs := sAttrs.Intersection(tAttrs)
	sOnlyAttrs := sAttrs.Minus(commonAttrs)
	tOnlyAttrs := tAttrs.Minus(commonAttrs)
	group := func(s Set, attrs Set) Map {
		return s.GroupBy(func(tuple interface{}) interface{} {
			return tuple.(Map).Project(commonAttrs)
		}).Map(func(_, val interface{}) interface{} {
			return val.(Set).Project(attrs)
		})
	}
	sGroup := group(s, sOnlyAttrs)
	tGroup := group(t, tOnlyAttrs)
	joiner := sGroup.Merge(tGroup, func(_, a, b interface{}) interface{} {
		return [2]Set{a.(Set), b.(Set)}
	})

	var result Set
	for i := joiner.Range(); i.Next(); {
		commonTuple := i.Key().(Map)
		if sets, ok := i.Value().([2]Set); ok {
			for j := sets[0].Range(); j.Next(); {
				sTuple := commonTuple.Update(j.Value().(Map))
				for k := sets[1].Range(); k.Next(); {
					row := sTuple.Update(k.Value().(Map))
					result = result.With(row)
				}
			}
		}
	}
	return result
}

// Nest ...
func (s Set) Nest(attrAttrs Map) Set {
	var mb MapBuilder
	keyAttrs := Union(attrAttrs.Values())
	nestAttrs := attrAttrs.Keys()
	for i := s.Range(); i.Next(); {
		t := i.Value().(Map)
		key := t.Without(keyAttrs)
		var msb Map // Map<?, *SetBuilder>
		if val, has := mb.Get(key); has {
			msb = val.(Map)
		} else {
			msb = NewMapFromKeys(nestAttrs, func(_ interface{}) interface{} {
				return &SetBuilder{}
			})
			mb.Put(key, msb)
		}
		for a := attrAttrs.Range(); a.Next(); {
			msb.MustGet(a.Key()).(*SetBuilder).Add(t.Project(a.Value().(Set)))
		}
	}

	var sb SetBuilder
	for i := mb.Finish().Range(); i.Next(); {
		key := i.Key().(Map)
		for j := i.Value().(Map).Range(); j.Next(); {
			key = key.With(j.Key(), j.Value().(*SetBuilder).Finish())
		}
		sb.Add(key)
	}
	return sb.Finish()
}

// Unnest ...
func (s Set) Unnest(attrs Set) Set {
	var b SetBuilder
	for i := s.Range(); i.Next(); {
		t := i.Value().(Map)
		for j := Join(t.Project(attrs).Values().With(NewSet(t.Without(attrs)))).Range(); j.Next(); {
			b.Add(j.Value())
		}
	}
	return b.Finish()
}
