package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskerFirst(t *testing.T) {
	t.Parallel()

	assert.Equal(t, masker(0b0000), masker(0b0000).first())
	// TODO: more tests
}

func TestMaskerFirstIsIn(t *testing.T) {
	t.Parallel()

	for a := 1; a <= 0b1_0000; a += 2 {
		for b := 0; b < 0b1_0000; b += 2 {
			assert.False(t, masker(a).firstIsIn(masker(b)), "%b %b", a, b)
			assert.True(t, masker(a).firstIsIn(masker(b|1)), "%b %b", a, b)
		}
	}
	for a := 2; a <= 0b1_0000; a += 4 {
		for b := 0; b <= 0b1_0000; b += 4 {
			assert.False(t, masker(a).firstIsIn(masker(b)), "%b %b", a, b)
			assert.True(t, masker(a).firstIsIn(masker(b|10)), "%b %b", a, b)
		}
	}
	assert.False(t, masker(8).firstIsIn(masker(0)))
	assert.True(t, masker(8).firstIsIn(masker(8)))
}

func TestMaskerOffset(t *testing.T) {
	t.Parallel()

	// TODO: more tests
}
