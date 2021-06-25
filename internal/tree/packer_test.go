package tree

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	t.Parallel()

	p := &packer{}
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.With(i, theEmptyNode), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.With(i, newLeaf(1).Node()).With(i, theEmptyNode)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := (&packer{}).
		With(1, newLeaf(1, 2).Node()).
		With(3, newLeaf(10, 20).Node()).
		With(3, theEmptyNode).
		With(5, newLeaf(3, 4).Node())
	q := (&packer{}).
		With(1, newLeaf(1, 2).Node()).
		With(3, theEmptyNode).
		With(5, newLeaf(3, 4).Node())
	assertEqualPacked(t, p, q)
}

//nolint:unparam
func assertEqualPacked(t *testing.T, expected, actual *packer, msgAndArgs ...elementT) bool {
	t.Helper()

	if !expected.EqualPacker(actual) {
		expected.EqualPacker(actual)
		assert.Fail(t, fmt.Sprintf("packed unequal\nexpected: %v, actual:   %v",
			&branch{p: *expected}, &branch{p: *actual}), msgAndArgs...)
		return false
	}
	return true
}
