package kvi

import (
	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/pkg/kv"
)

// Empty is the empty iterator.
var Empty = empty{}

type empty struct{}

func (empty) Next() bool {
	return false
}

func (empty) Value() kv.KeyValue {
	panic(errors.WTF)
}
