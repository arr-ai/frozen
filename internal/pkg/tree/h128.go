package tree

import "fmt"

type H128 struct{ lo, hi uintptr }

func (h H128) xor(o H128) H128 { return H128{h.lo ^ o.lo, h.hi ^ o.hi} }
func (h H128) isZero() bool    { return h.lo == 0 && h.hi == 0 }
func (h H128) Lo() uintptr     { return h.lo }

func (h H128) String() string {
	return fmt.Sprintf("%x:%x", h.lo, h.hi)
}

func newElemH128[T any](v T, hf func(T, uintptr) uintptr) H128 {
	return H128{hf(v, 0), hf(v, 1)}
}
