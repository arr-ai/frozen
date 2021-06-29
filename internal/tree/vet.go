package tree

import (
	"log"
	"math/bits"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/pkg/vetctl"
)

const (
	vetting   = vetctl.Vetting
	vetReruns = vetctl.VetReruns
)

var vetFailed = false

func vet(rerun func(), ins ...node) func(out *node) {
	if !vetting {
		panic(errors.Errorf("call to (*Tree).vet() not wrapped in if Vetting { ... }"))
	}
	if vetFailed {
		return func(out *node) {}
	}

	check := func(nodes ...node) {
		for _, n := range nodes {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("vet failure: %v", r)
					if vetReruns {
						log.Print("spin-looping for vet repro. Hope you set a break-point!")

						// Prevent recursive calls to vet, which would break the stack.
						vetFailed = true
						for {
							rerun()
						}
					} else {
						panic(r)
					}
				}
			}()
			if n != nil {
				n.Vet()
			}
		}
	}
	check(ins...)
	return func(out *node) {
		if out != nil {
			ins = append(ins, *out)
		}
		check(ins...)
	}
}

func packedIteratorBuf(count int) [][]node {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]node, 0, depth)
}
