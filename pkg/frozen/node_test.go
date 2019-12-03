package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeSmall(t *testing.T) {
	t.Parallel()

	var h *node
	h, _ = h.put(KV("foo", 42))
	assert.NotNil(t, h)
	h, _ = h.put(KV("bar", 43))
	assert.NotNil(t, h)
	h, _ = h.put(KV("foo", 44))
	assert.NotNil(t, h)
}

func TestNodeLarge(t *testing.T) {
	t.Parallel()

	hh := []*node{}
	var h *node
	for i := 0; i < 500; i++ {
		hh = append(hh, h)
		for j := 0; j < i; j++ {
			kv := h.get(KV(j, nil))
			if assert.NotNil(t, kv, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, kv.(KeyValue).Value, "i=%v j=%v", i, j)
			}
		}
		kv := h.get(KV(i, nil))
		if !assert.Nil(t, kv, "i=%v v=%v", i, kv) {
			h.get(KV(i, nil))
		}
		h, _ = h.put(KV(i, i*i))
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			kv := h.get(KV(j, nil))
			if assert.NotNil(t, kv, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, kv.(KeyValue).Value, "i=%v j=%v", i, j)
			} else {
				h.get(KV(j, nil))
			}
		}
		kv := h.get(KV(i, nil))
		assert.Nil(t, kv, "i=%v v=%v", i, kv)
	}
}

func TestNodeGet(t *testing.T) {
	t.Parallel()

	hh := []*node{}
	var h *node
	for i := 0; i < 500; i++ {
		i := i
		hh = append(hh, h)
		var kv interface{}
		if assert.NotPanics(t, func() {
			kv = h.get(i)
		}, "i=%v", i) {
			assert.Nil(t, kv, "i=%v v=%v", i, kv)
		} else {
			h.get(i)
		}
		hOld := h
		h, _ = h.put(KV(i, i*i))
		if kv := h.get(KV(i, nil)); assert.NotNil(t, kv, "i=%v", i) {
			if !assert.Equal(t, i*i, kv.(KeyValue).Value, "i=%v", i) {
				hOld.put(KV(i, i*i))
				h.get(KV(i, nil))
			}
		}
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			if kv := h.get(KV(j, nil)); assert.NotNil(t, kv, "i=%v j=%v", i, j) {
				assert.Equal(t, j*j, kv.(KeyValue).Value, "i=%v j=%v kv=%v", i, j, kv)
			}
		}
	}
}

func TestNodeDelete(t *testing.T) {
	t.Parallel()

	var h *node
	const N = 1000
	for i := 0; i < N; i++ {
		h, _ = h.put(i)
	}

	d := h
	for i := 0; i < N; i++ {
		assert.NotNil(t, h)
		require.NotNil(t, d.get(i), "i=%v", i)
		d, _ = d.delete(i)
		assert.Nil(t, d.get(i), "i=%v", i)
	}
	assert.Nil(t, d)

	d = h
	for i := N; i > 0; {
		i--
		assert.NotNil(t, h)
		v := d.get(i)
		if assert.NotNil(t, v, "i=%v", i) {
			d, _ = d.delete(i)
			v := d.get(i)
			assert.Nil(t, v, "i=%v", i)
		}
	}
	assert.Nil(t, d)
}

func TestNodeDeleteMissing(t *testing.T) {
	t.Parallel()

	h, _ := ((*node)(nil)).put("foo")
	h, _ = h.delete("bar")
	assert.NotNil(t, h)
	h, _ = h.delete("foo")
	assert.True(t, h == nil)
}

func TestNodeIter(t *testing.T) {
	t.Parallel()

	var h *node
	for i := 0; i < 64; i++ {
		h, _ = h.put(i)
	}

	var a uint64 = 0
	n := 0
	for it := h.iterator(); it.next(); n++ {
		i, ok := it.elem.(int)
		require.True(t, ok)
		a |= uint64(1) << i
	}
	assert.Equal(t, 64, n, "h=%v a=%b", h, a)
	assert.Zero(t, ^a)
}

var prepopNode = memoizePrepop(func(n int) interface{} {
	var h *node
	for i := 0; i < n; i++ {
		h, _ = h.put(KV(i, i*i))
	}
	return h
})

func benchmarkInsertFrozenNode(b *testing.B, n int) {
	h := prepopNode(n).(*node)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		kv := KV(i, i*i)
		h.put(kv)
	}
}

func BenchmarkInsertFrozenNode0(b *testing.B) {
	benchmarkInsertFrozenNode(b, 0)
}

func BenchmarkInsertFrozenNode1k(b *testing.B) {
	benchmarkInsertFrozenNode(b, 1<<10)
}

func BenchmarkInsertFrozenNode1M(b *testing.B) {
	benchmarkInsertFrozenNode(b, 1<<20)
}
