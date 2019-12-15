package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b MapBuilder
	assertMapEqual(t, Map{}, b.Finish())
}

func TestMapBuilder(t *testing.T) {
	t.Parallel()

	var b MapBuilder
	for i := 0; i < 10; i++ {
		b.Put(i, i*i)
	}
	m := b.Finish()
	assert.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		assertMapHas(t, m, i, i*i)
	}
}

func TestMapBuilderRemove(t *testing.T) {
	t.Parallel()

	var b MapBuilder
	for i := 0; i < 15; i++ {
		b.Put(i, i*i)
	}
	for i := 5; i < 10; i++ {
		b.Remove(i)
	}
	m := b.Finish()

	assert.Equal(t, 10, m.Count())
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

func TestMapBuilderWithRedundantAddsAndRemoves(t *testing.T) {
	t.Parallel()

	var b MapBuilder
	for i := 0; i < 35; i++ {
		b.Put(i, i*i)
	}
	for i := 10; i < 25; i++ {
		b.Remove(i)
	}
	for i := 5; i < 15; i++ {
		b.Put(i, i*i*i)
	}
	for i := 20; i < 30; i++ {
		b.Remove(i)
	}
	m := b.Finish()

	for i := 0; i < 35; i++ {
		switch {
		case i < 5:
			assertMapHas(t, m, i, i*i)
		case i < 15:
			assertMapHas(t, m, i, i*i*i)
		case i < 30:
			assertMapNotHas(t, m, i)
		default:
			assertMapHas(t, m, i, i*i)
		}
	}
}
