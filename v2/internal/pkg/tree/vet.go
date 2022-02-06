package tree

import (
	"log"

	"github.com/arr-ai/frozen/v2/pkg/errors"
	"github.com/arr-ai/frozen/v2/internal/pkg/vetctl"
)

const (
	vetting   = vetctl.Vetting
	vetReruns = vetctl.VetReruns
)

var vetFailed = false

func vet[T any](rerun func(), ins ...*Tree[T]) func(out *Tree[T]) {
	if !vetting {
		panic(errors.Errorf("call to (*Tree[T]).vet() not wrapped in if Vetting { ... }"))
	}
	if vetFailed {
		return func(out *Tree[T]) {}
	}

	check := func(trees ...*Tree[T]) {
		for _, t := range trees {
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
			if t != nil {
				t.Vet()
			}
		}
	}
	check(ins...)
	return func(out *Tree[T]) {
		if out != nil {
			ins = append(ins, out)
		}
		check(ins...)
	}
}
