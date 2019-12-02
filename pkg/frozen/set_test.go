package frozen

import (
	"testing"
)

var prepopSetInt = memoizePrepop(func(n int) interface{} {
	m := map[int]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInt(b *testing.B, n int) {
	m := prepopSetInt(n).(map[int]struct{})
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

var prepopSetInterface = memoizePrepop(func(n int) interface{} {
	m := map[interface{}]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInterface(b *testing.B, n int) {
	m := prepopSetInterface(n).(map[interface{}]struct{})
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

var prepopFrozenSet = memoizePrepop(func(n int) interface{} {
	var s Set
	for i := 0; i < n; i++ {
		s = s.With(i)
	}
	return s
})

func benchmarkInsertFrozenSet(b *testing.B, n int) {
	s := prepopFrozenSet(n).(Set)
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
