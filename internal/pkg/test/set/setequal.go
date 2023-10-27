package testset

import (
	"testing"

	"github.com/arr-ai/frozen"
)

func AssertSetEqual[T any](t *testing.T, expected, actual frozen.Set[T], msgAndArgs ...any) bool {
	t.Helper()

	format := "\nexpected: %10v !=\nactual:   %10v"
	args := []any{}
	if len(msgAndArgs) > 0 {
		format = msgAndArgs[0].(string) + format
		args = append(args, msgAndArgs[1:]...)
	}

	args = append(args, expected, actual)
	if !expected.Equal(actual) {
		t.Errorf(format, args...)
		l := expected.Difference(actual)
		r := actual.Difference(expected)
		t.Errorf("l=%v\nr=%v", l.Count(), r.Count())
		return false
	}
	return true
}
