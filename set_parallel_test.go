package frozen

import (
	"runtime"
	"sync"
	"testing"
)

func benchmarkSequential(b *testing.B, name string, size int) { //nolint:gocognit,funlen
	b.Run(name, func(b *testing.B) {
		b.Run("SetInt/Prealloc", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				m := make(map[int]struct{}, size)
				for i := 0; i < size; i++ {
					m[i] = struct{}{}
				}
			}
		})

		b.Run("SetInt", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				m := map[int]struct{}{}
				for i := 0; i < size; i++ {
					m[i] = struct{}{}
				}
			}
		})

		b.Run("SetInterface/Prealloc", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				m := make(map[interface{}]struct{}, size)
				for i := 0; i < size; i++ {
					m[i] = struct{}{}
				}
			}
		})

		b.Run("SetInterface", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				m := map[interface{}]struct{}{}
				for i := 0; i < size; i++ {
					m[i] = struct{}{}
				}
			}
		})

		b.Run("Frozen/SetBuilder/Add/Prealloc", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				sb := NewSetBuilder(size)
				for i := 0; i < size; i++ {
					sb.Add(i)
				}
				_ = sb.Finish()
			}
		})

		b.Run("Frozen/SetBuilder/Add", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var sb SetBuilder
				for i := 0; i < size; i++ {
					sb.Add(i)
				}
				_ = sb.Finish()
			}
		})

		b.Run("Frozen/Set/With", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var s Set
				for i := 0; i < size; i++ {
					s = s.With(i)
				}
			}
		})
	})
}

func BenchmarkSequential(b *testing.B) {
	benchmarkSequential(b, "32", 32)
	benchmarkSequential(b, "1ki", 1<<10)
	benchmarkSequential(b, "1Mi", 1<<20)
}

func parallelUnion(sets []Set) Set {
	switch len(sets) {
	case 1:
		return sets[0]
	case 2:
		return sets[0].Union(sets[1])
	default:
		half := len(sets) / 2
		ch := make(chan Set)
		go func() {
			ch <- parallelUnion(sets[:half])
		}()
		return parallelUnion(sets[half:]).Union(<-ch)
	}
}

func BenchmarkSetParallelWith1M(b *testing.B) {
	D := runtime.GOMAXPROCS(0)
	for n := 0; n < b.N; n++ {
		sets := make([]Set, D)
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

func BenchmarkSetParallelBuilder1M(b *testing.B) {
	D := runtime.GOMAXPROCS(0)
	for n := 0; n < b.N; n++ {
		builders := make([]SetBuilder, D)
		var wg sync.WaitGroup
		wg.Add(D)
		for d := 0; d < D; d++ {
			d := d
			go func() {
				sb := &builders[d]
				for i := d; i < 1<<20; i += D {
					sb.Add(i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		sets := make([]Set, 0, D)
		for _, builder := range builders {
			sets = append(sets, builder.Finish())
		}
		s := parallelUnion(sets)
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}
