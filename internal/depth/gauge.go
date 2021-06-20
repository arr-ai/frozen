package depth

import (
	"math/bits"
	"os"
	"strconv"
	"strings"
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

	NonParallel Gauge = -1
)

type Gauge int

func NewGauge(count int) Gauge {
	return Gauge((bits.Len64(uint64(count)) - maxConcurrency) / 3)
}

func (pd Gauge) Parallel(depth int) bool {
	return depth < int(pd)
}
