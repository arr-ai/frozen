package test

import (
	"testing"

	"github.com/arr-ai/frozen/v2"
)

func AssertSetEqual[T comparable](t *testing.T, expected, actual frozen.Set[T], msgAndArgs ...interface{}) bool {
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
		e := mapFromSet(expected)
		a := mapFromSet(actual)
		l := map[T]struct{}{}
		r := map[T]struct{}{}
		for k := range e {
			if _, has := a[k]; !has {
				l[k] = struct{}{}
			}
		}
		for k := range a {
			if _, has := e[k]; !has {
				r[k] = struct{}{}
			}
		}
		t.Errorf("l=%v\nr=%v", len(l), len(r))
		return false
	}
	return true
}

func mapFromSet[T comparable](s frozen.Set[T]) map[T]struct{} {
	m := make(map[T]struct{}, s.Count())
	for r := s.Range(); r.Next(); {
		m[r.Value()] = struct{}{}
	}
	return m
}
