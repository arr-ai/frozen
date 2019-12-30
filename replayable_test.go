package frozen

import (
	"reflect"
	"runtime"
)

type markerKey struct {
	file string
	line int
	args []interface{}
}

type marker struct {
	key      *markerKey
	isTarget bool
}

func replayable(enabled bool, f func(mark func(args ...interface{}) *marker, replay func(m *marker))) {
	if enabled {
		var latest *markerKey
		var target *markerKey

		mark := func(args ...interface{}) *marker {
			var pc [1]uintptr
			if runtime.Callers(2, pc[:]) != 1 {
				panic("unable to set mark.")
			}
			frames := runtime.CallersFrames(pc[:])
			frame, _ := frames.Next()

			latest = &markerKey{
				file: frame.File,
				line: frame.Line,
				args: args,
			}
			return &marker{
				key:      latest,
				isTarget: target != nil && reflect.DeepEqual(*latest, *target),
			}
		}

		replay := func(m *marker) {
			if m == nil {
				panic(latest)
			}
			panic(m)
		}

		for func() (again bool) {
			defer func() {
				if err := recover(); err != nil {
					if m, ok := err.(*marker); ok {
						target = m.key
						again = true
					}
				} else {
					panic(err)
				}
			}()
			f(mark, replay)
			return false
		}() {
		}
	} else {
		m := &marker{}
		f(
			func(args ...interface{}) *marker { return m },
			func(m *marker) {},
		)
	}
}
