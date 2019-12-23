package frozen

// Join returns the n-ary join of a Set of Sets.
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
	if commonAttrs.IsEmpty() {
		return s.CartesianProduct(t)
	}
	projectCommon := func(tuple interface{}) interface{} {
		return tuple.(Map).Project(commonAttrs)
	}
	sGroup := s.GroupBy(projectCommon)
	tGroup := t.GroupBy(projectCommon)
	joinedGroup := sGroup.Merge(tGroup, func(_, a, b interface{}) interface{} {
		return [2]Set{a.(Set), b.(Set)}
	})

	var b SetBuilder
	for i := joinedGroup.Range(); i.Next(); {
		if sets, ok := i.Value().([2]Set); ok {
			buildCartesianProduct(&b, Map{}, sets[:]...)
		}
	}
	return b.Finish()
}

func (s Set) CartesianProduct(t Set) Set {
	return CartesianProduct(s, t)
}

func CartesianProduct(relations ...Set) Set {
	var b SetBuilder
	buildCartesianProduct(&b, Map{}, relations...)
	return b.Finish()
}

func buildCartesianProduct(b *SetBuilder, t Map, relations ...Set) {
	if len(relations) > 0 {
		for i := relations[0].Range(); i.Next(); {
			buildCartesianProduct(b, t.Update(i.Value().(Map)), relations[1:]...)
		}
	} else {
		b.Add(t)
	}
}

// Nest returns a relation with some attributes nested as subrelations.
//
// Example:
//
//   input:
//     / c | a  \
//     | 1 | 10 |
//     | 1 | 11 |
//     | 2 | 13 |
//     | 3 | 11 |
//     | 4 | 14 |
//     | 3 | 10 |
//     \ 4 | 13 /
//
//   nest(input, ("aa": {"a"})):
//     / c | aa     \
//     | 1 | / a  \ |
//     |   | | 10 | |
//     |   | \ 11 / |
//     | 2 | / a  \ |
//     |   | \ 13 / |
//     | 3 | / a  \ |
//     |   | | 10 | |
//     |   | \ 11 / |
//     | 4 | / a  \ |
//     |   | | 13 | |
//     \   | \ 14 / /
//
func (s Set) Nest(attrAttrs Map) Set {
	keyAttrs := Intersection(attrAttrs.Values())
	return s.
		GroupBy(func(el interface{}) interface{} {
			return el.(Map).Without(keyAttrs)
		}).
		Map(func(key, group interface{}) interface{} {
			return attrAttrs.Map(func(_, attrs interface{}) interface{} {
				return group.(Set).Project(attrs.(Set))
			}).Update(key.(Map))
		}).
		Values()
}

// Unnest returns a relation with some subrelations unnested. This is the
// reverse of Nest.
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
