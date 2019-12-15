package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	var b SetBuilder
	for i := 0; i < 35; i++ {
		b.Add(i)
	}
	for i := 10; i < 25; i++ {
		b.Remove(i)
	}
	for i := 5; i < 15; i++ {
		b.Add(i)
	}
	for i := 20; i < 30; i++ {
		b.Remove(i)
	}
	m := b.Finish()

	for i := 0; i < 35; i++ {
		switch {
		case i < 15:
			assertSetHas(t, m, i)
		case i < 30:
			assertSetNotHas(t, m, i)
		default:
			assertSetHas(t, m, i)
		}
	}
}
