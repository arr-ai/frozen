package tree

import "github.com/arr-ai/frozen/errors"

const vetting = false

func vetUnNode(nodes ...unNode) func(rets ...interface{}) {
	if !vetting {
		panic("Wrap calls to vetUnLeaf in if vetting { ... }")
	}
	vet := func() {
		for _, n := range nodes {
			switch n := n.(type) {
			case *unLeaf:
				count := 0
				for _, bucket := range n.data {
					for _, e := range bucket {
						switch e.(type) {
						case []interface{}:
							panic(errors.WTF)
						}
						count++
					}
				}
				if count != n.count {
					panic(errors.Errorf("vetUnNode: unLeaf count mismatch: count %d != n.count %d", count, n.count))
				}
			}
		}
	}
	vet()
	return func(rets ...interface{}) {
		nodes = nodes[:0]
		for _, ret := range rets {
			var n unNode
			switch ret := ret.(type) {
			case *unNode:
				n = *ret
			case **unEmptyNode:
				n = *ret
			case **unLeaf:
				n = *ret
			case **unBranch:
				n = *ret
			default:
				panic(errors.Errorf("can't vet %T", ret))
			}
			nodes = append(nodes, n)
		}
		vet()
	}
}
