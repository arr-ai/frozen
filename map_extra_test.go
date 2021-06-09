package frozen_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/arr-ai/frozen"
)

func TestMapMarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(NewMap(KV("a", 2), KV("b", 4), KV("c", 2)))
	if assert.NoError(t, err) {
		var s map[string]float64
		require.NoError(t, json.Unmarshal(j, &s))
		assert.Equal(t, map[string]float64{"a": 2, "b": 4, "c": 2}, s)
	}

	j, err = json.Marshal(NewMap(KV(1, 2), KV(3, 4), KV(4, 2)))
	if assert.NoError(t, err) {
		t.Log(string(j))
		var s [][]int
		require.NoError(t, json.Unmarshal(j, &s))
		assert.ElementsMatch(t, [][]int{{1, 2}, {3, 4}, {4, 2}}, s)
	}
}
