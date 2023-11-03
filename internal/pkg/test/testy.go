package test

import (
	"reflect"
	"testing"
)

func Fail(t *testing.T, msgAndArgs ...any) bool {
	t.Helper()
	return fail(t, msgAndArgs)
}

func FailNow(t *testing.T, msgAndArgs ...any) bool {
	t.Helper()
	return fail(t, msgAndArgs)
}

func ElementsMatch(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, sameElements(x, y), msgAndArgs, "elements differ:\n%v\n!=\n%v", x, y)
}

func Empty(t *testing.T, x any, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, reflect.ValueOf(x).Len() == 0, msgAndArgs, "%v not empty", x)
}

func Equal(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, reflect.DeepEqual(x, y), msgAndArgs, "%v != %v", x, y)
}

func False(t *testing.T, b bool, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, !b, msgAndArgs, "!false")
}

func NoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, err == nil, msgAndArgs, "error %v != nil", err)
}

func NotEqual(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, !reflect.DeepEqual(x, y), msgAndArgs, "%v != %v", x, y)
}

func NoPanic(t *testing.T, f func(), msgAndArgs ...any) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			b = fail(t, msgAndArgs, "panic(%v)", r)
		}
	}()
	f()
	return true
}

func Panic(t *testing.T, f func(), msgAndArgs ...any) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			b = true
		}
	}()
	f()
	return fail(t, msgAndArgs, "no panic")
}

func True(t *testing.T, b bool, msgAndArgs ...any) bool {
	t.Helper()
	return assert(t, b, msgAndArgs, "!true")
}
