package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

// Derived types for testing type independence.
type (
	ID    int
	Name  string
	Score float64
)

// setOps exercises all core Set operations for a given type.
func setOps[T comparable](t *testing.T, elems []T) {
	t.Helper()

	n := len(elems)
	if n < 4 {
		t.Fatal("need at least 4 elements")
	}

	// With + Has + Count
	var s frozen.Set[T]
	for _, e := range elems {
		s = s.With(e)
	}
	test.Equal(t, n, s.Count())
	for _, e := range elems {
		test.True(t, s.Has(e), "missing element after With")
	}

	// Idempotent With (no-op)
	for _, e := range elems {
		s2 := s.With(e)
		test.Equal(t, n, s2.Count())
	}

	// Without + idempotent Without (absent element)
	removed := s.Without(elems[0])
	test.Equal(t, n-1, removed.Count())
	test.False(t, removed.Has(elems[0]))
	test.Equal(t, n-1, removed.Without(elems[0]).Count())
	for _, e := range elems[1:] {
		test.True(t, removed.Has(e))
	}

	// Equal
	var s2 frozen.Set[T]
	for i := n - 1; i >= 0; i-- {
		s2 = s2.With(elems[i])
	}
	test.True(t, s.Equal(s2))

	// Union
	half := n / 2
	a := frozen.NewSet(elems[:half]...)
	b := frozen.NewSet(elems[half:]...)
	u := a.Union(b)
	test.Equal(t, n, u.Count())
	test.True(t, u.Equal(s))

	// Intersection
	overlap := frozen.NewSet(elems[:half]...)
	inter := s.Intersection(overlap)
	test.Equal(t, half, inter.Count())

	// Difference
	diff := s.Difference(overlap)
	test.Equal(t, n-half, diff.Count())
	for _, e := range elems[half:] {
		test.True(t, diff.Has(e))
	}
}

// mapOps exercises all core Map operations for a given key/value type.
func mapOps[K comparable, V comparable](t *testing.T, keys []K, vals []V) {
	t.Helper()

	n := len(keys)
	if n < 4 || len(vals) < n {
		t.Fatal("need at least 4 key-value pairs")
	}

	// With + Get + Count
	var m frozen.Map[K, V]
	for i := 0; i < n; i++ {
		m = m.With(keys[i], vals[i])
	}
	test.Equal(t, n, m.Count())
	for i := 0; i < n; i++ {
		got, ok := m.Get(keys[i])
		test.True(t, ok, "missing key after With")
		test.Equal(t, vals[i], got)
	}

	// Has
	for _, k := range keys {
		test.True(t, m.Has(k))
	}

	// Idempotent With (same key+value)
	for i := 0; i < n; i++ {
		m2 := m.With(keys[i], vals[i])
		test.Equal(t, n, m2.Count())
	}

	// Without
	removed := m.Without(keys[0])
	test.Equal(t, n-1, removed.Count())
	test.False(t, removed.Has(keys[0]))

	// Without absent key
	absent := m.Without(keys[0]).Without(keys[0])
	test.Equal(t, n-1, absent.Count())

	// Equal
	var m2 frozen.Map[K, V]
	for i := n - 1; i >= 0; i-- {
		m2 = m2.With(keys[i], vals[i])
	}
	test.True(t, m.Equal(m2))
}

func TestTypeMatrix_Set_Int(t *testing.T) {
	t.Parallel()
	setOps(t, []int{1, 2, 3, 4, 5, 100, 200, 300})
}

func TestTypeMatrix_Set_String(t *testing.T) {
	t.Parallel()
	setOps(t, []string{"a", "bb", "ccc", "dddd", "hello", "world", "foo", "bar"}) //nolint:goconst // test data
}

func TestTypeMatrix_Set_Float64(t *testing.T) {
	t.Parallel()
	setOps(t, []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8})
}

func TestTypeMatrix_Set_DerivedInt(t *testing.T) {
	t.Parallel()
	setOps(t, []ID{1, 2, 3, 4, 5, 100, 200, 300})
}

func TestTypeMatrix_Set_DerivedString(t *testing.T) {
	t.Parallel()
	setOps(t, []Name{"a", "bb", "ccc", "dddd", "hello", "world", "foo", "bar"})
}

func TestTypeMatrix_Set_DerivedFloat64(t *testing.T) {
	t.Parallel()
	setOps(t, []Score{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8})
}

type point struct {
	X, Y int
}

func TestTypeMatrix_Set_Struct(t *testing.T) {
	t.Parallel()
	setOps(t, []point{{1, 2}, {3, 4}, {5, 6}, {7, 8}})
}

func TestTypeMatrix_Map_IntInt(t *testing.T) {
	t.Parallel()
	mapOps(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, []int{10, 20, 30, 40, 50, 60, 70, 80})
}

func TestTypeMatrix_Map_StringString(t *testing.T) {
	t.Parallel()
	mapOps(t,
		[]string{"a", "bb", "ccc", "dddd", "e", "ff", "ggg", "hhhh"},
		[]string{"va", "vb", "vc", "vd", "ve", "vf", "vg", "vh"},
	)
}

