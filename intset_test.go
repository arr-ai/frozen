package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntSetEmpty(t *testing.T) {
	var s IntSet
	assert.True(t, s.IsEmpty())
}
