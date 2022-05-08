package frozen_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/arr-ai/frozen"
)

func TestSetOfSet(t *testing.T) {
	t.Parallel()

	s := NewSet(NewSet(10), NewSet(11))
	j, err := json.Marshal(s)
	require.NoError(t, err)
	var s2 []any
	require.NoError(t, json.Unmarshal(j, &s2))
	assert.ElementsMatch(t,
		[]any{
			[]any{10.0},
			[]any{11.0},
		},
		s2,
	)
}

func TestSetOfMap(t *testing.T) {
	t.Parallel()

	s := NewSet(NewMap(KV("a", 10)), NewMap(KV("a", 11)))
	j, err := json.Marshal(s)
	require.NoError(t, err)
	var s2 []any
	require.NoError(t, json.Unmarshal(j, &s2))
	assert.ElementsMatch(t,
		[]any{
			map[string]any{"a": 10.0},
			map[string]any{"a": 11.0},
		},
		s2,
	)
}
