package frozen

var useRHS = func(_, b interface{}) interface{} { return b }
var useLHS = func(a, _ interface{}) interface{} { return a }
