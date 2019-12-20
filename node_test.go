package frozen

import (
	"fmt"
	"math/bits"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeSmall(t *testing.T) {
	t.Parallel()

	putter := newUnionComposer(0)
	var h *node
	h = h.apply(putter, KV("foo", 42))
	assert.NotNil(t, h)
	assert.Equal(t, 0, putter.count())
	h = h.apply(putter, KV("bar", 43))
	assert.NotNil(t, h)
	assert.Equal(t, 0, putter.count())
	h = h.apply(putter, KV("foo", 44))
	assert.NotNil(t, h)
	assert.Equal(t, -1, putter.count())
}

func TestNodeEqual(t *testing.T) {
	t.Parallel()

	putter := newUnionComposer(0)

	foo1 := empty.apply(putter, KV("foo", 42))
	foo2 := empty.apply(putter, KV("foo", 100))
	assert.True(t, foo1.equal(foo2,
		func(a, b interface{}) bool { return a.(KeyValue).Key == b.(KeyValue).Key },
	))

	nodes := []*node{
		empty,
		foo1,
		foo2,
		empty.apply(putter, KV("bar", 42)),
		empty.apply(putter, KV("foo", 42)).apply(putter, KV("bar", 42)),
	}
	keys := []string{"a", "b", "c", "d", "e", "f", "g"}
	for i := 1; i < 1<<len(keys); i++ {
		n := empty
		for m := i; m != 0; m &= m - 1 {
			j := bits.TrailingZeros64(uint64(m))
			n = n.apply(putter, KV(keys[j], j))
		}
		nodes = append(nodes, n)
	}

	same := func(a, b interface{}) bool { return a == b }

	for i, a := range nodes {
		for j, b := range nodes {
			assert.Equal(t, i == j, a.equal(b, same), "i=%v j=%v a=%v b=%v", i, j, a, b)
		}
	}
}

func nodeStringRepr(a, b int, ha, hb hasher) string {
	if ha.hash() == hb.hash() {
		return nodeStringRepr(a, b, ha.next(a), hb.next(b))
	}
	children := []string{"∅", "∅", "∅", "∅", "∅", "∅", "∅", "∅"}
	children[ha.hash()%8] = fmt.Sprintf("%v", a)
	children[hb.hash()%8] = fmt.Sprintf("%v", b)
	return "[" + strings.Join(children, ",") + "]"
}

func TestNodeString(t *testing.T) {
	t.Parallel()

	putter := newUnionComposer(0)
	elems := []int{42, 43}
	a := empty
	for _, el := range elems {
		a = a.apply(putter, el)
	}

	expected := nodeStringRepr(elems[0], elems[1], newHasher(elems[0], 0), newHasher(elems[1], 0))
	assert.Equal(t, expected, a.String())
}

func TestNodeLarge(t *testing.T) {
	t.Parallel()

	hh := []*node{}
	var h *node
	putter := newUnionComposer(0)
	for i := 0; i < 500; i++ {
		i := i
		hh = append(hh, h)
		for j := 0; j < i; j++ {
			kv := h.get(KV(j, nil))
			if assert.NotNil(t, kv, "i=%v j=%v h=%v", i, j, h) {
				assert.Equal(t, j*j, kv.(KeyValue).Value, "i=%v j=%v", i, j)
			}
		}
		kv := h.get(KV(i, nil))
		if !assert.Nil(t, kv, "i=%v v=%v", i, kv) {
			h.get(KV(i, nil))
		}
		assert.NotPanics(t, func() {
			h = h.apply(putter, KV(i, i*i))
		}, "i=%v h=%v", i, h)
		assert.Equal(t, 0, putter.count(), "i=%v", i)
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

func TestNodeBatch(t *testing.T) {
	t.Parallel()

	elems := []interface{}{}
	putter := newUnionComposer(0)
	h := empty
	for i := 0; i < 500; i++ {
		elems = append(elems, i)
		h = h.apply(putter, i)
	}

	for i := range elems {
		assert.Equal(t, i, h.get(i))
	}
	assert.Nil(t, h.get(len(elems)))

	putter.mutate = true
	h = empty
	for i := 0; i < 500; i++ {
		h = h.apply(putter, i)
	}
	for i := range elems {
		assert.Equal(t, i, h.get(i))
	}
	assert.Nil(t, h.get(len(elems)))
}

func TestNodeGet(t *testing.T) {
	t.Parallel()

	hh := []*node{}
	var h *node
	putter := newUnionComposer(0)
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
		h = h.apply(putter, KV(i, i*i))
		if kv := h.get(KV(i, nil)); assert.NotNil(t, kv, "i=%v", i) {
			if !assert.Equal(t, i*i, kv.(KeyValue).Value, "i=%v", i) {
				hOld.apply(putter, KV(i, i*i))
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
	putter := newUnionComposer(0)
	for i := 0; i < N; i++ {
		h = h.apply(putter, i)
	}

	d := h
	deleter := newMinusComposer(0)
	for i := 0; i < N; i++ {
		assert.NotNil(t, h)
		require.NotNil(t, d.get(i), "i=%v", i)
		d = d.apply(deleter, i)
		assert.Nil(t, d.get(i), "i=%v", i)
	}
	assert.Nil(t, d)

	d = h
	for i := N; i > 0; {
		i--
		assert.NotNil(t, h)
		v := d.get(i)
		if assert.NotNil(t, v, "i=%v", i) {
			d = d.apply(deleter, i)
			assert.Nil(t, d.get(i), "i=%v", i)
		}
	}
	assert.Nil(t, d)
}

func TestNodeDeleteMissing(t *testing.T) {
	t.Parallel()

	putter := newUnionComposer(0)
	deleter := newMinusComposer(0)

	h := empty.apply(putter, "foo")
	h = h.apply(deleter, "bar")
	assert.NotNil(t, h)
	h = h.apply(deleter, "foo")
	assert.Nil(t, h)
}

func TestNodeIter(t *testing.T) {
	t.Parallel()

	putter := newUnionComposer(0)
	var h *node
	for i := 0; i < 64; i++ {
		h = h.apply(putter, i)
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
	putter := newUnionComposer(0)
	var h *node
	for i := 0; i < n; i++ {
		h = h.apply(putter, KV(i, i*i))
	}
	return h
})

func benchmarkInsertFrozenNode(b *testing.B, n int) {
	h := prepopNode(n).(*node)
	putter := newUnionComposer(0)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		kv := KV(i, i*i)
		h.apply(putter, kv)
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

func benchmarkPopulateFrozenNode(b *testing.B, n int, mutate bool) {
	data := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		data = append(data, i)
	}

	b.ResetTimer()
	putter := newUnionComposer(0)
	putter.mutate = mutate
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			empty.apply(putter, d)
		}
	}
}

func BenchmarkPopulateFrozenNode16(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<4, false)
}

func BenchmarkPopulateFrozenNode1k(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<10, false)
}

func BenchmarkPopulateFrozenNode1M(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<20, false)
}

func BenchmarkMutatingPopulateFrozenNode16(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<4, true)
}

func BenchmarkMutatingPopulateFrozenNode1k(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<10, true)
}

func BenchmarkMutatingPopulateFrozenNode1M(b *testing.B) {
	benchmarkPopulateFrozenNode(b, 1<<20, true)
}
