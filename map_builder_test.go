package frozen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/arr-ai/frozen"
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

func TestMapBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:cyclop
	t.Parallel()

	var b MapBuilder
	s := make([]interface{}, 35)
	requireMatch := func(format string, args ...interface{}) {
		for j, u := range s {
			v, has := b.Get(j)
			if u == nil {
				assert.Falsef(t, has, format+" j=%v v=%v", append(args, j, v)...)
			} else {
				assert.Truef(t, has, format+" j=%v", append(args, j)...)
			}
		}
	}

	put := func(i int, v int) {
		b.Put(i, v)
		s[i] = v
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
