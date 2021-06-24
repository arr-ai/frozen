package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	t.Parallel()

	p := packer{}
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.With(i, theEmptyNode), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.With(i, leaf{1}).With(i, theEmptyNode)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := packer{}.
		With(1, leaf{1, 2}).
		With(3, leaf{10, 20}).
		With(3, theEmptyNode).
		With(5, leaf{3, 4})
	q := packer{}.
		With(1, leaf{1, 2}).
		With(3, theEmptyNode).
		With(5, leaf{3, 4})
	assertEqualPacked(t, p, q)
}

//nolint:unparam
func assertEqualPacked(t *testing.T, expected, actual packer, msgAndArgs ...elementT) bool {
	t.Helper()

	return assert.Equal(t, expected, actual, msgAndArgs...)
}
