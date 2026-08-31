package frozen

import "github.com/arr-ai/frozen/internal/pkg/depth"

// Operations over a large enough collection fan out across goroutines: each
// level of fan-out splits the work eight ways, and a level is added only
// while every goroutine would still receive at least MinParallelChunk
// elements. Below that, starting a goroutine and collecting its result costs
// more than the work it is given.
//
// The default suits collections of a few thousand elements upwards. Raise it
// when each element is cheap to process, so that fan-out has to earn its
// keep over more of them; lower it when a caller's own per-element work is
// expensive, which is what actually pays for the goroutine.

// NoParallel disables auto-parallelism when passed to SetMinParallelChunk:
// every operation then runs on the calling goroutine.
const NoParallel = 0

// DisableParallel turns auto-parallelism off entirely and returns the
// previous minimum chunk size, so a caller can restore it:
//
//	defer frozen.SetMinParallelChunk(frozen.DisableParallel())
//
// Reach for it when goroutines are unwelcome rather than merely unhelpful —
// inside a benchmark that is measuring something else, or when the elements'
// own methods are not safe to run concurrently.
func DisableParallel() int {
	return SetMinParallelChunk(NoParallel)
}

// ParallelEnabled reports whether operations may fan out across goroutines.
func ParallelEnabled() bool {
	return MinParallelChunk() > 0
}

// MinParallelChunk returns the fewest elements an operation will hand to one
// goroutine, or NoParallel if parallelism is disabled.
func MinParallelChunk() int {
	return depth.MinChunk()
}

// SetMinParallelChunk sets the fewest elements an operation will hand to one
// goroutine, and returns the previous setting. NoParallel, or any value
// below 1, disables parallelism entirely; see DisableParallel.
//
// It applies process-wide and takes effect for operations started after it
// returns. It is safe to call while other operations are running, and the
// results of those operations are unaffected: parallelism changes only how
// the work is scheduled.
//
// The initial value comes from DefaultMinParallelChunk, or from the
// FROZEN_CONCURRENCY environment variable when that is set.
func SetMinParallelChunk(n int) int {
	return depth.SetMinChunk(n)
}

// DefaultMinParallelChunk is the value used when nothing overrides it.
const DefaultMinParallelChunk = depth.DefaultMinChunk
