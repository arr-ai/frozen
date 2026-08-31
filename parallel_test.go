package frozen_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

// Not parallel: these set a process-wide knob.
func TestMinParallelChunk(t *testing.T) { //nolint:paralleltest
	previous := frozen.MinParallelChunk()
	defer frozen.SetMinParallelChunk(previous)

	test.Equal(t, frozen.DefaultMinParallelChunk, previous, "default is in force")

	was := frozen.SetMinParallelChunk(1024)
	test.Equal(t, previous, was, "the setter returns the previous value")
	test.Equal(t, 1024, frozen.MinParallelChunk())

	// Whatever the setting, results are unchanged: it schedules work, it
	// does not change it.
	build := func() frozen.Set[int] {
		var b frozen.SetBuilder[int]
		for i := 0; i < 20000; i++ {
			b.Add(i)
		}
		return b.Finish()
	}
	frozen.SetMinParallelChunk(1)
	fine := build()
	test.True(t, frozen.ParallelEnabled())
	restore := frozen.DisableParallel()
	test.Equal(t, 1, restore, "DisableParallel returns the previous setting")
	test.False(t, frozen.ParallelEnabled(), "DisableParallel turns fan-out off")
	test.Equal(t, frozen.NoParallel, frozen.MinParallelChunk())
	sequential := build()

	test.True(t, fine.Equal(sequential), "the same elements either way")
	test.Equal(t, sequential.Count(), fine.Count())
	evens := func(s frozen.Set[int]) int { return s.Where(func(i int) bool { return i%2 == 0 }).Count() }
	test.Equal(t, evens(sequential), evens(fine), "Where agrees either way")
}

// Combine backs Union and Update, and runs on several goroutines once a
// tree fans out. Its arguments carry a flipped counterpart that used to be
// assigned on first use, which raced between them. The chunk size is forced
// down so that a modest collection is enough to fan out; with the default,
// this needs 2^17 elements to reach the same code.
func TestCombineIsSafeWhenParallel(t *testing.T) { //nolint:paralleltest
	defer frozen.SetMinParallelChunk(frozen.SetMinParallelChunk(1))

	build := func(from, to int) frozen.Set[int] {
		var b frozen.SetBuilder[int]
		for i := from; i < to; i++ {
			b.Add(i)
		}
		return b.Finish()
	}
	a, c := build(0, 5000), build(2500, 7500)

	union := a.Union(c)
	test.Equal(t, 7500, union.Count())
	test.True(t, union.Has(0) && union.Has(7499) && union.Has(3000))

	var m1, m2 frozen.MapBuilder[int, int]
	for i := 0; i < 5000; i++ {
		m1.Put(i, i)
		m2.Put(i+2500, -i)
	}
	updated := m1.Finish().Update(m2.Finish())
	test.Equal(t, 7500, updated.Count())
	v, has := updated.Get(3000)
	test.True(t, has)
	test.Equal(t, -500, v, "the right-hand map wins a clash")
}
