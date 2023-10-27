package test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func assert(t *testing.T, pass bool, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	t.Helper()
	return check(t, pass, false, msgAndArgs2, msgAndArgs1...)
}

func require(t *testing.T, pass bool, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	t.Helper()
	return check(t, pass, true, msgAndArgs2, msgAndArgs1...)
}

func check(t *testing.T, pass, require bool, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	if pass {
		return true
	}
	t.Helper()
	return failX(t, require, msgAndArgs2, msgAndArgs1...)
}

func fail(t *testing.T, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	t.Helper()
	return failX(t, false, msgAndArgs2, msgAndArgs1...)
}

func failNow(t *testing.T, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	t.Helper()
	return failX(t, true, msgAndArgs2, msgAndArgs1...)
}

func failX(t *testing.T, now bool, msgAndArgs2 []any, msgAndArgs1 ...any) bool {
	t.Helper()
	var sb strings.Builder
	for _, msgAndArgs := range [][]any{msgAndArgs1, msgAndArgs2} {
		if len(msgAndArgs) > 0 {
			if sb.Len() > 0 {
				sb.WriteString(": ")
			}
			format := msgAndArgs[0].(string)
			args := msgAndArgs[1:]
			fmt.Fprintf(&sb, format, args...)
		}
	}
	t.Log(sb.String())
	if now {
		t.FailNow()
	} else {
		t.Fail()
	}
	return false
}

func sameElements(x, y any) bool { //nolint:cyclop
	u := deref(x)
	v := deref(y)

	if u.Kind() != v.Kind() {
		return false
	}

	switch u.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Slice:
		n := v.Len()
		if n != u.Len() {
			return false
		}
		indices := map[int]struct{}{}
		for j := n - 1; j >= 0; j-- {
			indices[j] = struct{}{}
		}
	outer:
		for i := u.Len() - 1; i >= 0; i-- {
			a := u.Index(i).Interface()
			for j := range indices {
				b := v.Index(j).Interface()
				if reflect.DeepEqual(a, b) {
					delete(indices, j)
					continue outer
				}
			}
			return false
		}
	case reflect.String:
		m := []byte(u.String())
		n := []byte(v.String())
		sort.Slice(m, func(i, j int) bool { return m[i] < n[j] })
		sort.Slice(n, func(i, j int) bool { return m[i] < n[j] })
		return string(m) == string(n)
	case reflect.Map:
		m := mapOf(u)
		n := mapOf(v)
		if len(m) != len(n) {
			return false
		}
		for k, a := range m {
			b, has := n[k]
			if !has || !reflect.DeepEqual(a, b) {
				return false
			}
		}
	default:
		return false
	}
	return true
}

func deref(x any) reflect.Value {
	v := reflect.ValueOf(x)
	kind := v.Kind()
	if kind == reflect.Pointer || kind == reflect.Interface {
		return v.Elem()
	}
	return v
}

func mapOf(v reflect.Value) map[any]any {
	m := map[any]any{}
	for iter := v.MapRange(); iter.Next(); {
		m[iter.Key()] = iter.Value()
	}
	return m
}
