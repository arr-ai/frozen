package main

import (
	"fmt"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/slave/proto/slave"
)

func fromValue(v *slave.Value) interface{} {
	switch w := v.Choice.(type) {
	case *slave.Value_I:
		return w.I
	case *slave.Value_F:
		return w.F
	case *slave.Value_Set:
		return fromSet(w.Set)
	case *slave.Value_Map:
		return fromMap(w.Map)
	default:
		panic(fmt.Errorf("unexpected value %v", w))
	}
}

func fromSet(s *slave.Set) frozen.Set {
	var b frozen.SetBuilder
	for _, e := range s.Element {
		b.Add(fromValue(e))
	}
	return b.Finish()
}

func fromMap(s *slave.Map) frozen.Map {
	var b frozen.MapBuilder
	for _, e := range s.Entry {
		b.Put(fromValue(e.Key), fromValue(e.Value))
	}
	return b.Finish()
}
