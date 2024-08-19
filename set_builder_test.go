package frozen_test

import (
	"log"
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
	testset "github.com/arr-ai/frozen/internal/pkg/test/set"
)

func TestSetBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b frozen.SetBuilder[int]
	testset.AssertSetEqual(t, frozen.Set[int]{}, b.Finish())
}

func TestSetBuilder(t *testing.T) {
	t.Parallel()

	var b frozen.SetBuilder[int]
	for i := 0; i < 10; i++ {
		b.Add(i)
	}
	m := b.Finish()
	test.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		test.True(t, m.Has(i))
	}
}

func TestSetBuilderIncremental(t *testing.T) {
	t.Parallel()

	test.Replayable(false, func(*test.Replayer) {
		N := 1_000
		if testing.Short() {
			N /= 10
		}
		arr := make([]int, 0, N)
		for i := 0; i < N; i++ {
			arr = append(arr, i)
		}

		for i := N - 1; i >= 0; i-- {
			i := i
			corpus := arr[i:]
			assertSameElements(t, corpus, frozen.NewSet(arr[i:]...).Elements())
		}
	})
}

func TestSetBuilderRemove(t *testing.T) {
	t.Parallel()

	test.Replayable(true, func(*test.Replayer) {
		var b frozen.SetBuilder[int]
		for i := 0; i < 15; i++ {
			b.Add(i)
		}
		for i := 5; i < 10; i++ {
			b.Remove(i)
		}
		m := b.Finish()

		if !test.Equal(t, 10, m.Count()) {
			log.Print(m)
		}
		for i := 0; i < 15; i++ {
			switch {
			case i < 5:
				assertSetHas(t, m, i)
			case i < 10:
				assertSetNotHas(t, m, i)
			default:
				assertSetHas(t, m, i)
			}
		}
	})
}

func TestSetBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:cyclop,funlen,gocognit
	t.Parallel()

	test.Replayable(false, func(r *test.Replayer) {
		var b frozen.SetBuilder[int]

		s := uint64(0)

		assertMatch := func(format string, args ...any) bool {
			for j := 0; j < 60; j++ {
				if !test.Equal(t, s&(uint64(1)<<uint(j)) != 0, b.Has(j),
					append(append([]any{format + " j=%v"}, args...), j)...) {
					return false
				}
			}
			return true
		}

		add := func(i int) {
			b.Add(i)
			s |= uint64(1) << uint(i)
		}

		remove := func(i int) {
			b.Remove(i)
			s &^= uint64(1) << uint(i)
		}

		assertMatch("")
		for i := 0; i < 60; i++ {
			add(i)
			if !assertMatch("i=%v", i) {
				return
			}
		}
		mark := r.Mark()
		for i := 20; i < 50; i++ {
			if mark.IsTarget() {
				log.Printf("%+v", b)
			}
			remove(i)
			if !assertMatch("i=%v", i) {
				return
			}
		}
		if mark.IsTarget() {
			log.Printf("%+v", b)
		}

		for i := 10; i < 30; i++ {
			add(i)
			if !assertMatch("i=%v", i) {
				r.ReplayTo(mark)
			}
		}
		for i := 40; i < 55; i++ {
			remove(i)
			if !assertMatch("i=%v", i) {
				return
			}
		}
		m := b.Finish()

		for i := 0; i < 60; i++ {
			switch {
			case i < 30:
				assertSetHas(t, m, i)
			case i < 55:
				assertSetNotHas(t, m, i)
			default:
				assertSetHas(t, m, i)
			}
		}
	})
}
