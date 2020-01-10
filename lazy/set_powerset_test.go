package lazy

import (
	"testing"

	"github.com/arr-ai/frozen"
)

func TestSetPowerset(t *testing.T) {
	t.Parallel()

	f := frozen.NewSet(1, 2, 3)
	s := func() Set { return Frozen(f).Powerset() }

	assertSetOps(t, f.Powerset().Map(func(el interface{}) interface{} { return Frozen(el.(frozen.Set)) }), s())

	assertFastNotIsEmpty(t, s())
	assertFastCountEqual(t, 8, s())
	assertFastCountUpToEqual(t, 7, s(), 7)
	assertFastCountUpToEqual(t, 8, s(), 8)
	assertFastCountUpToEqual(t, 8, s(), 9)
}
