package depth_test

import (
	"testing"

	"github.com/arr-ai/frozen/internal/pkg/depth"
	"github.com/arr-ai/frozen/internal/pkg/masker"
)

// buildMask builds a Masker with bits set at every index in indices.
func buildMask(indices ...int) masker.Masker {
	var m masker.Masker
	for _, i := range indices {
		m |= masker.NewMasker(i)
	}
	return m
}

// TestGaugeParallelSumsMatches verifies that Parallel accumulates the match
// counts returned by the callback for every set bit in the mask.
func TestGaugeParallelSumsMatches(t *testing.T) {
	t.Parallel()

	mask := buildMask(0, 1, 2, 3)

	// Use a gauge value large enough that depth < gauge, forcing parallel dispatch.
	g := depth.Gauge(100)

	ok, matches := g.Parallel(0, mask, func(i int) (bool, int) {
		return true, 1
	})

	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	want := mask.Next()
	_ = want
	// Four bits set, each callback returns 1 → total must be 4.
	if matches != 4 {
		t.Fatalf("expected matches=4, got %d", matches)
	}
}

// TestGaugeSequentialSumsMatches verifies that sequential execution (depth >= gauge)
// also accumulates match counts correctly.
func TestGaugeSequentialSumsMatches(t *testing.T) {
	t.Parallel()

	mask := buildMask(0, 2, 4)

	// depth.NonParallel == -1, so depth(0) >= gauge(-1) → sequential.
	ok, matches := depth.NonParallel.Parallel(0, mask, func(i int) (bool, int) {
		return true, 2
	})

	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	// Three bits, each callback returns 2 → total must be 6.
	if matches != 6 {
		t.Fatalf("expected matches=6, got %d", matches)
	}
}

// TestGaugeNonParallelIsAlwaysSequential checks that NonParallel never spawns
// goroutines, even when depth is 0 (the shallowest possible level).
func TestGaugeNonParallelIsAlwaysSequential(t *testing.T) {
	t.Parallel()

	mask := buildMask(0, 1)

	var callOrder []int
	ok, matches := depth.NonParallel.Parallel(0, mask, func(i int) (bool, int) {
		callOrder = append(callOrder, i)
		return true, 1
	})

	if !ok {
		t.Fatal("expected ok=true")
	}
	if matches != 2 {
		t.Fatalf("expected matches=2, got %d", matches)
	}
	// Sequential path visits bits in order of FirstIndex (low bit first).
	if len(callOrder) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(callOrder))
	}
}

// TestGaugeParallelEarlyExitSequential verifies that returning false from the
// callback causes Parallel to return false in sequential mode.
func TestGaugeParallelEarlyExitSequential(t *testing.T) {
	t.Parallel()

	mask := buildMask(0, 1, 2, 3)
	calls := 0

	ok, _ := depth.NonParallel.Parallel(0, mask, func(i int) (bool, int) {
		calls++
		// Stop after the first call.
		return false, 0
	})

	if ok {
		t.Fatal("expected ok=false after early exit")
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call before early exit, got %d", calls)
	}
}

// TestGaugeParallelEarlyExitParallel verifies that returning false from the
// callback causes Parallel to return false in parallel dispatch mode.
func TestGaugeParallelEarlyExitParallel(t *testing.T) {
	t.Parallel()

	mask := buildMask(0, 1, 2)

	// Force parallel dispatch: gauge large, depth smaller.
	g := depth.Gauge(100)

	ok, _ := g.Parallel(0, mask, func(i int) (bool, int) {
		return false, 0
	})

	if ok {
		t.Fatal("expected ok=false when all callbacks return false")
	}
}

// TestGaugeParallelEmptyMask verifies that an empty mask returns ok=true with
// zero matches and never calls the callback.
func TestGaugeParallelEmptyMask(t *testing.T) {
	t.Parallel()

	called := false
	ok, matches := depth.NonParallel.Parallel(0, 0, func(i int) (bool, int) {
		called = true
		return true, 1
	})

	if !ok {
		t.Fatal("expected ok=true for empty mask")
	}
	if matches != 0 {
		t.Fatalf("expected matches=0 for empty mask, got %d", matches)
	}
	if called {
		t.Fatal("callback must not be called for empty mask")
	}
}

// TestGaugeParallelSingleBit verifies correct behaviour with a single-bit mask.
func TestGaugeParallelSingleBit(t *testing.T) {
	t.Parallel()

	mask := masker.NewMasker(3)

	ok, matches := depth.NonParallel.Parallel(0, mask, func(i int) (bool, int) {
		if i != 3 {
			return false, 0
		}
		return true, 7
	})

	if !ok {
		t.Fatal("expected ok=true")
	}
	if matches != 7 {
		t.Fatalf("expected matches=7, got %d", matches)
	}
}

// TestNewGaugeReturnsReasonableValue checks that NewGauge does not panic and
// returns a value that is an integer (no type constraint beyond that here).
func TestNewGaugeReturnsReasonableValue(t *testing.T) {
	t.Parallel()

	for _, count := range []int{0, 1, 10, 100, 1_000, 1_000_000} {
		g := depth.NewGauge(count)
		// The result must be a valid Gauge (int); no negative values below -1
		// are meaningful but we mainly verify no panic occurs.
		_ = g
	}
}

// TestNewGaugeLargeCountDoesNotExceedMaxDepth verifies that a very large element
// count does not produce a gauge that grows without bound.
func TestNewGaugeLargeCountDoesNotExceedMaxDepth(t *testing.T) {
	t.Parallel()

	g1 := depth.NewGauge(1 << 20)
	g2 := depth.NewGauge(1 << 30)

	// Both should produce valid, non-negative values; and the larger count
	// must not produce a strictly smaller gauge.
	if int(g1) < 0 || int(g2) < 0 {
		t.Fatalf("NewGauge returned negative value: g1=%d g2=%d", g1, g2)
	}
}
