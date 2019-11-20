package frozen

import "testing"

func TestNest(t *testing.T) {
	t.Parallel()

	ca := EmptySet().
		With(NewMap(KV("c", 1), KV("a", 10))).
		With(NewMap(KV("c", 1), KV("a", 11))).
		With(NewMap(KV("c", 2), KV("a", 13))).
		With(NewMap(KV("c", 3), KV("a", 11))).
		With(NewMap(KV("c", 4), KV("a", 14))).
		With(NewMap(KV("c", 3), KV("a", 10))).
		With(NewMap(KV("c", 4), KV("a", 13)))
	t.Log(ca)
	caa := ca.Nest("aa", "a")
	t.Log(caa)
	aacc := caa.Nest("cc", "c")
	t.Log(aacc)
}
