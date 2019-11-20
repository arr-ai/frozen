package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNest(t *testing.T) {
	t.Parallel()

	ca := NewRelation(
		[]interface{}{"c", "a"},
		[]interface{}{1, 10},
		[]interface{}{1, 11},
		[]interface{}{2, 13},
		[]interface{}{3, 11},
		[]interface{}{4, 14},
		[]interface{}{3, 10},
		[]interface{}{4, 13},
	)
	sharing := ca.
		Nest("aa", "a").
		Nest("cc", "c").
		Where(func(tuple interface{}) bool {
			return tuple.(Map).MustGet("cc").(Set).Count() > 1
		})
	expected := NewRelation(
		[]interface{}{"aa", "cc"},
		[]interface{}{
			NewRelation(
				[]interface{}{"a"},
				[]interface{}{10},
				[]interface{}{11},
			),
			NewRelation(
				[]interface{}{"c"},
				[]interface{}{1},
				[]interface{}{3},
			),
		},
	)
	assert.True(t, sharing.Equal(expected))
}
