package resolvers

import (
	"github.com/arr-ai/frozen/internal/tree"
	"github.com/arr-ai/frozen/types"
)

func NewKeyValueResolver(name string, resolve func(key, a, b interface{}) interface{}) *tree.Resolver {
	return tree.NewResolver(name, func(a, b interface{}) interface{} {
		i := a.(types.KeyValue)
		j := b.(types.KeyValue)
		return types.KV(i.Key, resolve(i.Key, i.Value, j.Value))
	})
}
