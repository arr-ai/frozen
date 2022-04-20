package lazy_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	. "github.com/arr-ai/frozen/lazy"
)

func TestSetPowerset(t *testing.T) {
	t.Parallel()

	f := frozen.NewSet[any](1, 2, 3)
	s := func() Set { return Frozen(f).Powerset() }

	assertSetOps(t,
		frozen.SetMap(
			frozen.Powerset(f),
			func(el frozen.Set[any]) any { return Frozen(el) },
		),
		s())

	assertFastNotIsEmpty(t, s())
	assertFastCountEqual(t, 8, s())
	assertFastCountUpToEqual(t, 7, s(), 7)
	assertFastCountUpToEqual(t, 8, s(), 8)
	assertFastCountUpToEqual(t, 8, s(), 9)
}
