package frozen

// Intersection returns the n-ary intersection of a Set of Sets.
func Intersection[T any](sets ...Set[T]) Set[T] {
	if len(sets) == 0 {
		panic("must have at least one set to intersect")
	}
	result := sets[0]
	for _, s := range sets {
		result = result.Intersection(s)
	}
	return result
}

// Union returns the n-ary union of a Set of Sets.
func Union[T any](sets ...Set[T]) Set[T] {
	var result Set[T]
	for _, s := range sets {
		result = result.Union(s)
	}
	return result
}
