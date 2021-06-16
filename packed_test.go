package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedWith(t *testing.T) {
	p := packed{}
	for i := 0; i < maxLeafLen; i++ {
		assertEqualPacked(t, p, p.With(newMasker(i), emptyNode{}, false), i)
	}
	for i := 0; i < maxLeafLen; i++ {
		q := p.With(newMasker(i), leaf{1}, false).With(newMasker(i), emptyNode{}, false)
		assertEqualPacked(t, p, q, i)
	}
}

func TestPackedWithMulti(t *testing.T) {
	p := packed{}.
		With(newMasker(1), leaf{1, 2}, false).
		With(newMasker(3), leaf{10, 20}, false).
		With(newMasker(3), emptyNode{}, false).
		With(newMasker(5), leaf{3, 4}, false)
	q := packed{}.
		With(newMasker(1), leaf{1, 2}, false).
		With(newMasker(3), emptyNode{}, false).
		With(newMasker(5), leaf{3, 4}, false)
	assertEqualPacked(t, p, q)
}

func assertEqualPacked(t *testing.T, expected, actual packed, msgAndArgs ...interface{}) bool {
	t.Helper()

	return assert.Equal(t, expected.mask, actual.mask, msgAndArgs...) &&
		assert.ElementsMatch(t, expected.data, actual.data, msgAndArgs...)
}
