package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	var p packed
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.with(newMasker(i), emptyNode{}), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.with(newMasker(i), leaf{1}).with(newMasker(i), emptyNode{})
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	p := packed{}.
		with(newMasker(1), leaf{1, 2}).
		with(newMasker(3), leaf{10, 20}).
		with(newMasker(3), emptyNode{}).
		with(newMasker(5), leaf{3, 4})
	q := packed{}.
		with(newMasker(1), leaf{1, 2}).
		with(newMasker(3), emptyNode{}).
		with(newMasker(5), leaf{3, 4})
	assertEqualPacked(t, p, q)
}

func assertEqualPacked(t *testing.T, expected, actual packed, msgAndArgs ...interface{}) bool {
	t.Helper()

	return assert.Equal(t, expected.mask, actual.mask, msgAndArgs...) &&
		assert.ElementsMatch(t, expected.data, actual.data, msgAndArgs...)
}
