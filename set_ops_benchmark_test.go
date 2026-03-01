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
//	Near:     a=[0,n), b=[0,n) with 1% replaced — 99% overlap.
//	Half:     a=[0,n), b=[n/2, n+n/2) — 50% overlap.
//	Sparse:   a=[0,n), b=[n,2n) with 1% from a — 1% overlap.
//	Disjoint: a=[0,n), b=[n,2n) — 0% overlap.
// onePercent returns max(n/100, 1).
func onePercent(n int) int {
	if d := n / 100; d > 0 {
		return d
	}
	return 1
}

func buildSetPair(n int, overlap string) (a, b frozen.Set[int]) {
	a = frozen.Iota(n)
	switch overlap {
	case "Same":
		b = a
	case "Equal":
		b = frozen.Iota(n)
	case "Near":
		// 99% identical: start with same elements, replace 1%.
		b = a
		for i, delta := 0, onePercent(n); i < delta; i++ {
			b = b.Without(i).With(n + i)
		}
	case "Half":
		b = frozen.Iota2(n/2, n+n/2)
	case "Sparse":
		// 1% overlap: mostly disjoint, with 1% of a's elements.
		b = frozen.Iota2(n, 2*n)
		for i, delta := 0, onePercent(n); i < delta; i++ {
			b = b.Without(n + i).With(i)
		}
	case "Disjoint":
		b = frozen.Iota2(n, 2*n)
	default:
		panic("unknown overlap: " + overlap)
	}
	return a, b
}

var overlaps = []string{"Same", "Equal", "Near", "Half", "Sparse", "Disjoint"}

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
