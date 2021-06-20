package iterator

import (
	"github.com/arr-ai/frozen/errors"
)

var Empty = empty{}

type empty struct{}

func (empty) Next() bool {
	return false
}

func (empty) Value() interface{} {
	panic(errors.WTF)
}
