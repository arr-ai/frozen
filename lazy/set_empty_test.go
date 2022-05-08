package lazy_test

import (
	"testing"

	"github.com/arr-ai/frozen"
	. "github.com/arr-ai/frozen/lazy"
)

func TestSetEmpty(t *testing.T) {
	t.Parallel()

	s := EmptySet{}

	assertSetOps(t, frozen.Set[any]{}, s)

	assertFastIsEmpty(t, s)
	assertFastCountEqual(t, 0, s)
	assertFastCountUpToEqual(t, 0, s, 0)
	assertFastCountUpToEqual(t, 0, s, 1)
	assertFastNotHas(t, s, 3)
	assertFastIsEmpty(t, s.Where(func(_ any) bool { return true }))
	assertFastNotIsEmpty(t, s.With(2))
	assertFastIsEmpty(t, s.Without(2))
	assertFastIsEmpty(t, s.With(2).Without(2))
	assertFastIsEmpty(t, s.Map(func(_ any) any { return 42 }))
	assertFastIsEmpty(t, s.Intersection(Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastIsEmpty(t, s.Intersection(Frozen(frozen.NewSet[any](1, 2, 3))))
	assertFastNotIsEmpty(t, s.Powerset())
}
