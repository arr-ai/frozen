package test

import (
	"reflect"
	"runtime"
)

type markerKey struct {
	file string
	line int
	args []any
}

type Marker struct {
	key      *markerKey
	isTarget bool
}

func (m *Marker) IsTarget() bool {
	return m.isTarget
}

type Replayer struct {
	mark     func(args ...any) *Marker
	replayTo func(m *Marker)
}

func (r *Replayer) Mark(args ...any) *Marker {
	return r.mark(args...)
}

func (r *Replayer) ReplayTo(m *Marker) {
	r.replayTo(m)
}

func Replayable(enabled bool, f func(r *Replayer)) {
	if !enabled {
		m := &Marker{}
		f(&Replayer{
			mark:     func(...any) *Marker { return m },
			replayTo: func(*Marker) {},
		})
		return
	}
	var latest *markerKey
	var target *markerKey

	mark := func(args ...any) *Marker {
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
		return &Marker{
			key:      latest,
			isTarget: target != nil && reflect.DeepEqual(*latest, *target),
		}
	}

	replayTo := func(m *Marker) {
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
		f(&Replayer{mark: mark, replayTo: replayTo})
		return false
	}() { //nolint:revive
	}
}
