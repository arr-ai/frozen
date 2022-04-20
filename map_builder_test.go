package frozen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/arr-ai/frozen"
)

func TestMapBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b MapBuilder[int, int]
	assertMapEqual(t, Map[int, int]{}, b.Finish())
}

func TestMapBuilder(t *testing.T) {
	t.Parallel()

	var b MapBuilder[int, int]
	for i := 0; i < 10; i++ {
		b.Put(i, i*i)
	}
	m := b.Finish()
	assert.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		assertMapHas(t, m, i, i*i)
	}
}

func TestMapBuilderSimple(t *testing.T) {
	t.Parallel()

	var b MapBuilder[int, int]
	b.Put(0, 0)
	assert.Equal(t, 1, b.Count())
	b.Remove(0)
	assert.Equal(t, 0, b.Count())

	b.Put(0, 0)
	b.Put(1, 1)
	b.Put(2, 2)
	assert.Equal(t, 3, b.Count())

	b.Put(3, 3)
	b.Put(4, 4)
	b.Put(5, 5)
	assert.Equal(t, 6, b.Count())
	b.Remove(5)
	assert.Equal(t, 5, b.Count())
}

func TestMapBuilderRemove(t *testing.T) {
	t.Parallel()

	var b MapBuilder[int, int]

	for i := 0; i < 15; i++ {
		b.Put(i, i*i)
	}
	for i := 5; i < 10; i++ {
		b.Remove(i)
		v, has := b.Get(i)
		require.False(t, has, v)
	}
	m := b.Finish()

	assert.Equal(t, 10, m.Count(), m)
	for i := 0; i < 15; i++ {
		switch {
		case i < 5:
			assertMapHas(t, m, i, i*i)
		case i < 10:
			assertMapNotHas(t, m, i)
		default:
			assertMapHas(t, m, i, i*i)
		}
	}
}

func TestMapBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:cyclop
	t.Parallel()

	var b MapBuilder[int, int]
	s := make([]*int, 35)
	requireMatch := func(format string, args ...any) {
		for j, u := range s {
			v, has := b.Get(j)
			if u == nil {
				require.Falsef(t, has, format+" j=%v v=%v", append(args, j, v)...)
			} else {
				require.Truef(t, has, format+" j=%v", append(args, j)...)
			}
		}
	}

	put := func(i int, v int) {
		b.Put(i, v)
		s[i] = &v
	}
	remove := func(i int) {
		b.Remove(i)
		s[i] = nil
	}

	for i := 0; i < 35; i++ {
		put(i, i*i)
		requireMatch("i=%v", i)
	}
	for i := 10; i < 25; i++ {
		remove(i)
		requireMatch("i=%v", i)
	}
	for i := 5; i < 15; i++ {
		put(i, i*i*i)
		requireMatch("i=%v", i)
	}
	for i := 20; i < 30; i++ {
		remove(i)
		requireMatch("i=%v", i)
	}
	m := b.Finish()

loop:
	for i := 0; i < 35; i++ {
		switch {
		case i < 5:
			if !assertMapHas(t, m, i, i*i) {
				break loop
			}
		case i < 15:
			if !assertMapHas(t, m, i, i*i*i) {
				break loop
			}
		case i < 30:
			if !assertMapNotHas(t, m, i) {
				break loop
			}
		default:
			if !assertMapHas(t, m, i, i*i) {
				break loop
			}
		}
	}
}
