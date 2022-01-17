package masker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arr-ai/frozen/v2/internal/pkg/masker"
)

func TestMaskerFirst(t *testing.T) {
	t.Parallel()

	assert.Equal(t, masker.Masker(0b0000), masker.Masker(0b0000).First())
	// TODO: more tests
}

func TestMaskerFirstIsIn(t *testing.T) {
	t.Parallel()

	for a := 1; a <= 0b1_0000; a += 2 {
		for b := 0; b < 0b1_0000; b += 2 {
			assert.False(t, masker.Masker(a).FirstIsIn(masker.Masker(b)), "%b %b", a, b)
			assert.True(t, masker.Masker(a).FirstIsIn(masker.Masker(b|1)), "%b %b", a, b)
		}
	}
	for a := 2; a <= 0b1_0000; a += 4 {
		for b := 0; b <= 0b1_0000; b += 4 {
			assert.False(t, masker.Masker(a).FirstIsIn(masker.Masker(b)), "%b %b", a, b)
			assert.True(t, masker.Masker(a).FirstIsIn(masker.Masker(b|10)), "%b %b", a, b)
		}
	}
	assert.False(t, masker.Masker(8).FirstIsIn(masker.Masker(0)))
	assert.True(t, masker.Masker(8).FirstIsIn(masker.Masker(8)))
}

func TestMaskerOffset(t *testing.T) {
	t.Parallel()

	// TODO: more tests
}
