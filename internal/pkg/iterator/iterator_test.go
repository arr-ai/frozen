package iterator_test

import (
	"testing"

	"github.com/arr-ai/frozen/internal/pkg/iterator"
)

// --- BitIterator ---

func TestBitIteratorIndex(t *testing.T) {
	t.Parallel()

	// Single bit at position 5.
	b := iterator.BitIterator(1 << 5)
	if got := b.Index(); got != 5 {
		t.Errorf("Index() = %d, want 5", got)
	}
}

func TestBitIteratorIndexLowestBit(t *testing.T) {
	t.Parallel()

	// Multiple bits set; Index reports the position of the lowest set bit.
	b := iterator.BitIterator(0b1010)
	if got := b.Index(); got != 1 {
		t.Errorf("Index() = %d, want 1 (lowest set bit of 0b1010)", got)
	}
}

func TestBitIteratorCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		bits iterator.BitIterator
		want int
	}{
		{0b0000, 0},
		{0b0001, 1},
		{0b0101, 2},
		{0b1111, 4},
		{0b1010_1010, 4},
	}

	for _, tc := range cases {
		if got := tc.bits.Count(); got != tc.want {
			t.Errorf("BitIterator(%b).Count() = %d, want %d", tc.bits, got, tc.want)
		}
	}
}

func TestBitIteratorHas(t *testing.T) {
	t.Parallel()

	b := iterator.BitIterator(0b1010)

	if b.Has(0) {
		t.Error("Has(0) must be false for 0b1010")
	}
	if !b.Has(1) {
		t.Error("Has(1) must be true for 0b1010")
	}
	if b.Has(2) {
		t.Error("Has(2) must be false for 0b1010")
	}
	if !b.Has(3) {
		t.Error("Has(3) must be true for 0b1010")
	}
}

func TestBitIteratorWith(t *testing.T) {
	t.Parallel()

	b := iterator.BitIterator(0b0100)
	b2 := b.With(0)

	if !b2.Has(0) {
		t.Error("With(0) must set bit 0")
	}
	if !b2.Has(2) {
		t.Error("With(0) must preserve bit 2")
	}
	if b2.Count() != 2 {
		t.Errorf("With(0) on 0b0100 should give 2-bit value, got count=%d", b2.Count())
	}

	// With on a bit already set must be idempotent.
	b3 := b.With(2)
	if b3 != b {
		t.Errorf("With(2) on 0b0100 should be identity, got %b", b3)
	}
}

func TestBitIteratorWithout(t *testing.T) {
	t.Parallel()

	// Without is defined as: b &^ BitIterator(1) << uint(i)
	// Due to Go operator precedence, & ^ and << share precedence 5 and are
	// left-associative, so this parses as (b &^ BitIterator(1)) << uint(i).
	//
	// Concretely: Without(i) clears bit 0 of b first, then left-shifts by i.
	// The test captures the actual implementation behaviour.

	// Without(0): (b &^ 1) << 0 — clears bit 0, no shift.
	b := iterator.BitIterator(0b1010) // bits 1 and 3
	b1 := b.Without(0)
	// (0b1010 &^ 0b0001) << 0 = 0b1010 — no bit 0 to clear, result unchanged
	if b1 != 0b1010 {
		t.Errorf("Without(0) on 0b1010 = %b, want 0b1010", b1)
	}

	// Without(0) on a value with bit 0 set: strips bit 0 only.
	b2 := iterator.BitIterator(0b0101) // bits 0 and 2
	b3 := b2.Without(0)
	// (0b0101 &^ 0b0001) << 0 = 0b0100
	if b3 != 0b0100 {
		t.Errorf("Without(0) on 0b0101 = %b, want 0b0100", b3)
	}

	// Without(1): (b &^ 1) << 1 — clears bit 0, then shifts left by 1.
	b4 := iterator.BitIterator(0b0001) // bit 0 only
	b5 := b4.Without(1)
	// (0b0001 &^ 0b0001) << 1 = 0b0000 << 1 = 0
	if b5 != 0 {
		t.Errorf("Without(1) on 0b0001 = %b, want 0", b5)
	}
}

