package value_test

import (
	"testing"

	"github.com/arr-ai/frozen/internal/pkg/value"
)

// --- helper types ---

// equalerInt wraps an int and implements Equaler[equalerInt].
type equalerInt struct{ v int }

func (e equalerInt) Equal(other equalerInt) bool { return e.v == other.v }

// samerStr wraps a string and implements Samer.
type samerStr struct{ s string }

func (s samerStr) Same(a any) bool {
	other, ok := a.(samerStr)
	return ok && s.s == other.s
}

// --- Equal ---

func TestEqualInt(t *testing.T) {
	t.Parallel()

	if !value.Equal(1, 1) {
		t.Error("Equal(1, 1) must be true")
	}
	if value.Equal(1, 2) {
		t.Error("Equal(1, 2) must be false")
	}
}

func TestEqualString(t *testing.T) {
	t.Parallel()

	if !value.Equal("hello", "hello") {
		t.Error(`Equal("hello", "hello") must be true`)
	}
	if value.Equal("hello", "world") {
		t.Error(`Equal("hello", "world") must be false`)
	}
}

func TestEqualWithEqualerType(t *testing.T) {
	t.Parallel()

	a := equalerInt{3}
	b := equalerInt{3}
	c := equalerInt{4}

	if !value.Equal(a, b) {
		t.Error("Equal(equalerInt{3}, equalerInt{3}) must be true")
	}
	if value.Equal(a, c) {
		t.Error("Equal(equalerInt{3}, equalerInt{4}) must be false")
	}
}

func TestEqualWithSamerType(t *testing.T) {
	t.Parallel()

	a := samerStr{"foo"}
	b := samerStr{"foo"}
	c := samerStr{"bar"}

	if !value.Equal(a, b) {
		t.Error("Equal(samerStr{foo}, samerStr{foo}) must be true")
	}
	if value.Equal(a, c) {
		t.Error("Equal(samerStr{foo}, samerStr{bar}) must be false")
	}
}

// --- EqualFuncFor ---

func TestEqualFuncForInt(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[int]()

	if !eq(42, 42) {
		t.Error("EqualFuncFor[int]()(42, 42) must be true")
	}
	if eq(42, 43) {
		t.Error("EqualFuncFor[int]()(42, 43) must be false")
	}
}

func TestEqualFuncForString(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[string]()

	if !eq("abc", "abc") {
		t.Error(`EqualFuncFor[string]()("abc", "abc") must be true`)
	}
	if eq("abc", "def") {
		t.Error(`EqualFuncFor[string]()("abc", "def") must be false`)
	}
}

func TestEqualFuncForFloat64(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[float64]()

	if !eq(3.14, 3.14) {
		t.Error("EqualFuncFor[float64]()(3.14, 3.14) must be true")
	}
	if eq(3.14, 2.71) {
		t.Error("EqualFuncFor[float64]()(3.14, 2.71) must be false")
	}
}

func TestEqualFuncForEqualerType(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[equalerInt]()

	if !eq(equalerInt{5}, equalerInt{5}) {
		t.Error("EqualFuncFor[equalerInt]()(5, 5) must be true")
	}
	if eq(equalerInt{5}, equalerInt{6}) {
		t.Error("EqualFuncFor[equalerInt]()(5, 6) must be false")
	}
}

func TestEqualFuncForSamerType(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[samerStr]()

	if !eq(samerStr{"x"}, samerStr{"x"}) {
		t.Error("EqualFuncFor[samerStr]()({x}, {x}) must be true")
	}
	if eq(samerStr{"x"}, samerStr{"y"}) {
		t.Error("EqualFuncFor[samerStr]()({x}, {y}) must be false")
	}
}

// TestEqualFuncForBool exercises the scalar path for a 1-byte type.
func TestEqualFuncForBool(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[bool]()

	if !eq(true, true) {
		t.Error("EqualFuncFor[bool]()(true, true) must be true")
	}
	if eq(true, false) {
		t.Error("EqualFuncFor[bool]()(true, false) must be false")
	}
}

// TestEqualFuncForUint32 exercises the scalar path for a 4-byte type.
func TestEqualFuncForUint32(t *testing.T) {
	t.Parallel()

	eq := value.EqualFuncFor[uint32]()

	if !eq(100, 100) {
		t.Error("EqualFuncFor[uint32]()(100, 100) must be true")
	}
	if eq(100, 200) {
		t.Error("EqualFuncFor[uint32]()(100, 200) must be false")
	}
}

// --- uncomparable dynamic types (regression: comparing uncomparable type) ---

// uncomparableStruct embeds a slice, so its zero value is comparable
// (interfaces are always "comparable" per the language spec at compile
// time), but any two instances of it panic on == at runtime.
type uncomparableStruct struct {
	items []int
}

func TestEqualWithUncomparableDynamicTypeDoesNotPanic(t *testing.T) {
	t.Parallel()

	a := uncomparableStruct{items: []int{1, 2, 3}}
	b := uncomparableStruct{items: []int{1, 2, 3}}

	// The dynamic type (uncomparableStruct) implements neither Equaler nor
	// Samer and is not comparable with ==, so the only panic-free answer is
	// false, even though the two values are "the same" by content.
	if value.Equal(a, b) {
		t.Error("Equal(a, b) must be false for uncomparable dynamic types (no panic, but no equality either)")
	}
	if value.Equal(a, a) {
		t.Error("Equal(a, a) must be false for uncomparable dynamic types, even for the same value")
	}
}

func TestEqualFuncForAnyWithUncomparableDynamicTypeDoesNotPanic(t *testing.T) {
	t.Parallel()

	// T = any is the shape that triggers equalSlow: EqualFuncFor can't tell
	// from any's nil zero value whether real values will be comparable.
	eq := value.EqualFuncFor[any]()

	var a any = uncomparableStruct{items: []int{1, 2, 3}}
	var b any = uncomparableStruct{items: []int{4, 5, 6}}

	if eq(a, b) {
		t.Error("EqualFuncFor[any]()(a, b) must be false for uncomparable dynamic types")
	}
}

func TestEqualWithMismatchedUncomparableDynamicTypesDoesNotPanic(t *testing.T) {
	t.Parallel()

	var a any = uncomparableStruct{items: []int{1}}
	var b any = []int{1}

	if value.Equal(a, b) {
		t.Error("Equal(a, b) must be false for mismatched dynamic types")
	}
}

// TestEqualTableDriven provides a broader set of int cases.
func TestEqualTableDriven(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b int
		want bool
	}{
		{0, 0, true},
		{-1, -1, true},
		{1, 2, false},
		{-1, 1, false},
	}

	for _, tc := range cases {
		got := value.Equal(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("Equal(%d, %d) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
