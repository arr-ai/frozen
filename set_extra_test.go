package frozen_test

import (
	"encoding/json"
	"testing"

	. "github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
	testset "github.com/arr-ai/frozen/internal/pkg/test/set"
)

func TestIota(t *testing.T) {
	t.Parallel()

	testset.AssertSetEqual(t, Set[int]{}, Iota(0))
	testset.AssertSetEqual(t, NewSet(0), Iota(1))
	testset.AssertSetEqual(t, NewSet(0, 1, 2, 3, 4, 5), Iota(6))
}

func TestIota2(t *testing.T) {
	t.Parallel()

	testset.AssertSetEqual(t, Set[int]{}, Iota2(6, 6))
	testset.AssertSetEqual(t, NewSet(5), Iota2(5, 6))
	testset.AssertSetEqual(t, NewSet(1, 2, 3, 4, 5), Iota2(1, 6))
}

func TestIota3(t *testing.T) {
	t.Parallel()

	testset.AssertSetEqual(t, Set[int]{}, Iota3(1, 1, 0))

	test.Panic(t, func() { Iota3(1, 2, 0) })

	testset.AssertSetEqual(t, NewSet(1, 3, 5), Iota3(1, 6, 2))
	testset.AssertSetEqual(t, NewSet(1, 3, 5), Iota3(1, 7, 2))
	testset.AssertSetEqual(t, NewSet(1, 3, 5), Iota3(5, 0, -2))
	testset.AssertSetEqual(t, NewSet(1, 3, 5), Iota3(5, -1, -2))
}

func TestNewSetFromMask64(t *testing.T) {
	t.Parallel()

	testset.AssertSetEqual(t, Set[int]{}, NewSetFromMask64(0))
	for i := 0; i < 64; i++ {
		testset.AssertSetEqual(t, NewSet(i), NewSetFromMask64(uint64(1)<<uint(i)), "%v", i)
	}
	for i := 0; i < 64; i++ {
		testset.AssertSetEqual(t, Iota(i), NewSetFromMask64(uint64(1)<<uint(i)-1), "%v", i)
	}
}

func TestSetMarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(Iota3(0, 10, 3))
	if test.NoError(t, err) {
		var s []float64
		test.RequireNoError(t, json.Unmarshal(j, &s))
		test.ElementsMatch(t, []float64{0, 3, 6, 9}, s)
	}
}
