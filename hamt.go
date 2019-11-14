package frozen

import "math/bits"

type hamt struct {
	bitmap uint64
	base   []interface{}
}

type entry struct {
	key, value interface{}
}

func (t hamt) isEmpty() bool {
	return t.bitmap == 0
}

func (t hamt) count() int {
	c := 0
	for _, b := range t.base {
		switch b := b.(type) {
		case hamt:
			c += b.count()
		case entry:
			c++
		}
	}
	return c
}

func (t hamt) put(key, value interface{}) hamt {
	return t.putImpl(key, value, 0, newHasher(key, 0))
}

func (t hamt) putImpl(key, value interface{}, depth int, h *hasher) hamt {
	bit := h.next()
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	if t.bitmap&bit == 0 {
		return t.insert(bit, offset, entry{key: key, value: value})
	}
	switch i := t.base[offset].(type) {
	case entry:
		if equal(i.key, key) {
			return t.update(offset, entry{key: key, value: value})
		}
		return t.update(offset,
			hamt{}.
				putImpl(i.key, i.value, depth+1, newHasher(i.key, depth+1)).
				putImpl(key, value, depth+1, h),
		)
	case hamt:
		return t.update(offset, i.putImpl(key, value, depth+1, h))
	default:
		panic("wat?")
	}
}

func (t hamt) get(key interface{}) (interface{}, bool) {
	return t.getImpl(key, newHasher(key, 0))
}

func (t hamt) getImpl(key interface{}, h *hasher) (interface{}, bool) {
	bit := h.next()
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	if t.bitmap&bit != 0 {
		switch i := t.base[offset].(type) {
		case entry:
			if equal(i.key, key) {
				return i.value, true
			}
		case hamt:
			return i.getImpl(key, h)
		}
	}
	return nil, false
}

func (t hamt) delete(key interface{}) hamt {
	return t.deleteImpl(key, newHasher(key, 0))
}

func (t hamt) deleteImpl(key interface{}, h *hasher) hamt {
	bit := h.next()
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	if t.bitmap&bit != 0 {
		switch i := t.base[offset].(type) {
		case entry:
			if equal(i.key, key) {
				return t.remove(bit, offset)
			}
		case hamt:
			if child := i.deleteImpl(key, h); child.bitmap != 0 {
				return t.update(offset, child)
			}
			return t.remove(bit, offset)
		}
	}
	return t
}

func (t hamt) iterator() *hamtIter {
	return &hamtIter{stk: [][]interface{}{t.base}}
}

type hamtIter struct {
	stk [][]interface{}
	e   entry
}

func (i *hamtIter) next() bool {
	for {
		if basep := &i.stk[len(i.stk)-1]; len(*basep) > 0 {
			b := (*basep)[0]
			*basep = (*basep)[1:]
			switch b := b.(type) {
			case entry:
				i.e = b
				return true
			case hamt:
				i.stk = append(i.stk, b.base)
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}

func (t hamt) insert(bit uint64, offset int, item interface{}) hamt {
	base := make([]interface{}, len(t.base)+1)
	copy(base, t.base[:offset])
	copy(base[offset+1:], t.base[offset:])
	base[offset] = item
	return hamt{
		bitmap: t.bitmap | bit,
		base:   base,
	}
}

func (t hamt) update(offset int, item interface{}) hamt {
	base := make([]interface{}, len(t.base))
	copy(base, t.base)
	base[offset] = item
	return hamt{bitmap: t.bitmap, base: base}
}

func (t hamt) remove(bit uint64, offset int) hamt {
	base := make([]interface{}, len(t.base)-1)
	copy(base, t.base[:offset])
	copy(base[offset:], t.base[offset+1:])
	return hamt{bitmap: t.bitmap & ^bit, base: base}
}
