package tree

import "fmt"

type h128 struct{ lo, hi uintptr }

func (h h128) xor(o h128) h128 { return h128{h.lo ^ o.lo, h.hi ^ o.hi} }
func (h h128) isZero() bool    { return h.lo == 0 && h.hi == 0 }
func (h h128) Lo() uintptr     { return h.lo }

func (h h128) String() string {
	return fmt.Sprintf("%x:%x", h.lo, h.hi)
}

func newElemH128[T any](v T, hf func(T, uintptr) uintptr) h128 {
	return h128{hf(v, 0), hf(v, 1)}
}
