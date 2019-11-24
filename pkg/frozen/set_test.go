package frozen

import (
	"testing"

	"github.com/mediocregopher/seq"
)

func benchmarkInsertSetInt(b *testing.B, n int) {
	m := map[int]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInt0(b *testing.B) {
	benchmarkInsertSetInt(b, 0)
}

func BenchmarkInsertSetInt1k(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<10)
}

func BenchmarkInsertSetInt1M(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<20)
}

func benchmarkInsertSetInterface(b *testing.B, n int) {
	m := map[interface{}]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInterface0(b *testing.B) {
	benchmarkInsertSetInterface(b, 0)
}

func BenchmarkInsertSetInterface1k(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<10)
}

func BenchmarkInsertSetInterface1M(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<20)
}

var frozenSetPrepop = func() map[int]Set {
	prepop := map[int]Set{}
	for _, n := range []int{0, 1 << 10, 1 << 20} {
		var s Set
		for i := 0; i < n; i++ {
			s = s.With(i)
		}
		prepop[n] = s
	}
	return prepop
}()

func benchmarkInsertFrozenSet(b *testing.B, n int) {
	s := frozenSetPrepop[n]
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		s.With(i)
	}
}

func BenchmarkInsertFrozenSet0(b *testing.B) {
	benchmarkInsertFrozenSet(b, 0)
}

func BenchmarkInsertFrozenSet1k(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<10)
}

func BenchmarkInsertFrozenSet1M(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<20)
}

var mediocreSetPrepop = func() map[int]*seq.Set {
	prepop := map[int]*seq.Set{}
	for _, n := range []int{0, 10, 10 << 10} {
		s := seq.NewSet()
		for i := 0; i < n; i++ {
			s, _ = s.SetVal(i)
		}
		prepop[n] = s
	}
	return prepop
}()

func benchmarkInsertMediocreSet(b *testing.B, n int) {
	s := mediocreSetPrepop[n]
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		s.SetVal(i)
	}
}

func BenchmarkInsertMediocreSet0(b *testing.B) {
	benchmarkInsertMediocreSet(b, 0)
}

func BenchmarkInsertMediocreSet10(b *testing.B) {
	benchmarkInsertMediocreSet(b, 10)
}

func BenchmarkInsertMediocreSet10k(b *testing.B) {
	benchmarkInsertMediocreSet(b, 10<<10)
}
