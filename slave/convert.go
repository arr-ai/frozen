package main

// func fromValue(v *slave.Value) interface{} {
// 	switch w := v.Choice.(type) {
// 	case *slave.Value_I:
// 		return w.I
// 	case *slave.Value_F:
// 		return w.F
// 	case *slave.Value_Kv:
// 		return fromKV(w.Map)
// 	case *slave.Value_Set:
// 		return fromSet(w.Set)
// 	default:
// 		panic(fmt.Errorf("unexpected value %v", w))
// 	}
// }

// func fromTree(s *slave.Tree) frozen.Set {
// 	for _, e := range s.Element {
// 		b.Add(fromValue(e))
// 	}
// 	return b.Finish()
// }

// func fromKV(s *slave.KV) frozen.Map {
// 	var b frozen.MapBuilder
// 	for _, e := range s.Entry {
// 		b.Put(fromValue(e.Key), fromValue(e.Value))
// 	}
// 	return b.Finish()
// }
