package lazy

func symmetricDifference(a, b Set) Set {
	return a.Difference(b).Union(b.Difference(a))
}
