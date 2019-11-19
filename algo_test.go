package frozen

import "testing"

func TestAlgo(t *testing.T) {
	t.Parallel()

	ca1 := EmptyMap().With("c", 1).With("a", 10)
	ca2 := EmptyMap().With("c", 1).With("a", 11)

	ca := EmptySet().
		With(ca1).
		With(ca2).
		With(EmptyMap().With("c", 2).With("a", 13)).
		With(EmptyMap().With("c", 3).With("a", 11)).
		With(EmptyMap().With("c", 4).With("a", 14)).
		With(EmptyMap().With("c", 3).With("a", 10)).
		With(EmptyMap().With("c", 4).With("a", 13))
	t.Log(ca.Nest("aa", "a").Nest("cc", "c"))
}
