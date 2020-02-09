package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/slave/proto/slave"
	"github.com/arr-ai/frozen/types"
)

func FromSlaveValue(v *slave.Value) (interface{}, error) {
	switch c := v.Choice.(type) {
	case *slave.Value_I:
		return int(c.I), nil
	case *slave.Value_U:
		return uint(c.U), nil
	case *slave.Value_D:
		return c.D, nil
	case *slave.Value_Kv:
		key, err := FromSlaveValue(c.Kv.Key)
		if err != nil {
			return nil, err
		}
		value, err := FromSlaveValue(c.Kv.Value)
		if err != nil {
			return nil, err
		}
		return types.KV(key, value), nil
	case *slave.Value_B:
		return c.B, nil
	case *slave.Value_S:
		return c.S, nil
	case *slave.Value_Data:
		return c.Data, nil
	case *slave.Value_I8:
		return int8(c.I8), nil
	case *slave.Value_I16:
		return int16(c.I16), nil
	case *slave.Value_I32:
		return c.I32, nil
	case *slave.Value_I64:
		return c.I64, nil
	case *slave.Value_U8:
		return uint8(c.U8), nil
	case *slave.Value_U16:
		return uint16(c.U16), nil
	case *slave.Value_U32:
		return c.U32, nil
	case *slave.Value_U64:
		return c.U64, nil
	case *slave.Value_Uptr:
		return uintptr(c.Uptr), nil
	case *slave.Value_F:
		return c.F, nil
	case *slave.Value_C64:
		return complex(c.C64.Re, c.C64.Im), nil
	case *slave.Value_C128:
		return complex(c.C128.Re, c.C128.Im), nil
	default:
		return nil, fmt.Errorf("unexpected value (%T)(%[1]v)", c)
	}
}

func ToSlaveValue(i interface{}) (*slave.Value, error) {
	switch i := i.(type) {
	case int:
		return &slave.Value{Choice: &slave.Value_I{I: int64(i)}}, nil
	case uint:
		return &slave.Value{Choice: &slave.Value_U{U: uint64(i)}}, nil
	case float64:
		return &slave.Value{Choice: &slave.Value_D{D: i}}, nil
	case types.KeyValue:
		key, err := ToSlaveValue(i.Key)
		if err != nil {
			return nil, err
		}
		value, err := ToSlaveValue(i.Value)
		if err != nil {
			return nil, err
		}
		return &slave.Value{Choice: &slave.Value_Kv{Kv: &slave.Value_KV{
			Key:   key,
			Value: value,
		}}}, nil
	case bool:
		return &slave.Value{Choice: &slave.Value_B{B: i}}, nil
	case string:
		return &slave.Value{Choice: &slave.Value_S{S: i}}, nil
	case []byte:
		return &slave.Value{Choice: &slave.Value_Data{Data: i}}, nil
	case int8:
		return &slave.Value{Choice: &slave.Value_I8{I8: int32(i)}}, nil
	case int16:
		return &slave.Value{Choice: &slave.Value_I16{I16: int32(i)}}, nil
	case int32:
		return &slave.Value{Choice: &slave.Value_I32{I32: i}}, nil
	case int64:
		return &slave.Value{Choice: &slave.Value_I64{I64: i}}, nil
	case uint8:
		return &slave.Value{Choice: &slave.Value_U8{U8: uint32(i)}}, nil
	case uint16:
		return &slave.Value{Choice: &slave.Value_U16{U16: uint32(i)}}, nil
	case uint32:
		return &slave.Value{Choice: &slave.Value_U32{U32: i}}, nil
	case uint64:
		return &slave.Value{Choice: &slave.Value_U64{U64: i}}, nil
	case uintptr:
		return &slave.Value{Choice: &slave.Value_Uptr{Uptr: uint64(i)}}, nil
	case float32:
		return &slave.Value{Choice: &slave.Value_F{F: i}}, nil
	case complex64:
		return &slave.Value{Choice: &slave.Value_C64{
			C64: &slave.Value_Complex64{Re: real(i), Im: imag(i)},
		}}, nil
	case complex128:
		return &slave.Value{Choice: &slave.Value_C128{
			C128: &slave.Value_Complex128{Re: real(i), Im: imag(i)},
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported value (%T)(%[1]v)", i)
	}
}

func FromSlaveTree(s *slave.Tree) (*Node, error) {
	switch c := s.Choice.(type) {
	case *slave.Tree_Node_:
		children := c.Node.Children
		var nodes [NodeCount]*Node
		for mask := types.MaskIterator(c.Node.Mask); mask != 0; mask = mask.Next() {
			if len(children) == 0 {
				return nil, fmt.Errorf("not enough children for mask %d", c.Node.Mask)
			}
			node, err := FromSlaveTree(children[0])
			if err != nil {
				return nil, err
			}
			nodes[mask.Index()] = node
			children = children[1:]
		}
		if len(children) > 0 {
			return nil, fmt.Errorf("too many children for mask %d: %v", c.Node.Mask, children)
		}
		var node Node
		node.SetChildren(types.MaskIterator(c.Node.Mask), &nodes)
		return &node, nil
	case *slave.Tree_Leaf_:
		elts := make([]interface{}, 0, len(c.Leaf.Values))
		for _, v := range c.Leaf.Values {
			i, err := FromSlaveValue(v)
			if err != nil {
				return nil, err
			}
			elts = append(elts, i)
		}
		return NewLeaf(elts...).Node(), nil
	default:
		return nil, fmt.Errorf("unexpected value (%T)(%[1]v)", c)
	}
}

func ToSlaveTree(t *Node) (*slave.Tree, error) {
	if !t.IsLeaf() {
		mask, children := t.GetChildren()
		node := &slave.Tree_Node{
			Mask: uint32(mask),
		}
		for m := mask; m != 0; m = m.Next() {
			c, err := ToSlaveTree(children[m.Index()])
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, c)
		}
		return &slave.Tree{Choice: &slave.Tree_Node_{Node: node}}, nil
	}
	leaf := &slave.Tree_Leaf{}
	for i := t.Leaf().Range(); i.Next(); {
		value, err := ToSlaveValue(i.Value())
		if err != nil {
			return nil, err
		}
		leaf.Values = append(leaf.Values, value)
	}
	return &slave.Tree{Choice: &slave.Tree_Leaf_{Leaf: leaf}}, nil
}
