package frozen

import (
	"github.com/arr-ai/frozen/internal/iterator"
	"github.com/arr-ai/frozen/internal/value"
	"github.com/arr-ai/frozen/pkg/kv"
)

type (
	KeyValue = kv.KeyValue
	Iterator = iterator.Iterator
	Key      = value.Key
)

// KV creates a KeyValue.
func KV(key, val interface{}) kv.KeyValue {
	return kv.KV(key, val)
}
