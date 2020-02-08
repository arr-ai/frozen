package frozen

import (
	"testing"
)

func benchmarkNewIntSet(b *testing.B, n int) {
	arr, _ := generateIntArrayAndSet()
	max := n
	if n > len(arr) {
			b.Logf("n of %d is bigger than the generated array length of %d", n, len(arr))
			max = len(arr)
	}
	arr = arr[:max]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewIntSet(arr...)
	}
}

func benchmarkWithIntSet(b *testing.B, n int) {
	arr, _ := generateIntArrayAndSet()
	max := n
	if n > len(arr) {
			b.Logf("n of %d is bigger than the generated array length of %d", n, len(arr))
			max = len(arr)
	}
	arr = arr[:max]
	set := NewIntSet()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.With(arr...)
	}
}

func BenchmarkNewIntSet100(b *testing.B) {benchmarkNewIntSet(b, 100)}
func BenchmarkNewIntSet1k(b *testing.B) {benchmarkNewIntSet(b, 1000)}
func BenchmarkNewIntSet100k(b *testing.B) {benchmarkNewIntSet(b, 100000)}
func BenchmarkNewIntSet1M(b *testing.B) {benchmarkNewIntSet(b, 1000000)}

func BenchmarkWithIntSet100(b *testing.B) {benchmarkWithIntSet(b, 100)}
func BenchmarkWithIntSet1k(b *testing.B) {benchmarkWithIntSet(b, 1000)}
func BenchmarkWithIntSet100k(b *testing.B) {benchmarkWithIntSet(b, 100000)}
func BenchmarkWithIntSet1M(b *testing.B) {benchmarkWithIntSet(b, 1000000)}