package frozen_test

import (
	"runtime"
	"sync"
	"testing"

	. "github.com/arr-ai/frozen"
)

func benchmarkSequential(b *testing.B, name string, size int) {
	b.Helper()

	b.Run(name, func(b *testing.B) {
		b.Run("SetInt-Prealloc", func(t *testing.B) { benchmarkSetIntPrealloc(t, size) })
		b.Run("SetInt", func(t *testing.B) { benchmarkSetInt(t, size) })
		b.Run("SetInterface-Prealloc", func(t *testing.B) { benchmarkSetInterfacePrealloc(t, size) })
		b.Run("SetInterface", func(t *testing.B) { benchmarkSetInterface(t, size) })
		b.Run("Frozen-SetBuilder-Add-Prealloc", func(t *testing.B) { benchmarkFrozenSetBuilderAddPrealloc(t, size) })
		b.Run("Frozen-SetBuilder-Add", func(t *testing.B) { benchmarkFrozenSetBuilderAdd(t, size) })
		b.Run("Frozen-Set-With", func(t *testing.B) { benchmarkFrozenSetWith(t, size) })
	})
}

func benchmarkSetIntPrealloc(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		m := make(map[int]struct{}, size)
		for i := 0; i < size; i++ {
			m[i] = struct{}{}
		}
	}
}

func benchmarkSetInt(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		m := map[int]struct{}{}
		for i := 0; i < size; i++ {
			m[i] = struct{}{}
		}
	}
}

func benchmarkSetInterfacePrealloc(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		m := make(map[any]struct{}, size)
		for i := 0; i < size; i++ {
			m[i] = struct{}{}
		}
	}
}

func benchmarkSetInterface(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		m := map[any]struct{}{}
		for i := 0; i < size; i++ {
			m[i] = struct{}{}
		}
	}
}

func benchmarkFrozenSetBuilderAddPrealloc(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		sb := NewSetBuilder[int](size)
		for i := 0; i < size; i++ {
			sb.Add(i)
		}
		_ = sb.Finish()
	}
}

func benchmarkFrozenSetBuilderAdd(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		var sb SetBuilder[int]
		for i := 0; i < size; i++ {
			sb.Add(i)
		}
		_ = sb.Finish()
	}
}

func benchmarkFrozenSetWith(b *testing.B, size int) {
	b.Helper()

	for n := 0; n < b.N; n++ {
		var s Set[int]
		for i := 0; i < size; i++ {
			s = s.With(i)
		}
	}
}

func BenchmarkSequential(b *testing.B) {
	benchmarkSequential(b, "32", 32)
	benchmarkSequential(b, "1ki", 1<<10)
	benchmarkSequential(b, "1Mi", 1<<20)
}

func parallelUnion(sets []Set[int]) Set[int] {
	switch len(sets) {
	case 1:
		return sets[0]
	case 2:
		return sets[0].Union(sets[1])
	default:
		half := len(sets) / 2
		ch := make(chan Set[int])
		go func() {
			ch <- parallelUnion(sets[:half])
		}()
		return parallelUnion(sets[half:]).Union(<-ch)
	}
}

func BenchmarkSetParallelWith1M(b *testing.B) {
	D := runtime.GOMAXPROCS(0)
	for n := 0; n < b.N; n++ {
		sets := make([]Set[int], D)
		var wg sync.WaitGroup
		wg.Add(D)
		for d := 0; d < D; d++ {
			d := d
			go func() {
				s := &sets[d]
				for i := d; i < 1<<20; i += D {
					*s = s.With(i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		s := parallelUnion(sets)
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}

func BenchmarkSetUnion1M(b *testing.B) {
	s5 := SetMap(Iota(1<<19), func(i int) int { return 5 * i })
	s7 := SetMap(Iota(1<<20), func(i int) int { return 7 * i })
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s5.Union(s7)
	}
}
