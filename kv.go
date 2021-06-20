package frozen

import "github.com/arr-ai/frozen/pkg/kv"

type KeyValue = kv.KeyValue

// KV creates a KeyValue.
func KV(key, val interface{}) kv.KeyValue {
	return kv.KV(key, val)
}
