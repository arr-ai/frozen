package frozen

import (
	"github.com/arr-ai/frozen/types"
)

type Iterator = types.Iterator

type KeyValue = types.KeyValue

func KV(key, val interface{}) KeyValue {
	return types.KV(key, val)
}
