package frozen_test

import (
	"sync/atomic"
	"testing"

	"github.com/arr-ai/hash/hash128"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

// h128Val implements only Hashable128.
type h128Val struct{ n int }

func (v h128Val) Hash128() hash128.H128 { return hash128.Int(v.n) }
func (v h128Val) Equal(o h128Val) bool  { return v.n == o.n }

// bothVal implements Hashable128 and the seeded Hashable; the seeded hash is
// deliberately unusable so the test fails loudly if it is ever consulted.
type bothVal struct{ n int }

var seededCalls int64

func (v bothVal) Hash128() hash128.H128 { return hash128.Int(v.n) }
func (v bothVal) Hash(_ uintptr) uintptr {
	atomic.AddInt64(&seededCalls, 1)
	return 0
}
func (v bothVal) Equal(o bothVal) bool { return v.n == o.n }

func h128Vals(n int) []h128Val {
	vals := make([]h128Val, n)
	for i := range vals {
		vals[i] = h128Val{i}
	}
	return vals
}

func TestHashable128_SetMatrix(t *testing.T) {
	t.Parallel()
	setOps(t, h128Vals(64))
}

func TestHashable128_MapMatrix(t *testing.T) {
	t.Parallel()
	vals := make([]int, 64)
	for i := range vals {
		vals[i] = i * 3
	}
	mapOps(t, h128Vals(64), vals)
}

func TestHashable128_TakesPrecedenceOverSeeded(t *testing.T) {
	t.Parallel()
	var s frozen.Set[bothVal]
	for i := 0; i < 100; i++ {
		s = s.With(bothVal{i})
	}
	test.True(t, s.Has(bothVal{42}))
	s2 := s.Where(func(v bothVal) bool { return v.n%2 == 0 })
	test.Equal(t, 50, s2.Count())
	var m frozen.Map[bothVal, int]
	for i := 0; i < 100; i++ {
		m = m.With(bothVal{i}, i)
	}
	v, has := m.Get(bothVal{7})
	test.True(t, has)
	test.Equal(t, 7, v)
	test.Equal(t, int64(0), atomic.LoadInt64(&seededCalls), "seeded Hash must not be used")
}

func TestHashable128_DynamicDispatchThroughAny(t *testing.T) {
	t.Parallel()
	// T = any: the element's dynamic type decides how it is hashed.
	var s frozen.Set[any]
	for i := 0; i < 50; i++ {
		s = s.With(bothVal{i})
	}
	test.True(t, s.Has(bothVal{9}))
	test.False(t, s.Has(bothVal{99}))
	test.Equal(t, int64(0), atomic.LoadInt64(&seededCalls), "seeded Hash must not be used")

	var m frozen.Map[any, string]
	m = m.With(bothVal{1}, "one").With("two", "two").With(3, "three")
	for _, k := range []any{bothVal{1}, "two", 3} {
		_, has := m.Get(k)
		test.True(t, has)
	}
}

func TestHash128_SetAndMapAreOrderIndependent(t *testing.T) {
	t.Parallel()
	a := frozen.NewSet(1, 2, 3, 4, 5)
	b := frozen.NewSet(5, 4, 3, 2, 1)
	test.Equal(t, a.Hash128(), b.Hash128())
	test.NotEqual(t, a.Hash128(), frozen.NewSet(1, 2, 3, 4).Hash128())
	test.True(t, frozen.Set[int]{}.Hash128().IsZero())

	var m1, m2 frozen.Map[string, int]
	m1 = m1.With("a", 1).With("b", 2)
	m2 = m2.With("b", 2).With("a", 1)
	test.Equal(t, m1.Hash128(), m2.Hash128())
	test.NotEqual(t, m1.Hash128(), m1.Without("a").Hash128())
}

func TestHashable128_SetOfSets(t *testing.T) {
	t.Parallel()
	inner := func(xs ...int) frozen.Set[int] { return frozen.NewSet(xs...) }
	outer := frozen.NewSet(inner(1, 2), inner(3), inner(1, 2))
	test.Equal(t, 2, outer.Count())
	test.True(t, outer.Has(inner(2, 1)))
}
