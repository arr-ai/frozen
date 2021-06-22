package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	t.Parallel()

	p := packer{}
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.With(newMasker(i), theEmptyNode), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.With(newMasker(i), leaf{1}).With(newMasker(i), theEmptyNode)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	t.Parallel()

	p := packer{}.
		With(newMasker(1), leaf{1, 2}).
		With(newMasker(3), leaf{10, 20}).
		With(newMasker(3), theEmptyNode).
		With(newMasker(5), leaf{3, 4})
	q := packer{}.
		With(newMasker(1), leaf{1, 2}).
		With(newMasker(3), theEmptyNode).
		With(newMasker(5), leaf{3, 4})
	assertEqualPacked(t, p, q)
}

//nolint:unparam
func assertEqualPacked(t *testing.T, expected, actual packer, msgAndArgs ...interface{}) bool {
	t.Helper()

	return assert.Equal(t, expected.mask, actual.mask, msgAndArgs...) &&
		assert.ElementsMatch(t, expected.data, actual.data, msgAndArgs...)
}