func TestTypeMatrix_Map_Float64Int(t *testing.T) {
	t.Parallel()
	mapOps(t,
		[]float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8},
		[]int{10, 20, 30, 40, 50, 60, 70, 80},
	)
}

func TestTypeMatrix_Map_DerivedIntInt(t *testing.T) {
	t.Parallel()
	mapOps(t, []ID{1, 2, 3, 4, 5, 6, 7, 8}, []int{10, 20, 30, 40, 50, 60, 70, 80})
}

func TestTypeMatrix_Map_DerivedStringString(t *testing.T) {
	t.Parallel()
	mapOps(t,
		[]Name{"a", "bb", "ccc", "dddd", "e", "ff", "ggg", "hhhh"},
		[]string{"va", "vb", "vc", "vd", "ve", "vf", "vg", "vh"},
	)
}

func TestTypeMatrix_Map_DerivedFloat64Int(t *testing.T) {
	t.Parallel()
	mapOps(t,
		[]Score{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8},
		[]int{10, 20, 30, 40, 50, 60, 70, 80},
	)
}

func TestTypeMatrix_Map_StructInt(t *testing.T) {
	t.Parallel()
	mapOps(t,
		[]point{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
		[]int{10, 20, 30, 40},
	)
}

// TestTypeMatrix_DerivedTypeIndependence verifies that Set[int] and Set[ID]
// with the same underlying values produce independent sets that function
// correctly. Since these are separate generic instantiations, they cannot
// interfere, but this test confirms the hash/equality dispatch resolves
// correctly for each type.
func TestTypeMatrix_DerivedTypeIndependence(t *testing.T) {
	t.Parallel()

	vals := []int{1, 2, 3, 4, 5}

	// Build Set[int] and Set[ID] with same underlying values.
	si := frozen.NewSet(vals...)

	ids := make([]ID, len(vals))
	for i, v := range vals {
		ids[i] = ID(v)
	}
	sd := frozen.NewSet(ids...)

	// Both should have the same count.
	test.Equal(t, len(vals), si.Count())
	test.Equal(t, len(vals), sd.Count())

	// Each should find its own elements.
	for _, v := range vals {
		test.True(t, si.Has(v))
	}
	for _, id := range ids {
		test.True(t, sd.Has(id))
	}

	// Operations work independently.
	si2 := si.With(100)
	sd2 := sd.With(ID(100))
	test.Equal(t, len(vals)+1, si2.Count())
	test.Equal(t, len(vals)+1, sd2.Count())

	si3 := si.Without(1)
	sd3 := sd.Without(ID(1))
	test.Equal(t, len(vals)-1, si3.Count())
	test.Equal(t, len(vals)-1, sd3.Count())
	test.False(t, si3.Has(1))
	test.False(t, sd3.Has(ID(1)))
}

// TestTypeMatrix_MapDerivedKeyIndependence verifies Map[int,V] and Map[ID,V]
// with same underlying key values function independently.
func TestTypeMatrix_MapDerivedKeyIndependence(t *testing.T) {
	t.Parallel()

	mi := frozen.NewMap(
		frozen.KV(1, "a"),
		frozen.KV(2, "b"),
		frozen.KV(3, "c"),
	)

	md := frozen.NewMap(
		frozen.KV(ID(1), "x"),
		frozen.KV(ID(2), "y"),
		frozen.KV(ID(3), "z"),
	)

	test.Equal(t, 3, mi.Count())
	test.Equal(t, 3, md.Count())

	// Values are independent.
	vi, _ := mi.Get(1)
	test.Equal(t, "a", vi)
	vd, _ := md.Get(ID(1))
	test.Equal(t, "x", vd)

	// Mutations are independent.
	mi2 := mi.With(4, "d")
	md2 := md.With(ID(4), "w")
	test.Equal(t, 4, mi2.Count())
	test.Equal(t, 4, md2.Count())
	test.False(t, mi.Has(4))
	test.False(t, md.Has(ID(4)))
}

// TestTypeMatrix_LargeRoundTrip verifies round-trip correctness at scale
// for derived types (insert N, verify all present, verify count, remove all).
func TestTypeMatrix_LargeRoundTrip(t *testing.T) {
	t.Parallel()

	const n = 1000

	// Set[ID]
	var s frozen.Set[ID]
	for i := 0; i < n; i++ {
		s = s.With(ID(i))
	}
	test.Equal(t, n, s.Count())
	for i := 0; i < n; i++ {
		test.True(t, s.Has(ID(i)), "Set[ID] missing", i)
	}

	// Remove all
	for i := 0; i < n; i++ {
		s = s.Without(ID(i))
	}
	test.Equal(t, 0, s.Count())

	// Map[Name, Score]
	var m frozen.Map[Name, Score]
	for i := 0; i < n; i++ {
		m = m.With(Name("key"+string(rune('A'+i%26))+string(rune('0'+i/26%10))), Score(float64(i)*1.5))
	}
	// Count may be less than n due to key collisions in the name scheme,
	// but all present keys should round-trip correctly.
	for i := m.Range(); i.Next(); {
		k, v := i.Entry()
		got, ok := m.Get(k)
		test.True(t, ok)
		test.Equal(t, v, got)
	}
}
