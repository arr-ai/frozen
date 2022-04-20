package lazy_test

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/arr-ai/frozen"
	. "github.com/arr-ai/frozen/lazy"
)

type eagerLazyPair struct {
	index int
	eager Set
	lazy  Set
}

type eagerLazySlice []eagerLazyPair

func (s eagerLazySlice) Len() int {
	return len(s)
}

func (s eagerLazySlice) Less(i, j int) bool {
	return s[i].index < s[j].index
}

func (s eagerLazySlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func TestSet(t *testing.T) { //nolint:funlen,cyclop
	t.Parallel()

	if testing.Short() {
		return
	}

	type work struct {
		line  int
		index int
		eager func() Set
		lazy  Set
	}
	pairsCh := make(chan eagerLazySlice)
	pairCh := make(chan eagerLazyPair)
	workCh := make(chan work)
	var added, done uint64 = 0, 0
	var wg sync.WaitGroup
	go func() {
		pairs := eagerLazySlice{}
		for {
			select {
			case pairsCh <- pairs:
			case p, ok := <-pairCh:
				if !ok {
					return
				}
				pairs = append(pairs, p)
				atomic.AddUint64(&done, 1)
				wg.Done()
			}
		}
	}()
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for w := range workCh {
				f := w.eager()
				assertEqualSet(t, f, w.lazy, "line=%v", w.line)
				pairCh <- eagerLazyPair{index: w.index, eager: f, lazy: w.lazy}
			}
		}()
	}
	workIndex := 0
	test := func(eager func() Set, lazy Set) {
		atomic.AddUint64(&added, 1)
		wg.Add(1)
		_, _, line, _ := runtime.Caller(1)
		workCh <- work{
			line:  line,
			index: workIndex,
			eager: eager,
			lazy:  lazy,
		}
		workIndex++
	}

	for i := uint64(0); i < 1<<3; i++ {
		f := Frozen(frozen.SetAs[any](frozen.NewSetFromMask64(i)))
		test(func() Set { return f }, f)
	}
	for i := 0; i < 2; i++ {
		wg.Wait()
		pairs := <-pairsCh
		sort.Sort(pairs)
		for _, p := range pairs {
			p := p
			for _, pred := range []func(any) bool{
				func(el any) bool { return false },
				func(el any) bool { return true },
				func(el any) bool { return extractInt(el) < 2 },
				func(el any) bool { return extractInt(el)%2 == 0 },
			} {
				pred := pred
				test(func() Set { return p.eager.Where(pred) }, p.lazy.Where(pred))
			}
			for _, m := range []func(any) any{
				func(el any) any { return 42 },
				func(el any) any { return extractInt(el) * 2 },
				func(el any) any { return extractInt(el) / 2 },
				func(el any) any { return extractInt(el) % 2 },
			} {
				m := m
				test(func() Set { return p.eager.Map(m) }, p.lazy.Map(m))
			}
			for _, q := range pairs {
				q := q
				test(func() Set { return p.eager.Intersection(q.eager) }, p.lazy.Intersection(q.lazy))
				test(func() Set { return p.eager.Union(q.eager) }, p.lazy.Union(q.lazy))
				test(func() Set { return p.eager.Difference(q.eager) }, p.lazy.Difference(q.lazy))
				test(func() Set { return p.eager.SymmetricDifference(q.eager) }, p.lazy.SymmetricDifference(q.lazy))
			}
			test(func() Set { return p.eager.Powerset() }, p.lazy.Powerset())
		}
	}
	wg.Wait()
	close(pairCh)
	close(workCh)
}

func ConvertSlice[T, U any](slice []T) []U {
	result := make([]U, len(slice))
	for _, t := range slice {
		var a any = t
		result = append(result, a.(U))
	}
	return result
}
