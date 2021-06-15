package frozen

import (
	"math/bits"
	"os"
	"strconv"
	"strings"
)

var (
	maxConcurrency = func() int {
		frozenConcurrency := os.Getenv("FROZEN_CONCURRENCY")
		var maxConcurrency int
		var err error
		if strings.ToLower(frozenConcurrency) == "off" {
			maxConcurrency = 1<<(bits.UintSize-1) - 1
		} else {
			maxConcurrency, err = strconv.Atoi(frozenConcurrency)
			if err != nil {
				maxConcurrency = 15
			}
		}
		return maxConcurrency
	}()

	nonParallel parallelDepthGauge = -1
)

type parallelDepthGauge int

func newParallelDepthGauge(count int) parallelDepthGauge {
	return parallelDepthGauge((bits.Len64(uint64(count)) - maxConcurrency) / 3)
}

func (pd parallelDepthGauge) parallel(depth int) bool {
	return depth < int(pd)
}
