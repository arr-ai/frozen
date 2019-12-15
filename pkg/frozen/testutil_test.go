package frozen

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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

func assertSetEqual(t *testing.T, expected, actual Set, msgAndArgs ...interface{}) bool {
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

func assertSetHas(t *testing.T, s Set, i interface{}) bool {
	return assert.True(t, s.Has(i), "i=%v", i)
}

func assertSetNotHas(t *testing.T, s Set, i interface{}) bool {
	return assert.False(t, s.Has(i), "i=%v", i)
}

func assertMapEqual(t *testing.T, expected, actual Map, msgAndArgs ...interface{}) bool {
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

func assertMapHas(t *testing.T, m Map, i, expected interface{}) bool {
	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.True(t, has, "i=%v", i) && assert.Equal(t, expected, v, "i=%v", i)
	return ok1 && ok2
}

func assertMapNotHas(t *testing.T, m Map, i interface{}) bool {
	v, has := m.Get(i)
	ok1 := assert.Equal(t, has, m.Has(i))
	ok2 := assert.False(t, has, "i=%v v=%v", i, v)
	return ok1 && ok2
}
