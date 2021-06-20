package tree

type unDefroster struct {
	n node
}

var _ unNode = unDefroster{}

func (d unDefroster) Add(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Add(args, v, depth, h, matches)
}

func (d unDefroster) copyTo(to *unLeaf) {
	for i := d.n.Iterator(make([]packer, 0, maxTreeDepth)); i.Next(); {
		to.data = append(to.data, i.Value())
	}
}

func (d unDefroster) countUpTo(max int) int {
	return d.n.CountUpTo(max)
}

func (d unDefroster) Freeze() node {
	return d.n
}

func (d unDefroster) Get(args *EqArgs, v interface{}, h hasher) *interface{} {
	return d.n.Get(args, v, h)
}

func (d unDefroster) Remove(args *EqArgs, v interface{}, depth int, h hasher, matches *int) unNode {
	return d.n.Defrost().Remove(args, v, depth, h, matches)
}
