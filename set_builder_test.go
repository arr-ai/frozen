package frozen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/arr-ai/frozen"
)

func TestSetBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	assertSetEqual(t, Set{}, b.Finish())
}

func TestSetBuilder(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	for i := 0; i < 10; i++ {
		b.Add(i)
	}
	m := b.Finish()
	assert.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		assert.True(t, m.Has(i))
	}
}

func TestSetBuilderIncremental(t *testing.T) {
	t.Parallel()

	replayable(true, func(r replayer) {
		N := 1_000
		if testing.Short() {
			N /= 10
		}
		arr := make([]interface{}, 0, N)
		for i := 0; i < N; i++ {
			arr = append(arr, i)
		}

		for i := N - 1; i >= 0; i-- {
			i := i
			corpus := arr[i:]
			assertSameElements(t, corpus, NewSet(arr[i:]...).Elements())
		}
	})
}

func TestSetBuilderRemove(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	for i := 0; i < 15; i++ {
		b.Add(i)
	}
	for i := 5; i < 10; i++ {
		b.Remove(i)
	}
	m := b.Finish()

	assert.Equal(t, 10, m.Count())
	for i := 0; i < 15; i++ {
		switch {
		case i < 5:
			assertSetHas(t, m, i)
		case i < 10:
			assertSetNotHas(t, m, i)
		default:
			assertSetHas(t, m, i)
		}
	}
}

func TestSetBuilderWithRedundantAddsAndRemoves(t *testing.T) {
	t.Parallel()

	replayable(true, func(r replayer) {
		var b SetBuilder

		s := uint64(0)

		assertMatch := func(format string, args ...interface{}) bool {
			for j := 0; j < 60; j++ {
				if !assert.Equalf(t, s&(uint64(1)<<uint(j)) != 0, b.Has(j), format+" j=%v", append(args, j)...) {
					return false
				}
			}
			return true
		}

		add := func(i int) {
			b.Add(i)
			s |= uint64(1) << uint(i)
		}

		remove := func(i int) {
			b.Remove(i)
			s &^= uint64(1) << uint(i)
		}

		assertMatch("")
		for i := 0; i < 60; i++ {
			add(i)
			assertMatch("i=%v", i)
		}
		for i := 20; i < 50; i++ {
			m := r.mark(i)
			remove(i)
			if !assertMatch("i=%v", i) {
				r.replayTo(m)
			}
		}

		for i := 10; i < 30; i++ {
			add(i)
			assertMatch("i=%v", i)
		}
		for i := 40; i < 55; i++ {
			remove(i)
			assertMatch("i=%v", i)
		}
		m := b.Finish()

		for i := 0; i < 60; i++ {
			switch {
			case i < 30:
				assertSetHas(t, m, i)
			case i < 55:
				assertSetNotHas(t, m, i)
			default:
				assertSetHas(t, m, i)
			}
		}
	})
}
