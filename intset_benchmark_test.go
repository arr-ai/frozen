package frozen_test

import (
	"strconv"
	"testing"

	. "github.com/arr-ai/frozen"
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

func BenchmarkNewIntSetN(b *testing.B) {
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
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			benchmarkNewIntSet(b, n)
		})
	}

	for _, n := range sizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			benchmarkWithIntSet(b, n)
		})
	}
}

func BenchmarkNewIntSet100(b *testing.B) {
	for _, e := range []struct {
		name string
		n    int
	}{{"100", 100}, {"100k", 100_000}, {"1M", 1_000_000}} {
		e := e
		b.Run(e.name, func(b *testing.B) {
			benchmarkNewIntSet(b, e.n)
		})
		b.Run(e.name, func(b *testing.B) {
			benchmarkWithIntSet(b, e.n)
		})
	}
}
