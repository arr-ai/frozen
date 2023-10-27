package frozen_test

import (
	"encoding/json"
	"testing"

	. "github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

func TestMapMarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(NewMap(KV("a", 2), KV("b", 4), KV("c", 2)))
	if test.NoError(t, err) {
		var s map[string]float64
		test.RequireNoError(t, json.Unmarshal(j, &s))
		test.Equal(t, map[string]float64{"a": 2, "b": 4, "c": 2}, s)
	}

	j, err = json.Marshal(NewMap(KV(1, 2), KV(3, 4), KV(4, 2)))
	if test.NoError(t, err) {
		var s [][]int
		test.RequireNoError(t, json.Unmarshal(j, &s))
		test.ElementsMatch(t, [][]int{{1, 2}, {3, 4}, {4, 2}}, s)
	}
}
