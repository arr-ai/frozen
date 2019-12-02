package frozen

import "sync"

var memoizePrepop = func(prepare func(n int) interface{}) func(n int) interface{} {
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
