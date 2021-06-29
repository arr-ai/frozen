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
		assertEqualPacked(t, p, p.With(i, nil), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.With(i, newLeaf1(1)).With(i, nil)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := (&packer{}).
		With(1, newLeaf2(1, 2)).
		With(3, newLeaf2(10, 20)).
		With(3, nil).
		With(5, newLeaf2(3, 4))
	q := (&packer{}).
		With(1, newLeaf2(1, 2)).
		With(3, nil).
		With(5, newLeaf2(3, 4))
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
