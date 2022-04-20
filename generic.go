package frozen

import (
	"github.com/arr-ai/frozen/pkg/kv"
)

// KV creates a KeyValue.
func KV[K, V any](k K, v V) kv.KeyValue[K, V] {
	return kv.KV(k, v)
}
