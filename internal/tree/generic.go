package tree

import "github.com/arr-ai/hash"

type (
	actualInterface = interface{}
)

func hashValue(i interface{}, seed uintptr) uintptr {
	return hash.Interface(i, seed)
}
