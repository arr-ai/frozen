package tree_test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arr-ai/frozen/v2/internal/pkg/test"
	"github.com/arr-ai/frozen/v2/internal/pkg/tree"
)

func TestBranchRemove(t *testing.T) {
	t.Parallel()

	const N = 1 << 15

	test.Replayable(true, func(r *test.Replayer) {
		var b tree.Builder[int]
		has := func(i int) bool {
			return b.Get(tree.DefaultNPEqArgs[int](), i) != nil
		}
		for i := 0; i < N; i++ {
			require.False(t, has(i), i)
			b.Add(tree.DefaultNPCombineArgs[int](), i)
			require.True(t, has(i), i)
		}

		for i := 0; i < N; i++ {
			require.True(t, has(i), i)
			m := r.Mark(i)
			if m.IsTarget() {
				log.Printf("%+v", b)
			}
			b.Remove(tree.DefaultNPEqArgs[int](), i)
			if !assert.False(t, has(i), i) {
				log.Printf("%+v", b)
				r.ReplayTo(m)
				return
			}
		}
	})
}
