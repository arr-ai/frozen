package tree_test

import (
	"log"
	"testing"

	"github.com/arr-ai/frozen/internal/pkg/test"
	"github.com/arr-ai/frozen/internal/pkg/tree"
)

func TestBranchRemove(t *testing.T) {
	t.Parallel()

	const N = 1 << 15

	test.Replayable(true, func(r *test.Replayer) {
		var b tree.Builder[int]
		has := func(i int) bool {
			return b.Get(i) != nil
		}
		for i := 0; i < N; i++ {
			test.RequireFalse(t, has(i), i)
			b.Add(i)
			test.RequireTrue(t, has(i), i)
		}

		for i := 0; i < N; i++ {
			test.RequireTrue(t, has(i), i)
			m := r.Mark(i)
			if m.IsTarget() {
				log.Printf("%+v", b)
			}
			b.Remove(i)
			if !test.False(t, has(i), i) {
				log.Printf("%+v", b)
				r.ReplayTo(m)
				return
			}
		}
	})
}
