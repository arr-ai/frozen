package frozen_test

import (
	"fmt"
	"testing"

	. "github.com/arr-ai/frozen/v2"
)

func benchmarkNewIntSet(b *testing.B, n int) {
	b.Helper()

	arr, _ := generateIntArrayAndSet(n)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewIntSet(arr...)
	}
}

func benchmarkWithIntSet(b *testing.B, n int) {
	b.Helper()

	_, set := generateIntArrayAndSet(n)
	multiplier := 2147483647 % n
	withouts := make([]int, 0, b.N)
	for i := 0; i < b.N; i++ {
		withouts = append(withouts, i*multiplier)
	}
	set = set.Without(withouts...)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.With(i * multiplier)
	}
}

func BenchmarkIntSetN(b *testing.B) {
	// Uncomment for occasional use
	b.Skip()

	sizes := []int{
		100,
		1_000,
		10_000,
		80_000,
		100_000,
		200_000,
		300_000,
		500_000,
		1_000_000,
		2_000_000,
	}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("New/%d", n), func(b *testing.B) {
			benchmarkNewIntSet(b, n)
		})
	}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("With/%d", n), func(b *testing.B) {
			benchmarkWithIntSet(b, n)
		})
	}
}

func BenchmarkIntSet(b *testing.B) {
	for _, e := range []struct {
		name string
		n    int
	}{{"100", 100}, {"1k", 1_000}, {"100k", 100_000}, {"1M", 1_000_000}} {
		e := e
		b.Run("New/"+e.name, func(b *testing.B) {
			benchmarkNewIntSet(b, e.n)
		})
		b.Run("With/"+e.name, func(b *testing.B) {
			benchmarkWithIntSet(b, e.n)
		})
	}
}
