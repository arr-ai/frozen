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
func Union(sets ...Set) Set {
	var result Set
	for _, s := range sets {
		result = result.Union(s)
	}
	return result
}
