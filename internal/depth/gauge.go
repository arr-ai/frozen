package depth

import (
	"math/bits"
	"os"
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

	// NonParallel is a Gauge that never triggers parallel behaviour.
	NonParallel Gauge = -1
)

type Gauge int

func NewGauge(count int) Gauge {
	return Gauge((bits.Len64(uint64(count)) - maxConcurrency) / 3)
}

func (pd Gauge) Parallel(depth int, mask masker.Masker, matches *int, f func(i int, matches *int) bool) bool {
	totalMatches := 0

	if depth < int(pd) {
		type outcome struct {
			matches int
			ok      bool
		}
		outcomes := make(chan outcome, Fanout)
		for m := mask; m != 0; m = m.Next() {
			i := m.FirstIndex()
			go func() {
				var matches int
				ok := f(i, &matches)
				outcomes <- outcome{matches: matches, ok: ok}
			}()
		}
		for m := mask; m != 0; m = m.Next() {
			if o := <-outcomes; o.ok {
				totalMatches += o.matches
			} else {
				return false
			}
		}
	} else {
		for m := mask; m != 0; m = m.Next() {
			i := m.FirstIndex()
			if !f(i, &totalMatches) {
				return false
			}
		}
	}

	if matches != nil {
		*matches += totalMatches
	}
	return true
}
