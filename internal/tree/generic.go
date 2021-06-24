package tree

import "github.com/arr-ai/hash"

type (
	elementT = interface{}
)

func hashValue(i elementT, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}
