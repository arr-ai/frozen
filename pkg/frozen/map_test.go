package frozen

import (
	"testing"

	"github.com/mediocregopher/seq"
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

func benchmarkInsertMapInt(b *testing.B, n int) {
	m := map[int]int{}
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
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

func benchmarkInsertMapInterface(b *testing.B, n int) {
	m := map[interface{}]interface{}{}
	for i := 0; i < n; i++ {
		m[i] = i * i
	}
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

var frozenMapPrepop = func() map[int]Map {
	prepop := map[int]Map{}
	for _, n := range []int{0, 1 << 10, 1 << 20} {
		var m Map
		for i := 0; i < n; i++ {
			m = m.With(i, i*i)
		}
		prepop[n] = m
	}
	return prepop
}()

func benchmarkInsertFrozenMap(b *testing.B, n int) {
	m := frozenMapPrepop[n]
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

var mediocreHashMapPrepop = func() map[int]*seq.HashMap {
	prepop := map[int]*seq.HashMap{}
	for _, n := range []int{0, 10, 10 << 10} {
		m := seq.NewHashMap()
		for i := 0; i < n; i++ {
			m, _ = m.Set(i, i*i)
		}
		prepop[n] = m
	}
	return prepop
}()

func benchmarkInsertMediocreHashMap(b *testing.B, n int) {
	m := mediocreHashMapPrepop[n]
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m.Set(i, i*i)
	}
}

func BenchmarkInsertMediocreHashMap0(b *testing.B) {
	benchmarkInsertMediocreHashMap(b, 0)
}

func BenchmarkInsertMediocreHashMap10(b *testing.B) {
	benchmarkInsertMediocreHashMap(b, 10)
}

func BenchmarkInsertMediocreHashMap10k(b *testing.B) {
	benchmarkInsertMediocreHashMap(b, 10<<10)
}
