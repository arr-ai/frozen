package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHamtEmpty(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	assert.True(t, h.isEmpty())
}

func TestHamtSmall(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	assert.True(t, h.isEmpty())
	h, _ = h.put(KV("foo", 42), newBuffer(0))
	assert.False(t, h.isEmpty())
	h, _ = h.put(KV("bar", 43), newBuffer(1))
	assert.False(t, h.isEmpty())
	h, _ = h.put(KV("foo", 44), newBuffer(2))
	assert.False(t, h.isEmpty())
}

func TestHamtLarge(t *testing.T) {
	t.Parallel()

	hh := []hamt{}
	var h hamt = empty{}
	for i := 0; i < 500; i++ {
		hh = append(hh, h)
		h, _ = h.put(KV(i, i*i), newBuffer(i))
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			v, has := h.get(KV(j, nil))
			if assert.True(t, has, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, v.(KeyValue).Value, "i=%v j=%v", i, j)
			}
		}
		kv, has := h.get(KV(i, nil))
		assert.False(t, has, "i=%v v=%v", i, kv)
	}
}

func TestHamtGet(t *testing.T) {
	t.Parallel()

	hh := []hamt{}
	var h hamt = empty{}
	for i := 0; i < 500; i++ {
		hh = append(hh, h)
		var kv interface{}
		var has bool
		if assert.NotPanics(t, func() {
			kv, has = h.get(i)
		}, "i=%v", i) {
			assert.False(t, has, "i=%v v=%v", i, kv)
		} else {
			h.get(i)
		}
		hOld := h
		h, _ = h.put(KV(i, i*i), newBuffer(i))
		if kv, has := h.get(KV(i, nil)); assert.True(t, has, "i=%v", i) {
			if !assert.Equal(t, i*i, kv.(KeyValue).Value, "i=%v", i) {
				hOld.put(KV(i, i*i), newBuffer(i))
				h.get(KV(i, nil))
			}
		}
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			if kv, has := h.get(KV(j, nil)); assert.True(t, has, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, kv.(KeyValue).Value, "i=%v j=%v kv=%v", i, j, kv)
			}
		}
	}
}

func TestHamtDelete(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	const N = 1000
	for i := 0; i < N; i++ {
		h, _ = h.put(i, newBuffer(i))
	}

	d := h
	for i := 0; i < N; i++ {
		assert.False(t, h.isEmpty())
		_, has := d.get(i)
		if assert.True(t, has, "i=%v", i) {
			d, _ = d.delete(i, newBuffer(N-i))
			v, has := d.get(i)
			assert.False(t, has, "i=%v v=%v", i, v)
		}
	}
	assert.True(t, d.isEmpty())

	d = h
	for i := N; i > 0; {
		i--
		assert.False(t, h.isEmpty())
		_, has := d.get(i)
		if assert.True(t, has, "i=%v", i) {
			d, _ = d.delete(i, newBuffer(i-1))
			v, has := d.get(i)
			assert.False(t, has, "i=%v, v=%v", i, v)
		}
	}
	assert.True(t, d.isEmpty())
}

func TestHamtDeleteMissing(t *testing.T) {
	t.Parallel()

	h, _ := empty{}.put("foo", newBuffer(0))
	h, _ = h.delete("bar", newBuffer(0))
	assert.False(t, h.isEmpty())
	h, _ = h.delete("foo", newBuffer(0))
	assert.True(t, h.isEmpty())
}

func TestHamtIter(t *testing.T) {
	t.Parallel()

	var h hamt = empty{}
	for i := 0; i < 64; i++ {
		h, _ = h.put(i, newBuffer(i))
	}

	var a uint64 = 0
	n := 0
	for it := h.iterator(); it.next(); n++ {
		i := it.e.elem.(int)
		a |= uint64(1) << i
	}
	assert.Equal(t, 64, n, "h=%v a=%b", h, a)
	assert.Zero(t, ^a)
}

var hamtPrepop = func() map[int]hamt {
	prepop := map[int]hamt{}
	for _, n := range []int{0, 1 << 10, 1 << 20} {
		var h hamt = empty{}
		for i := 0; i < n; i++ {
			h, _ = h.put(KV(i, i*i), newBuffer(i))
		}
		prepop[n] = h
	}
	return prepop
}()

func benchmarkInsertFrozenHamt(b *testing.B, n int) {
	h := hamtPrepop[n]
	for i := n; i < n+b.N; i++ {
		h.put(KV(i, i*i), newBuffer(i))
	}
}

func BenchmarkInsertFrozenHamt0(b *testing.B) {
	benchmarkInsertFrozenHamt(b, 0)
}

func BenchmarkInsertFrozenHamt1k(b *testing.B) {
	benchmarkInsertFrozenHamt(b, 1<<10)
}

func BenchmarkInsertFrozenHamt1M(b *testing.B) {
	benchmarkInsertFrozenHamt(b, 1<<20)
}
