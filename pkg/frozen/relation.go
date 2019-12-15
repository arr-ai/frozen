package frozen

import "fmt"

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
	var logs []string
	logf := func(format string, args ...interface{}) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}
	logf("t=%v", t)

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
			logf("val=%s val{%v}=%v", val, attrs, val.(Set).Project(attrs))
			return val.(Set).Project(attrs)
		})
	}
	sGroup := group(s, sOnlyAttrs)
	logf("sGroup=%v", sGroup)
	tGroup := group(t, tOnlyAttrs)
	logf("tGroup=%v", tGroup)
	joiner := sGroup.Merge(tGroup, func(_, a, b interface{}) interface{} {
		return [2]Set{a.(Set), b.(Set)}
	})
	logf("joiner=%v", joiner)

	var result Set
	for i := joiner.Range(); i.Next(); {
		commonTuple := i.Key().(Map)
		if sets, ok := i.Value().([2]Set); ok {
			for j := sets[0].Range(); j.Next(); {
				sTuple := commonTuple.Update(j.Value().(Map))
				for k := sets[1].Range(); k.Next(); {
					row := sTuple.Update(k.Value().(Map))
					logf("%v", row)
					result = result.With(row)
				}
			}
		}
	}
	return result
}

// Nest ...
func (s Set) Nest(attr interface{}, attrs Set) Set {
	m := s.Reduce(func(acc, i interface{}) interface{} {
		t := i.(Map)
		key := t.Without(attrs)
		nested := acc.(Map).GetElse(key, Set{})
		return acc.(Map).With(key, nested.(Set).With(t.Project(attrs)))
	}, Map{}).(Map)
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		return acc.(Set).With(key.(Map).With(attr, val))
	}, Set{}).(Set)
}

// Unnest ...
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
