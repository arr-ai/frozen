package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
)

type benchSize struct {
	name string
	n    int
}

var benchSizes = []benchSize{
	{"1ki", 1 << 10},
	{"1Mi", 1 << 20},
}

// buildSetPair constructs two sets with a specified overlap pattern.
//
//	Same:     a and b are the same Go value (pointer-identical tree).
//	Equal:    a and b independently built with identical elements.
//	Half:     a=[0,n), b=[n/2, n+n/2) — 50% overlap.
//	Disjoint: a=[0,n), b=[n,2n) — 0% overlap.
func buildSetPair(n int, overlap string) (a, b frozen.Set[int]) {
	a = frozen.Iota(n)
	switch overlap {
	case "Same":
		b = a
	case "Equal":
		b = frozen.Iota(n)
	case "Half":
		b = frozen.Iota2(n/2, n+n/2)
	case "Disjoint":
		b = frozen.Iota2(n, 2*n)
	default:
		panic("unknown overlap: " + overlap)
	}
	return
}

var overlaps = []string{"Same", "Equal", "Half", "Disjoint"}

func benchSetOp(b *testing.B, name string, op func(a, other frozen.Set[int])) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		for _, sz := range benchSizes {
			for _, ov := range overlaps {
				sz, ov := sz, ov
				b.Run(sz.name+"/"+ov, func(b *testing.B) {
					a, other := buildSetPair(sz.n, ov)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						op(a, other)
					}
				})
			}
		}
	})
}

func BenchmarkSetOps(b *testing.B) {
	benchSetOp(b, "Equal", func(a, o frozen.Set[int]) { a.Equal(o) })
	benchSetOp(b, "Intersection", func(a, o frozen.Set[int]) { a.Intersection(o) })
	benchSetOp(b, "Difference", func(a, o frozen.Set[int]) { a.Difference(o) })
	benchSetOp(b, "SubsetOf", func(a, o frozen.Set[int]) { a.IsSubsetOf(o) })

	b.Run("Has", func(b *testing.B) {
		for _, sz := range benchSizes {
			sz := sz
			b.Run(sz.name+"/Hit", func(b *testing.B) {
				s := frozen.Iota(sz.n)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					s.Has(i % sz.n)
				}
			})
			b.Run(sz.name+"/Miss", func(b *testing.B) {
				s := frozen.Iota(sz.n)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					s.Has(sz.n + i)
				}
			})
		}
	})
}
