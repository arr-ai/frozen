//nolint:dupl
package frozen

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrMapBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b StrMapBuilder
	assertStrMapEqual(t, StrMap{}, b.Finish())
}

func TestStrMapBuilder(t *testing.T) {
	t.Parallel()

	var b StrMapBuilder
	for i := 0; i < 10; i++ {
		b.Put(fmt.Sprintf("%d", i), i*i)
	}
	m := b.Finish()
	assert.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i)
	}
}

func TestStrMapBuilderRemove(t *testing.T) {
	t.Parallel()

	var b StrMapBuilder
	for i := 0; i < 15; i++ {
		b.Put(fmt.Sprintf("%d", i), i*i)
	}
	for i := 5; i < 10; i++ {
		b.Remove(fmt.Sprintf("%d", i))
	}
	m := b.Finish()

	assert.Equal(t, 10, m.Count())
	for i := 0; i < 15; i++ {
		switch {
		case i < 5:
			assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i)
		case i < 10:
			assertStrMapNotHas(t, m, fmt.Sprintf("%d", i))
		default:
			assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i)
		}
	}
}

func TestStrMapBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:funlen
	t.Parallel()

	var b StrMapBuilder
	s := make([]interface{}, 35)
	requireMatch := func(format string, args ...interface{}) {
		for j, u := range s {
			v, has := b.Get(fmt.Sprintf("%d", j))
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
		b.Put(fmt.Sprintf("%d", i), v)
		s[i] = v
	}
	remove := func(i int) {
		b.Remove(fmt.Sprintf("%d", i))
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
			assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i)
		case i < 15:
			assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i*i)
		case i < 30:
			assertStrMapNotHas(t, m, fmt.Sprintf("%d", i))
		default:
			assertStrMapHas(t, m, fmt.Sprintf("%d", i), i*i)
		}
	}
}

func TestStrMapMarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(NewStrMap(StrKV("a", 2), StrKV("b", 4), StrKV("c", 2)))
	if assert.NoError(t, err) {
		var s map[string]float64
		require.NoError(t, json.Unmarshal(j, &s))
		assert.Equal(t, map[string]float64{"a": 2, "b": 4, "c": 2}, s)
	}
}
