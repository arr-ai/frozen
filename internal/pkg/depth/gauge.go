package depth

import (
	"math/bits"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/arr-ai/frozen/internal/pkg/masker"
)

const (
	// Fanout determines the number of children each branch will have.
	Fanout = 1 << FanoutBits
)

var (
	maxConcurrency = func() int {
		frozenConcurrency := os.Getenv("FROZEN_CONCURRENCY")
		var ret int
		var err error
		if strings.ToLower(frozenConcurrency) == "off" {
			ret = 1<<(bits.UintSize-1) - 1
		} else {
			ret, err = strconv.Atoi(frozenConcurrency)
			if err != nil {
				ret = 15
			}
		}
		return ret
	}()

	// Ensure parallelism is enough to keep all cores busy.
	maxDepth = (bits.Len64(uint64(runtime.GOMAXPROCS(0)-1))-1)/FanoutBits + 1

	// NonParallel is a Gauge that never triggers parallel behaviour.
	NonParallel Gauge = -1
)

type Gauge int

func NewGauge(count int) Gauge {
	g := (bits.Len64(uint64(count)) - maxConcurrency) / FanoutBits
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
