package lazy

import (
	"testing"

	"github.com/marcelocantos/frozen"
)

func TestSetEmpty(t *testing.T) {
	t.Parallel()

	s := emptySet{}

	assertSetOps(t, frozen.Set{}, s)

	assertFastIsEmpty(t, s)
	assertFastCountEqual(t, 0, s)
	assertFastCountUpToEqual(t, 0, s, 0)
	assertFastCountUpToEqual(t, 0, s, 1)
	assertFastNotHas(t, s, 3)
	assertFastIsEmpty(t, s.Where(func(_ interface{}) bool { return true }))
	assertFastNotIsEmpty(t, s.With(2))
	assertFastIsEmpty(t, s.Without(2))
	assertFastIsEmpty(t, s.With(2).Without(2))
	assertFastIsEmpty(t, s.Map(func(_ interface{}) interface{} { return 42 }))
	assertFastIsEmpty(t, s.Intersection(Frozen(frozen.NewSet(1, 2, 3))))
	assertFastIsEmpty(t, s.Intersection(Frozen(frozen.NewSet(1, 2, 3))))
	assertFastNotIsEmpty(t, s.Powerset())
}
