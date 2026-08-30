package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

// str is uncomparable (it holds a slice) and defines equality over an
// interface that a Set[any] or Map[any, …] does not name — the shape of
// arr.ai's Value. Equality must still dispatch to its Equal method.
type equatable interface{ id() string }

type str struct{ runes []rune }

func (s str) id() string { return string(s.runes) }

func (s str) Equal(o equatable) bool { return o != nil && s.id() == o.id() }

func (s str) Hash(seed uintptr) uintptr {
	h := seed
	for _, r := range s.runes {
		h = h*1099511628211 ^ uintptr(uint32(r)) //nolint:gosec // test hash, value range irrelevant
	}
	return h
}

func newStr(s string) str { return str{runes: []rune(s)} }

// A slice-typed element must not panic: the comparability probe has to
// recover properly rather than let its own panic escape.
type slice []int

func (s slice) Equal(o any) bool {
	t, is := o.(slice)
	if !is || len(s) != len(t) {
		return false
	}
	for i, v := range s {
		if v != t[i] {
			return false
		}
	}
	return true
}

func (s slice) Hash(seed uintptr) uintptr {
	h := seed
	for _, v := range s {
		h = h*1099511628211 ^ uintptr(v)
	}
	return h
}

func TestSetOfSliceTypeDoesNotPanic(t *testing.T) {
	t.Parallel()

	var s frozen.Set[slice]
	s = s.With(slice{1, 2, 3}).With(slice{4, 5}).With(slice{1, 2, 3})
	test.Equal(t, 2, s.Count())
	test.True(t, s.Has(slice{1, 2, 3}))
	test.True(t, s.Has(slice{4, 5}))
	test.False(t, s.Has(slice{9}))
}

// An array of interfaces is comparable by type, so == compiles, but panics
// once the interfaces hold uncomparable values.
type pair [2]equatable

func (p pair) Equal(o any) bool {
	q, is := o.(pair)
	return is && p[0].id() == q[0].id() && p[1].id() == q[1].id()
}

func (p pair) Hash(seed uintptr) uintptr {
	return newStr(p[0].id()).Hash(newStr(p[1].id()).Hash(seed))
}

func TestSetOfArrayOfInterfacesDoesNotPanic(t *testing.T) {
	t.Parallel()

	a := pair{newStr("x"), newStr("y")}
	b := pair{newStr("p"), newStr("q")}
	var s frozen.Set[pair]
	s = s.With(a).With(b).With(pair{newStr("x"), newStr("y")})
	test.Equal(t, 2, s.Count())
	test.True(t, s.Has(a))
	test.True(t, s.Has(pair{newStr("x"), newStr("y")}))
	test.False(t, s.Has(pair{newStr("x"), newStr("z")}))
}

// The case that matters most: an `any`-keyed collection whose elements
// define Equal over their own interface. Neither Equaler[any] nor Samer
// matches, and the type is uncomparable, so equality must dispatch on the
// dynamic type rather than silently reporting equal values as different.
func TestAnyContainerUsesDynamicEqual(t *testing.T) {
	t.Parallel()

	var s frozen.Set[any]
	s = s.With(any(newStr("hello"))).With(any(newStr("world")))
	test.Equal(t, 2, s.Count())
	test.True(t, s.Has(any(newStr("hello"))), "an equal value must be found")
	test.False(t, s.Has(any(newStr("nope"))))
	s = s.With(any(newStr("hello")))
	test.Equal(t, 2, s.Count(), "adding an equal value must not grow the set")

	var m frozen.Map[any, int]
	m = m.With(any(newStr("k")), 1)
	v, has := m.Get(any(newStr("k")))
	test.True(t, has, "an equal key must be found")
	test.Equal(t, 1, v)
	m = m.With(any(newStr("k")), 2)
	test.Equal(t, 1, m.Count(), "rebinding an equal key must not grow the map")
	v, _ = m.Get(any(newStr("k")))
	test.Equal(t, 2, v)
}

// Builders compare through mapEntry.Equal rather than the function
// EqualFuncFor returns, so they need the same dispatch: SetGroupBy and the
// map/set builders must deduplicate keys whose type defines Equal over its
// own interface.
func TestBuildersUseDynamicEqual(t *testing.T) {
	t.Parallel()

	var mb frozen.MapBuilder[any, int]
	mb.Put(any(newStr("k")), 1)
	mb.Put(any(newStr("k")), 2)
	m := mb.Finish()
	test.Equal(t, 1, m.Count(), "equal keys must collapse in a MapBuilder")
	v, has := m.Get(any(newStr("k")))
	test.True(t, has)
	test.Equal(t, 2, v)

	var sb frozen.SetBuilder[any]
	sb.Add(any(newStr("e")))
	sb.Add(any(newStr("e")))
	test.Equal(t, 1, sb.Finish().Count(), "equal elements must collapse in a SetBuilder")

	// SetGroupBy is the builder path arr.ai's relation indexes rely on.
	elems := frozen.NewSet[any](
		any(newStr("a")), any(newStr("b")), any(newStr("c")),
	)
	groups := frozen.SetGroupBy(elems, func(e any) any {
		if e.(str).id() == "c" { //nolint:forcetypeassert
			return any(newStr("x"))
		}
		return any(newStr("y"))
	})
	test.Equal(t, 2, groups.Count(), "equal group keys must collapse")
	g, has := groups.Get(any(newStr("y")))
	test.True(t, has)
	test.Equal(t, 2, g.Count())
}
