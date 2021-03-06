package frozen_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/arr-ai/frozen"
)

func memoizePrepop(prepare func(n int) interface{}) func(n int) interface{} {
	var lk sync.Mutex
	prepop := map[int]interface{}{}
	return func(n int) interface{} {
		lk.Lock()
		defer lk.Unlock()
		if data, has := prepop[n]; has {
			return data
		}
		data := prepare(n)
		prepop[n] = data
		return data
	}
}

// func intSetToMap(s Set) map[int]bool {
// 	result := make(map[int]bool, s.Count())
// 	for i := s.Range(); i.Next(); {
// 		result[i.Value().(int)] = true
// 	}
// 	return result
// }

// func intSetDiff(a, b Set) (l, m, r []int) {
// 	ma := intSetToMap(a)
// 	mb := intSetToMap(b)
// 	for e := range ma {
// 		if mb[e] {
// 			m = append(m, e)
// 		} else {
// 			l = append(l, e)
// 		}
// 	}
// 	for e := range mb {
// 		if !ma[e] {
// 			r = append(r, e)
// 		}
// 	}
// 	sort.Ints(l)
// 	sort.Ints(m)
// 	sort.Ints(r)
// 	return
// }

func assertSetHas(t *testing.T, s Set, i interface{}) bool { //nolint:unparam
	t.Helper()

	return assert.True(t, s.Has(i), "i=%v", i)
}

func assertSetNotHas(t *testing.T, s Set, i interface{}) bool { //nolint:unparam
	t.Helper()

	return assert.False(t, s.Has(i), "i=%v", i)
}

func assertMapEqual(t *testing.T, expected, actual Map, msgAndArgs ...interface{}) bool { //nolint:unparam
	t.Helper()

	format := "\nexpected %v != \nactual   %v"
	args := []interface{}{}
	if len(msgAndArgs) > 0 {
		format = msgAndArgs[0].(string) + format
		args = append(append(args, format), msgAndArgs[1:]...)
	} else {
		args = append(args, format)
	}
	args = append(args, expected, actual)
	return assert.True(t, expected.Equal(actual), args...)
}

func assertMapHas(t *testing.T, m Map, i, expected interface{}) bool { //nolint:unparam
	t.Helper()

	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.True(t, has, "i=%v", i) && assert.Equal(t, expected, v, "i=%v", i)
	return ok1 && ok2
}

func assertMapNotHas(t *testing.T, m Map, i interface{}) bool { //nolint:unparam
	t.Helper()

	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.False(t, has, "i=%v v=%v", i, v)
	return ok1 && ok2
}

func assertStringMapEqual(t *testing.T, expected, actual StringMap, msgAndArgs ...interface{}) bool {
	t.Helper()

	format := "\nexpected %v != \nactual   %v"
	args := []interface{}{}
	if len(msgAndArgs) > 0 {
		format = msgAndArgs[0].(string) + format
		args = append(append(args, format), msgAndArgs[1:]...)
	} else {
		args = append(args, format)
	}
	args = append(args, expected, actual)
	return assert.True(t, expected.Equal(actual), args...)
}

func assertStringMapHas(t *testing.T, m StringMap, i string, expected interface{}) bool { //nolint:unparam
	t.Helper()

	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.True(t, has, "i=%v", i) && assert.Equal(t, expected, v, "i=%v", i)
	return ok1 && ok2
}

func assertStringMapNotHas(t *testing.T, m StringMap, i string) bool { //nolint:unparam
	t.Helper()

	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.False(t, has, "i=%v v=%v", i, v)
	return ok1 && ok2
}

type mapOfSet map[string]Set

func generateSortedIntArray(start, end, step int) []interface{} {
	if step == 0 {
		if start == step {
			return []interface{}{}
		}
		panic("zero step size")
	}
	if (step > 0 && start > end) || (step < 0 && start < end) {
		panic("array will never reach end value")
	}
	n := (start - end) / step
	if n < 0 {
		n *= -1
	}
	result := make([]interface{}, n)
	currentVal := start
	for i := 0; i < n; i++ {
		result[i] = currentVal + step*i
	}
	return result
}
