package frozen

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
