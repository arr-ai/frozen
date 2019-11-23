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
		Nest("aa", NewSet("a")).
		Nest("cc", NewSet("c")).
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
	if !assert.True(t, sharing.Equal(expected)) {
		t.Log(ca)
		t.Log(ca.Nest("aa", NewSet("a")))
		t.Log(ca.Nest("aa", NewSet("a")).Nest("cc", NewSet("c")))
		t.Log(ca.Nest("aa", NewSet("a")).Nest("cc", NewSet("c")).
			Where(func(tuple interface{}) bool {
				return tuple.(Map).MustGet("cc").(Set).Count() > 1
			}))
		t.Log(expected)
	}
}

func TestUnnest(t *testing.T) {
	t.Parallel()

	sharing := NewRelation(
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
	expected := NewRelation(
		[]interface{}{"c", "a"},
		[]interface{}{1, 10},
		[]interface{}{1, 11},
		[]interface{}{3, 11},
		[]interface{}{3, 10},
	)

	actual := sharing.Unnest(NewSet("cc", "aa"))
	assert.True(t, actual.Equal(expected), "%v !=\n%v", expected, actual)

	actual = sharing.Unnest(NewSet("aa", "cc"))
	assert.True(t, actual.Equal(expected), "%v !=\n%v", expected, actual)
}
