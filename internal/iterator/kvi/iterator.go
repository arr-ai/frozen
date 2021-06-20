package kvi

import "github.com/arr-ai/frozen/pkg/kv"

// Iterator provides for iterating over a Set.
type Iterator interface {
	Next() bool
	Value() kv.KeyValue
}
