package frozen

import "github.com/arr-ai/frozen/internal/tree"

var useRHS = tree.NewResolver(func(_, b interface{}) interface{} { return b })
var useLHS = tree.NewResolver(func(a, _ interface{}) interface{} { return a })
