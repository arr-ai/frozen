package rel

import (
	"log"

	"github.com/arr-ai/frozen"
)

type (
	Tuple           = frozen.Map[string, any]
	Relation        = frozen.Set[Tuple]
	RelationBuilder = frozen.SetBuilder[Tuple]
)

func NewTuple(kvs ...frozen.KeyValue[string, any]) Tuple {
	return frozen.NewMap(kvs...)
}

var trueSet = frozen.NewSet(NewTuple())

// New returns a new relation.
func New(header []string, tuples ...[]any) Relation {
	r := Relation{}
	for _, row := range tuples {
		if len(row) != len(header) {
			panic("header/row size mismatch")
		}
		t := NewTuple()
		for i, h := range header {
			t = t.With(h, row[i])
		}
		r = r.With(t)
	}
	return r
}

// Project returns a Set with the result of projecting each map.
func Project(s Relation, attrs frozen.Set[string]) Relation {
	return frozen.SetMap(s, func(t Tuple) Tuple {
		return t.Project(attrs)
	})
}

// Join returns all {x, y, z} such that s has {x, y} and t has {y, z}.
// x, y and z represent sets of keys:
//   x: keys unique to maps in s
//   y: keys common to maps in both
//   z: keys unique to maps in t
// It is assumed that all maps in s have the same keys and likewise for t.
func Join(relations ...Relation) Relation {
	s := relations[0]
	for _, t := range relations[1:] {
		s = join(s, t)
	}
	return s
}

func join(s, t Relation) Relation {
	if s.IsEmpty() || t.IsEmpty() {
		return s
	}
	if s.Equal(trueSet) {
		return t
	}
	if t.Equal(trueSet) {
		return s
	}
	sAttrs := s.Any().Keys()
	tAttrs := t.Any().Keys()
	commonAttrs := sAttrs.Intersection(tAttrs)
	if commonAttrs.IsEmpty() {
		return CartesianProduct(s, t)
	}
	projectCommon := func(t Tuple) frozen.Map[string, any] {
		return t.Project(commonAttrs)
	}
	sGroup := frozen.SetGroupBy(s, projectCommon)
	tGroup := frozen.SetGroupBy(t, projectCommon)
	commonKeys := sGroup.Keys().Intersection(tGroup.Keys())

	var rb RelationBuilder
	for i := commonKeys.Range(); i.Next(); {
		key := i.Value()
		a, has := sGroup.Get(key)
		if !has {
			panic("wat?")
		}
		b, has := tGroup.Get(key)
		if !has {
			panic("wat?")
		}
		buildCartesianProduct(&rb, Tuple{}, a, b)
	}
	return rb.Finish()
}

func CartesianProduct(relations ...Relation) Relation {
	var b RelationBuilder
	buildCartesianProduct(&b, Tuple{}, relations...)
	return b.Finish()
}

func buildCartesianProduct(b *RelationBuilder, t Tuple, relations ...Relation) {
	if len(relations) > 0 {
		for i := relations[0].Range(); i.Next(); {
			buildCartesianProduct(b, t.Update(i.Value()), relations[1:]...)
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
//      _c_ _a__
//     |_1_|_10_|
//     |_1_|_11_|
//     |_2_|_13_|
//     |_3_|_11_|
//     |_4_|_14_|
//     |_3_|_10_|
//     |_4_|_13_|
//
//   nest(input, {aa: {a}}):
//      _c_ ___aa___
//     | 1 |  _a__  |
//     |   | |_10_| |
//     |___|_|_11_|_|
//     | 2 |  _a__  |
//     |___|_|_13_|_|
//     | 3 |  _a__  |
//     |   | |_10_| |
//     |___|_|_11_|_|
//     | 4 |  _a__  |
//     |   | |_13_| |
//     |___|_|_14_|_|
//
func Nest(s Relation, attrAttrs frozen.Map[string, frozen.Set[string]]) Relation {
	log.Print("s = ", s)
	// attrAttrs = {aa: {a}}

	// {a}
	keyAttrs := frozen.Intersection(attrAttrs.Values().Elements()...)
	log.Print("keyAttrs = ", keyAttrs)

	// {
	//   {c: 1}: {{a: 10, c: 1}, {a: 11, c: 1}},
	//   {c: 2}: {{a: 13, c: 2}},
	//   {c: 3}: {{a: 10, c: 3}, {a: 11, c: 3}},
	//   {c: 4}: {{a: 13, c: 4}, {a: 14, c: 4}},
	// }
	grouped := frozen.SetGroupBy(s, func(el Tuple) Tuple {
		return el.Without(keyAttrs)
	})
	log.Print("grouped = ", grouped)

	// {
	//   {c: 1}: {c: 1, aa: {{a: 10}, {a: 11}}},
	//   {c: 2}: {c: 2, aa: {{a: 13}}},
	//   {c: 3}: {c: 3, aa: {{a: 10}, {a: 11}}},
	//   {c: 4}: {c: 4, aa: {{a: 13}, {a: 14}}},
	// }
	mapped := frozen.MapMap(grouped, func(key Tuple, group Relation) Tuple {
		// {c: 1} => {aa: {{a: 10}, {a: 11}}}
		a := frozen.MapMap(attrAttrs, func(_ string, attrs frozen.Set[string]) any {
			return Project(group, attrs)
		})
		// {c: 1} => {c: 1, aa: {{a: 10}, {a: 11}}}
		return a.Update(key)
	})
	log.Print("mapped = ", mapped)

	// {
	//   {c: 1, aa: {{a: 10}, {a: 11}}},
	//   {c: 2, aa: {{a: 13}}},
	//   {c: 3, aa: {{a: 10}, {a: 11}}},
	//   {c: 4, aa: {{a: 13}, {a: 14}}},
	// }
	result := mapped.Values()
	log.Print("result = ", result)
	return result
}

// Unnest returns a relation with some subrelations unnested. This is the
// reverse of Nest.
func Unnest(s Relation, attrs frozen.Set[string]) Relation {
	var b RelationBuilder
	for i := s.Range(); i.Next(); {
		t := i.Value()
		key := t.Without(attrs)
		nestedValues := frozen.MapMap(t.Project(attrs), func(_ string, val any) Relation {
			return val.(Relation)
		}).Values()
		all := nestedValues.With(frozen.NewSet(key))
		joined := Join(all.Elements()...)
		for j := joined.Range(); j.Next(); {
			b.Add(j.Value())
		}
	}
	return b.Finish()
}
