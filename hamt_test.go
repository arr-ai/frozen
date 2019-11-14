package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHamtEmpty(t *testing.T) {
	h := hamt{}
	assert.Zero(t, h.count())
}

func TestHamtSmall(t *testing.T) {
	h := hamt{}
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
	hh := []hamt{}
	h := hamt{}
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
	hh := []hamt{}
	h := hamt{}
	for i := 0; i < 100; i++ {
		v, has := h.get(i)
		assert.False(t, has, "%v", v)
		hh = append(hh, h)
		h = h.put(i, i*i)
	}
	for i, h := range hh {
		for j := 0; j < i; j++ {
			v, has := h.get(j)
			if assert.True(t, has) {
				assert.Equal(t, j*j, v.(int))
			}
		}
	}
}

func TestHamtDelete(t *testing.T) {
	h := hamt{}
	for i := 0; i < 1000; i++ {
		h = h.put(i, i*i)
	}

	d := h
	for i := 0; i < 1000; i++ {
		assert.Equal(t, 1000-i, d.count())
		assert.Equal(t, h.count() == 0, h.isEmpty())
		_, has := d.get(i)
		assert.True(t, has, "%v", i)
		d = d.delete(i)
		_, has = d.get(i)
		assert.False(t, has, "%v", i)
	}
	assert.Zero(t, d.count())
	assert.True(t, d.isEmpty())

	d = h
	for i := 999; i >= 0; i-- {
		assert.Equal(t, i+1, d.count())
		assert.Equal(t, h.count() == 0, h.isEmpty())
		_, has := d.get(i)
		assert.True(t, has, "%v", i)
		d = d.delete(i)
		_, has = d.get(i)
		assert.False(t, has, "%v", i)
	}
	assert.Zero(t, d.count())
	assert.True(t, d.isEmpty())
}

func TestHamtIter(t *testing.T) {
	a := make([]int, 1000)
	h := hamt{}
	for i := 0; i < 1000; i++ {
		h = h.put(i, i*i)
	}
	for i := h.iterator(); i.next(); {
		a[i.e.key.(int)] = i.e.value.(int)
	}
	for i, n := range a {
		assert.Equal(t, i*i, n, "%v", i)
	}
}
