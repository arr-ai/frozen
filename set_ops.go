package frozen

// Intersection returns the n-ary intersection of a Set of Sets.
func Intersection(sets Set) Set {
	if sets.IsEmpty() {
		panic("must have at least one set to intersect")
	}
	r := sets.Range()
	r.Next()
	result := r.Value().(Set)
	for r.Next() {
		result = result.Join(r.Value().(Set))
	}
	return result
}

// Union returns the n-ary union of a Set of Sets.
func Union(relations Set) Set {
	var b SetBuilder
	for r := relations.Range(); r.Next(); {
		for t := r.Value().(Set).Range(); t.Next(); {
			b.Add(t.Value())
		}
	}
	return b.Finish()
}
