// +build !unsafe

package tree

type node struct {
	b      branch
	l      leaf
	isLeaf bool
}

func (n *node) Leaf() *leaf {
	if n.isLeaf {
		return &n.l
	}
	return nil
}

func (n *node) Branch() *branch {
	if !n.isLeaf {
		return &n.b
	}
	return nil
}

type leaf struct {
	data []elementT
	n    *node
}

func newLeaf(data ...elementT) *leaf {
	n := &node{isLeaf: true, l: leaf{data: data}}
	n.l.n = n
	return &n.l
}

func (l *leaf) Node() *node {
	return l.n
}

type branch struct {
	p packer
	n *node
}

func newBranch(p *packer) *branch {
	n := &node{}
	if p != nil {
		n.b.p = *p
	}
	n.b.n = n
	return &n.b
}

func (b *branch) Node() *node {
	return b.n
}
