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

func TestMapBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:funlen
	t.Parallel()

	var b MapBuilder
	s := make([]interface{}, 35)
	requireMatch := func(format string, args ...interface{}) {
		for j, u := range s {
			v, has := b.Get(j)
			if u == nil {
				if !assert.Falsef(t, has, format+" j=%v v=%v", append(args, j, v)...) {
					t.Log(b.root)
					t.FailNow()
				}
			} else {
				if !assert.Truef(t, has, format+" j=%v", append(args, j)...) {
					t.Log(b.root)
					t.FailNow()
				}
				if !assert.Equal(t, u, v, "h(u)=%v h(v)=%v", newHasher(u, 0), newHasher(v, 0)) {
					t.Log(b.root)
					t.FailNow()
				}
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
		t.Log(b.root)
		put(i, i*i)
		requireMatch("i=%v", i)
	}
	for i := 10; i < 25; i++ {
		t.Log(b.root)
		remove(i)
		requireMatch("i=%v", i)
	}
	for i := 5; i < 15; i++ {
		t.Log(b.root)
		put(i, i*i*i)
		requireMatch("i=%v", i)
	}
	for i := 20; i < 30; i++ {
		t.Log(b.root)
		remove(i)
		requireMatch("i=%v", i)
	}
	t.Log(b.root)
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
