package test

import (
	"testing"

	"github.com/arr-ai/frozen"
)

func AssertSetEqual(t *testing.T, expected, actual frozen.Set, msgAndArgs ...interface{}) bool {
	t.Helper()

	format := "\nexpected: %10v !=\nactual:   %10v"
	args := []interface{}{}
	if len(msgAndArgs) > 0 {
		format = msgAndArgs[0].(string) + format
		args = append(args, msgAndArgs[1:]...)
	}

	args = append(args, expected, actual)
	if !expected.Equal(actual) {
		t.Errorf(format, args...)
		return false
	}
	return true
}
