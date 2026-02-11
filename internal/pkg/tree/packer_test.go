package tree

import (
	"fmt"
	"testing"

	"github.com/arr-ai/frozen/internal/pkg/test"
)

func TestPackedWith(t *testing.T) {
	t.Parallel()

	p := &packer[int]{}
	for i := 0; i < fanout; i++ {
		assertEqualPacked(t, p, p.WithChild(i, nil), i)
	}
	for i := 0; i < fanout; i++ {
		q := p.WithChild(i, newLeaf1(1)).WithChild(i, nil)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := (&packer[int]{}).
		WithChild(1, newLeaf1(1)).
		WithChild(3, newLeaf1(10)).
		WithChild(3, nil).
		WithChild(5, newLeaf1(3))
	q := (&packer[int]{}).
		WithChild(1, newLeaf1(1)).
		WithChild(3, nil).
		WithChild(5, newLeaf1(3))
	assertEqualPacked(t, p, q)
}

func assertEqualPacked[T any](t *testing.T, expected, actual *packer[T], msgAndArgs ...any) bool {
	t.Helper()

	if !expected.EqualPacker(actual) {
		expected.EqualPacker(actual)
		test.Fail(t, append(
			[]any{
				fmt.Sprintf(
					"packed unequal\nexpected: %v, actual:   %v",
					&branch[T]{p: *expected}, &branch[T]{p: *actual},
				),
			},
			msgAndArgs...)...)
		return false
	}
	return true
}
