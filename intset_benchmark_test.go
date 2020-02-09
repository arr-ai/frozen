package frozen

import (
	"testing"
)

func benchmarkNewIntSet(b *testing.B, n int) {
	arr, _ := generateIntArrayAndSet(n)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewIntSet(arr...)
	}
}

func benchmarkWithIntSet(b *testing.B, n int) {
	_, set := generateIntArrayAndSet(n)
	multiplier := 2147483647 % n
	withouts := make([]int, 0, b.N)
	for i := 0; i < b.N; i++ {
		num := i * multiplier
		withouts = append(withouts, num)
	}
	set = set.Without(withouts...)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.With(i * multiplier)
	}
}

func BenchmarkNewIntSet100(b *testing.B)  { benchmarkNewIntSet(b, 100) }
func BenchmarkNewIntSet1k(b *testing.B)   { benchmarkNewIntSet(b, 1000) }
func BenchmarkNewIntSet100k(b *testing.B) { benchmarkNewIntSet(b, 100000) }
func BenchmarkNewIntSet1M(b *testing.B)   { benchmarkNewIntSet(b, 1000000) }

func BenchmarkWithIntSet100(b *testing.B)  { benchmarkWithIntSet(b, 100) }
func BenchmarkWithIntSet1k(b *testing.B)   { benchmarkWithIntSet(b, 1000) }
func BenchmarkWithIntSet100k(b *testing.B) { benchmarkWithIntSet(b, 100000) }
func BenchmarkWithIntSet1M(b *testing.B)   { benchmarkWithIntSet(b, 1000000) }
