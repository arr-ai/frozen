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

type replayer struct {
	mark     func(args ...interface{}) *marker
	replayTo func(m *marker)
}

func (r replayer) replay() {
	r.replayTo(nil)
}

func replayable(enabled bool, f func(r replayer)) {
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

		replayTo := func(m *marker) {
			if m == nil {
				panic(latest)
			}
			panic(m.key)
		}

		for func() (again bool) {
			defer func() {
				if err := recover(); err != nil {
					if key, ok := err.(*markerKey); ok {
						target = key
						again = true
					} else {
						panic(err)
					}
				}
			}()
			f(replayer{mark: mark, replayTo: replayTo})
			return false
		}() {
		}
	} else {
		m := &marker{}
		f(replayer{
			mark:     func(_ ...interface{}) *marker { return m },
			replayTo: func(_ *marker) {},
		})
	}
}
