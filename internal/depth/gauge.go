package depth

import (
	"math/bits"
	"os"
	"strconv"
	"strings"
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

func (pd Gauge) Parallel(depth int, matches *int, f func(i int, matches *int) bool) bool {
	totalMatches := 0

	if depth < int(pd) {
		type outcome struct {
			matches int
			ok      bool
		}
		outcomes := make(chan outcome, Fanout)
		for i := 0; i < Fanout; i++ {
			i := i
			go func() {
				var matches int
				ok := f(i, &matches)
				outcomes <- outcome{matches: matches, ok: ok}
			}()
		}
		for i := 0; i < Fanout; i++ {
			if o := <-outcomes; o.ok {
				totalMatches += o.matches
			} else {
				return false
			}
		}
	} else {
		for i := 0; i < Fanout; i++ {
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
