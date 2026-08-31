package depth

import (
	"math/bits"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/arr-ai/frozen/internal/pkg/masker"
)

const (
	// Fanout determines the number of children each branch will have.
	Fanout = 1 << FanoutBits
)

var (
	// parallelShift is bits.Len64(minimum chunk size): an operation fans out
	// one more level while each goroutine would still get at least that many
	// elements. Held atomically because SetMinChunk may be called while
	// operations are running.
	parallelShift atomic.Int64

	// Ensure parallelism is enough to keep all cores busy.
	maxDepth = (bits.Len64(uint64(runtime.GOMAXPROCS(0)-1))-1)/FanoutBits + 1

	// NonParallel is a Gauge that never triggers parallel behaviour.
	NonParallel Gauge = -1
)

// DefaultMinChunk is the fewest elements an operation will hand to a single
// goroutine. Below it, the cost of starting a goroutine and collecting its
// result outweighs the work, and fanning out further makes an operation
// slower rather than faster.
const DefaultMinChunk = 256

// shiftDisabled is a shift large enough that no element count can reach the
// first level of fan-out.
const shiftDisabled = 1<<(bits.UintSize-1) - 1

// initialShift resolves the starting shift from the environment.
var _ = func() struct{} {
	parallelShift.Store(int64(initialShift()))
	return struct{}{}
}()

func initialShift() int {
	shift := bits.Len64(DefaultMinChunk)
	// FROZEN_CONCURRENCY historically named this shift directly rather than
	// a chunk size, and "off" disables parallelism. Both still work.
	if env := os.Getenv("FROZEN_CONCURRENCY"); env != "" {
		if strings.EqualFold(env, "off") {
			shift = shiftDisabled
		} else if n, err := strconv.Atoi(env); err == nil {
			shift = n
		}
	}
	return shift
}

// MinChunk returns the current minimum chunk size, or 0 if parallelism is
// disabled.
func MinChunk() int {
	shift := parallelShift.Load()
	if shift >= shiftDisabled {
		return 0
	}
	return 1 << (shift - 1)
}

// SetMinChunk sets the minimum chunk size and returns the previous one. A
// value below 1 disables parallelism.
func SetMinChunk(n int) int {
	previous := MinChunk()
	if n < 1 {
		parallelShift.Store(shiftDisabled)
	} else {
		parallelShift.Store(int64(bits.Len64(uint64(n)))) //nolint:gosec // n >= 1
	}
	return previous
}

type Gauge int

func NewGauge(count int) Gauge {
	shift := int(parallelShift.Load())
	if shift >= shiftDisabled {
		return NonParallel
	}
	g := (bits.Len64(uint64(count)) - shift) / FanoutBits //nolint:gosec // count is a non-negative element count
	if g > maxDepth {
		g = maxDepth
	}
	return Gauge(g)
}

func (pd Gauge) Parallel(depth int, mask masker.Masker, f func(i int) (bool, int)) (_ bool, matches int) {
	if depth < int(pd) {
		type outcome struct {
			matches int
			ok      bool
		}
		outcomes := make(chan outcome, Fanout)
		for m := mask; m != 0; m = m.Next() {
			i := m.FirstIndex()
			go func() {
				ok, m := f(i)
				outcomes <- outcome{matches: m, ok: ok}
			}()
		}
		for m := mask; m != 0; m = m.Next() {
			if o := <-outcomes; o.ok {
				matches += o.matches
			} else {
				return false, matches
			}
		}
	} else {
		for m := mask; m != 0; m = m.Next() {
			i := m.FirstIndex()
			ok, m := f(i)
			if !ok {
				return false, matches
			}
			matches += m
		}
	}

	return true, matches
}
