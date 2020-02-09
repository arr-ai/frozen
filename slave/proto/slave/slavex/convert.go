package slavex

import (
	"fmt"

	"github.com/arr-ai/frozen/internal/tree"
	"github.com/arr-ai/frozen/slave/proto/slave"
	"github.com/arr-ai/frozen/types"
)

func ValueToIntf(v *slave.Value) (interface{}, error) {
	switch c := v.Choice.(type) {
	case *slave.Value_I:
		return c.I, nil
	case *slave.Value_F:
		return c.F, nil
	case *slave.Value_Kv:
		key, err := ValueToIntf(c.Kv.Key)
		if err != nil {
			return nil, err
		}
		value, err := ValueToIntf(c.Kv.Value)
		if err != nil {
			return nil, err
		}
		return types.KV(key, value), nil
	default:
		return nil, fmt.Errorf("unexpected value (%T)(%[1]v)", c)
	}
}

func IntfToValue(i interface{}) (*slave.Value, error) {
	switch i := i.(type) {
	case int64:
		return &slave.Value{Choice: &slave.Value_I{I: i}}, nil
	case float64:
		return &slave.Value{Choice: &slave.Value_F{F: i}}, nil
	case types.KeyValue:
		key, err := IntfToValue(i.Key)
		if err != nil {
			return nil, err
		}
		value, err := IntfToValue(i.Value)
		if err != nil {
			return nil, err
		}
		return &slave.Value{Choice: &slave.Value_Kv{Kv: &slave.Value_KV{
			Key:   key,
			Value: value,
		}}}, nil
	default:
		return nil, fmt.Errorf("unsupported value (%T)(%[1]v)", i)
	}
}

func TreeToNode(s *slave.Tree) (*tree.Node, error) {
	switch c := s.Choice.(type) {
	case *slave.Tree_Node_:
		children := c.Node.Children
		var nodes [tree.NodeCount]*tree.Node
		for mask := types.MaskIterator(c.Node.Mask); mask != 0; mask = mask.Next() {
			if len(children) == 0 {
				return nil, fmt.Errorf("not enough children for mask %d", c.Node.Mask)
			}
			node, err := TreeToNode(children[0])
			if err != nil {
				return nil, err
			}
			nodes[mask.Index()] = node
			children = children[1:]
		}
		if len(children) > 0 {
			return nil, fmt.Errorf("too many children for mask %d: %v", c.Node.Mask, children)
		}
		var node tree.Node
		node.SetChildren(types.MaskIterator(c.Node.Mask), &nodes)
		return &node, nil
	case *slave.Tree_Leaf_:
		elts := make([]interface{}, 0, len(c.Leaf.Values))
		for _, v := range c.Leaf.Values {
			i, err := ValueToIntf(v)
			if err != nil {
				return nil, err
			}
			elts = append(elts, i)
		}
		return tree.NewLeaf(elts...).Node(), nil
	default:
		return nil, fmt.Errorf("unexpected value (%T)(%[1]v)", c)
	}
}

func NodeToTree(t *tree.Node) (*slave.Tree, error) {
	if !t.IsLeaf() {
		mask, children := t.GetChildren()
		node := &slave.Tree_Node{
			Mask: uint32(mask),
		}
		for m := mask; m != 0; m = m.Next() {
			c, err := NodeToTree(children[m.Index()])
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, c)
		}
		return &slave.Tree{Choice: &slave.Tree_Node_{Node: node}}, nil
	}
	leaf := &slave.Tree_Leaf{}
	for i := t.Leaf().Range(); i.Next(); {
		value, err := IntfToValue(i.Value())
		if err != nil {
			return nil, err
		}
		leaf.Values = append(leaf.Values, value)
	}
	return &slave.Tree{Choice: &slave.Tree_Leaf_{Leaf: leaf}}, nil
}
