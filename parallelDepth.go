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

	nonParallel parallelDepthGauge = -1
)

type parallelDepthGauge int

func newParallelDepthGauge(count int) parallelDepthGauge {
	return parallelDepthGauge((bits.Len64(uint64(count)) - maxConcurrency) / 3)
}

func (pd parallelDepthGauge) parallel(depth int) bool {
	return depth < int(pd)
}
