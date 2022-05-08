package tree

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	t.Parallel()

	p := &packer[int]{}
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.WithChild(i, nil), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.WithChild(i, newLeaf1(1)).WithChild(i, nil)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := (&packer[int]{}).
		WithChild(1, newLeaf2(1, 2)).
		WithChild(3, newLeaf2(10, 20)).
		WithChild(3, nil).
		WithChild(5, newLeaf2(3, 4))
	q := (&packer[int]{}).
		WithChild(1, newLeaf2(1, 2)).
		WithChild(3, nil).
		WithChild(5, newLeaf2(3, 4))
	assertEqualPacked(t, p, q)
}

func assertEqualPacked[T any](t *testing.T, expected, actual *packer[T], msgAndArgs ...any) bool {
	t.Helper()

	if !expected.EqualPacker(actual) {
		expected.EqualPacker(actual)
		assert.Fail(t, fmt.Sprintf("packed unequal\nexpected: %v, actual:   %v",
			&branch[T]{p: *expected}, &branch[T]{p: *actual}), msgAndArgs...)
		return false
	}
	return true
}
