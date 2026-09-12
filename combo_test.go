package frozen_test

import (
	"encoding/json"
	"testing"

	"github.com/arr-ai/frozen/v2"
	"github.com/arr-ai/frozen/v2/internal/pkg/test"
)

func TestSetOfSet(t *testing.T) {
	t.Parallel()

	s := frozen.NewSet(frozen.NewSet(10), frozen.NewSet(11))
	j, err := json.Marshal(s)
	test.RequireNoError(t, err)
	var s2 []any
	test.RequireNoError(t, json.Unmarshal(j, &s2))
	test.ElementsMatch(t,
		[]any{
			[]any{10.0},
			[]any{11.0},
		},
		s2,
	)
}

func TestSetOfMap(t *testing.T) {
	t.Parallel()

	s := frozen.NewSet(frozen.NewMap(frozen.KV("a", 10)), frozen.NewMap(frozen.KV("a", 11)))
	j, err := json.Marshal(s)
	test.RequireNoError(t, err)
	var s2 []any
	test.RequireNoError(t, json.Unmarshal(j, &s2))
	test.ElementsMatch(t,
		[]any{
			map[string]any{"a": 10.0},
			map[string]any{"a": 11.0},
		},
		s2,
	)
}
