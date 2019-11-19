package frozen

import (
	"testing"

	"github.com/mediocregopher/seq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHamtEmpty(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	assert.Zero(t, h.count())
}

func TestHamtSmall(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	assert.Zero(t, h.count())
	assert.True(t, h.isEmpty())
	h = h.put("foo", 42)
	assert.Equal(t, 1, h.count())
	assert.False(t, h.isEmpty())
	h = h.put("bar", 43)
	assert.Equal(t, 2, h.count())
	assert.False(t, h.isEmpty())
	h = h.put("foo", 44)
	assert.Equal(t, 2, h.count())
	assert.False(t, h.isEmpty())
}

func TestHamtLarge(t *testing.T) {
	t.Parallel()

	hh := []hamt{}
	var h hamt = empty{}
	for i := 0; i < 1000; i++ {
		hh = append(hh, h)
		h = h.put(i, 42)
	}
	for i, h := range hh {
		assert.Equal(t, i, h.count())
		assert.Equal(t, h.count() == 0, h.isEmpty())
	}
}

func TestHamtGet(t *testing.T) {
	t.Parallel()

	hh := []hamt{}
	var h hamt = empty{}
	for i := 0; i < 100; i++ {
		hh = append(hh, h)
		var v interface{}
		var has bool
		if assert.NotPanics(t, func() {
			v, has = h.get(i)
		}, "i=%v", i) {
			assert.False(t, has, "i=%v v=%v", i, v)
		} else {
			h.get(i)
		}
		hOld := h
		h = h.put(i, i*i)
		if v2, has := h.get(i); assert.True(t, has, "i=%v", i) {
			if !assert.Equal(t, i*i, v2, "i=%v", i) {
				hOld.put(i, i*i)
				h.get(i)
			}
		}
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			if v, has := h.get(j); assert.True(t, has, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, v, "i=%v j=%v", i, j)
			}
		}
	}
}

func TestHamtDelete(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	const N = 1000
	for i := 0; i < N; i++ {
		h = h.put(i, i*i)
	}

	d := h
	for i := 0; i < N; i++ {
		assert.Equal(t, N-i, d.count())
		assert.False(t, h.isEmpty())
		_, has := d.get(i)
		if assert.True(t, has, "i=%v", i) {
			d = d.delete(i)
			v, has := d.get(i)
			assert.False(t, has, "i=%v v=%v", i, v)
		} else {
			d.get(i)
		}
	}
	assert.Zero(t, d.count())
	assert.True(t, d.isEmpty())

	d = h
	for i := N; i > 0; {
		assert.Equal(t, i, d.count())
		i--
		assert.False(t, h.isEmpty())
		_, has := d.get(i)
		if assert.True(t, has, "i=%v", i) {
			d = d.delete(i)
			v, has := d.get(i)
			assert.False(t, has, "i=%v, v=%v", i, v)
		}
	}
	assert.Zero(t, d.count())
	assert.True(t, d.isEmpty())
}

func TestHamtDeleteMissing(t *testing.T) {
	t.Parallel()

	h := empty{}.put("foo", 42)
	h = h.delete("bar")
	assert.Equal(t, 1, h.count())
	h = h.delete("foo")
	assert.Equal(t, 0, h.count())
}

func TestHamtIter(t *testing.T) {
	t.Parallel()

	a := make([]int, 18)
	var h hamt = empty{}
	for i := range a {
		h = h.put(i, i*i)
		a[i] = -1
	}
	n := 0
	for it := h.iterator(); it.next(); n++ {
		i := it.e.key.(int)
		v := it.e.value.(int)
		a[i] = v
		assert.Equal(t, i*i, v, "it=%v, h=%v", it, h)
	}
	require.Equal(t, len(a), n, "h=%v a=%v", h, a)
	for i, v := range a {
		assert.Equal(t, i*i, v, "i=%v", i)
	}
}

func benchmarkInsertMapInt(b *testing.B, N int) {
	m := map[int]int{}
	for i := 0; i < N; i++ {
		m[i] = i * i
	}
	b.ResetTimer()
	for i := N; i < N+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInt(b *testing.B) {
	benchmarkInsertMapInt(b, 0)
}

func BenchmarkInsertMapInt1M(b *testing.B) {
	benchmarkInsertMapInt(b, 1_000_000)
}

func benchmarkInsertMapInterface(b *testing.B, N int) {
	m := map[interface{}]interface{}{}
	for i := 0; i < N; i++ {
		m[i] = i * i
	}
	b.ResetTimer()
	for i := N; i < N+b.N; i++ {
		m[i] = i * i
	}
}

func BenchmarkInsertMapInterface(b *testing.B) {
	benchmarkInsertMapInterface(b, 0)
}

func BenchmarkInsertMapInterface1M(b *testing.B) {
	benchmarkInsertMapInterface(b, 1_000_000)
}

var hamtPrepop = map[int]hamt{
	0: empty{},
	1_000_000: func() hamt {
		var h hamt = empty{}
		for i := 0; i < 1_000_000; i++ {
			h = h.put(i, i*i)
		}
		return h
	}(),
}

func benchmarkInsertFrozen(b *testing.B, N int) {
	h := hamtPrepop[N]
	b.ResetTimer()
	for i := N; i < N+b.N; i++ {
		h = h.put(i, i*i)
	}
}

func BenchmarkInsertFrozen(b *testing.B) {
	benchmarkInsertFrozen(b, 0)
}

func BenchmarkInsertFrozen1M(b *testing.B) {
	benchmarkInsertFrozen(b, 1_000_000)
}

var mediocrePrepop = map[int]*seq.HashMap{
	0: seq.NewHashMap(),
	10_000: func() *seq.HashMap {
		s := seq.NewHashMap()
		for i := 0; i < 10_000; i++ {
			s, _ = s.Set(i, i*i)
		}
		return s
	}(),
}

func BenchmarkInsertMediocre(b *testing.B) {
	benchmarkInsertMediocre(b, 0)
}

func BenchmarkInsertMediocre10k(b *testing.B) {
	benchmarkInsertMediocre(b, 10_000)
}

func benchmarkInsertMediocre(b *testing.B, N int) {
	s := mediocrePrepop[N]
	b.ResetTimer()
	for i := N; i < N+b.N; i++ {
		s, _ = s.Set(i, i*i)
	}
}
