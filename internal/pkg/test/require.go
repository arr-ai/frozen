package test

import (
	"reflect"
	"testing"
)

func RequireElementsMatch(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, sameElements(x, y), msgAndArgs, "elements differ:\n%v\n!=\n%v", x, y)
}

func RequireEmpty(t *testing.T, x any, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, reflect.ValueOf(x).Len() == 0, msgAndArgs, "%v not empty", x)
}

func RequireEqual(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, reflect.DeepEqual(x, y), msgAndArgs, "%v != %v", x, y)
}

func RequireFalse(t *testing.T, b bool, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, !b, msgAndArgs, "!false")
}

func RequireNoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, err == nil, msgAndArgs, "error %v != nil", err)
}

func RequireNotEqual(t *testing.T, x, y any, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, !reflect.DeepEqual(x, y), msgAndArgs, "%v != %v", x, y)
}

func RequireNoPanic(t *testing.T, f func(), msgAndArgs ...any) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			b = failNow(t, msgAndArgs, "panic(%v)", r)
		}
	}()
	f()
	return true
}

func RequirePanic(t *testing.T, f func(), msgAndArgs ...any) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			b = true
		}
	}()
	f()
	return failNow(t, msgAndArgs, "no panic")
}

func RequireTrue(t *testing.T, b bool, msgAndArgs ...any) bool {
	t.Helper()
	return require(t, b, msgAndArgs, "!true")
}