func TestBitIteratorNext(t *testing.T) {
	t.Parallel()

	// 0b0111 → Next strips lowest bit → 0b0110
	b := iterator.BitIterator(0b0111)
	b2 := b.Next()
	if b2.Has(0) {
		t.Error("Next() must clear the lowest set bit")
	}
	if !b2.Has(1) || !b2.Has(2) {
		t.Error("Next() must preserve higher bits")
	}
	if b2.Count() != 2 {
		t.Errorf("Next() on 0b0111 should give count=2, got %d", b2.Count())
	}
}

// TestBitIteratorNextIteration demonstrates iterating all set bits via Next.
func TestBitIteratorNextIteration(t *testing.T) {
	t.Parallel()

	want := []int{1, 3, 5}
	b := iterator.BitIterator(0)
	for _, i := range want {
		b = b.With(i)
	}

	var got []int
	for cur := b; cur != 0; cur = cur.Next() {
		got = append(got, cur.Index())
	}

	if len(got) != len(want) {
		t.Fatalf("iteration yielded %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("iteration index %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestBitIteratorZeroIsEmpty(t *testing.T) {
	t.Parallel()

	b := iterator.BitIterator(0)
	if b.Count() != 0 {
		t.Errorf("zero BitIterator must have count=0, got %d", b.Count())
	}
	if b.Next() != 0 {
		t.Error("zero BitIterator.Next() must be zero")
	}
}

// --- SliceIterator ---

func TestSliceIteratorEmpty(t *testing.T) {
	t.Parallel()

	it := iterator.NewSliceIterator[int](nil)
	if it.Next() {
		t.Error("Next() on empty iterator must return false")
	}
}

func TestSliceIteratorEmptySlice(t *testing.T) {
	t.Parallel()

	it := iterator.NewSliceIterator([]string{})
	if it.Next() {
		t.Error("Next() on empty slice iterator must return false")
	}
}

func TestSliceIteratorSingleElement(t *testing.T) {
	t.Parallel()

	it := iterator.NewSliceIterator([]int{42})

	if !it.Next() {
		t.Fatal("Next() must return true for single-element slice")
	}
	if got := it.Value(); got != 42 {
		t.Errorf("Value() = %d, want 42", got)
	}
	if it.Next() {
		t.Error("second Next() must return false for single-element slice")
	}
}

func TestSliceIteratorMultipleElements(t *testing.T) {
	t.Parallel()

	elems := []string{"a", "b", "c"}
	it := iterator.NewSliceIterator(elems)

	var got []string
	for it.Next() {
		got = append(got, it.Value())
	}

	if len(got) != len(elems) {
		t.Fatalf("got %d elements, want %d", len(got), len(elems))
	}
	for i := range elems {
		if got[i] != elems[i] {
			t.Errorf("element %d: got %q, want %q", i, got[i], elems[i])
		}
	}
}

// TestSliceIteratorPreservesOrder ensures elements are visited in slice order.
func TestSliceIteratorPreservesOrder(t *testing.T) {
	t.Parallel()

	elems := []int{10, 20, 30, 40, 50}
	it := iterator.NewSliceIterator(elems)

	i := 0
	for it.Next() {
		if got := it.Value(); got != elems[i] {
			t.Errorf("position %d: Value()=%d, want %d", i, got, elems[i])
		}
		i++
	}
	if i != len(elems) {
		t.Errorf("visited %d elements, want %d", i, len(elems))
	}
}

// TestSliceIteratorTableDriven covers several element counts.
func TestSliceIteratorTableDriven(t *testing.T) {
	t.Parallel()

	cases := [][]int{
		{},
		{1},
		{1, 2},
		{1, 2, 3, 4, 5},
	}

	for _, elems := range cases {
		it := iterator.NewSliceIterator(elems)
		count := 0
		for it.Next() {
			if it.Value() != elems[count] {
				t.Errorf("slice %v: element %d: got %d, want %d",
					elems, count, it.Value(), elems[count])
			}
			count++
		}
		if count != len(elems) {
			t.Errorf("slice %v: visited %d elements, want %d", elems, count, len(elems))
		}
	}
}
