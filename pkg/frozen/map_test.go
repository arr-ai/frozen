package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyMap(t *testing.T) {
	var m Map
	assert.True(t, m.IsEmpty())
	m = m.With(1, 2)
	assert.False(t, m.IsEmpty())
	m = m.Without(NewSet(1))
	assert.True(t, m.IsEmpty())
}

var prepopMapInt = memoizePrepop(func(n int) interface{} {
	m := map[int]int{}
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInt(b *testing.B, n int) {
	m := prepopMapInt(n).(map[int]int)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInt0(b *testing.B) {
	benchmarkInsertMapInt(b, 0)
}

func BenchmarkInsertMapInt1k(b *testing.B) {
	benchmarkInsertMapInt(b, 1<<10)
}

func BenchmarkInsertMapInt1M(b *testing.B) {
	benchmarkInsertMapInt(b, 1<<20)
}

var prepopMapInterface = memoizePrepop(func(n int) interface{} {
	m := map[interface{}]interface{}{}
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
	return m
})

func benchmarkInsertMapInterface(b *testing.B, n int) {
	m := prepopMapInterface(n).(map[interface{}]interface{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInterface0(b *testing.B) {
	benchmarkInsertMapInterface(b, 0)
}

func BenchmarkInsertMapInterface1k(b *testing.B) {
	benchmarkInsertMapInterface(b, 1<<10)
}

func BenchmarkInsertMapInterface1M(b *testing.B) {
	benchmarkInsertMapInterface(b, 1<<20)
}

var prepopFrozenMap = memoizePrepop(func(n int) interface{} {
	var m Map
	for i := 0; i < n; i++ {
		m = m.With(i, i*i)
	}
	return m
})

func benchmarkInsertFrozenMap(b *testing.B, n int) {
	m := prepopFrozenMap(n).(Map)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m.With(i, i*i)
	}
}

func BenchmarkInsertFrozenMap0(b *testing.B) {
	benchmarkInsertFrozenMap(b, 0)
}

func BenchmarkInsertFrozenMap1k(b *testing.B) {
	benchmarkInsertFrozenMap(b, 1<<10)
}

func BenchmarkInsertFrozenMap1M(b *testing.B) {
	benchmarkInsertFrozenMap(b, 1<<20)
}
