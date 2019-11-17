package frozen

import (
	"fmt"
	"math/bits"
	"strings"
)

type hamt interface {
	isEmpty() bool
	count() int
	put(key, value interface{}) hamt
	putImpl(e entry, depth int, h *hasher) hamt
	get(key interface{}) (interface{}, bool)
	getImpl(key interface{}, h *hasher) (interface{}, bool)
	delete(key interface{}) hamt
	deleteImpl(key interface{}, h *hasher) (hamt, bool)
	validate()
	String() string
	iterator() *hamtIter
}

type empty struct{}

func (empty) isEmpty() bool {
	return true
}

func (empty) count() int {
	return 0
}

func (e empty) put(key, value interface{}) hamt {
	return e.putImpl(entry{key: key, value: value}, 0, nil)
}

func (empty) putImpl(e entry, _ int, _ *hasher) hamt {
	return e
}

func (empty) get(key interface{}) (interface{}, bool) {
	return nil, false
}

func (empty) getImpl(key interface{}, _ *hasher) (interface{}, bool) {
	return nil, false
}

func (e empty) delete(key interface{}) hamt {
	return e
}

func (e empty) deleteImpl(key interface{}, h *hasher) (hamt, bool) {
	return e, false
}

func (empty) validate() {}

func (empty) String() string {
	return "∅"
}

func (empty) iterator() *hamtIter {
	return newHamtIter(nil)
}

type part struct {
	bitmap uint64
	base   []hamt
}

func (t part) isEmpty() bool {
	return false
}

func (t part) count() int {
	c := 0
	for _, b := range t.base {
		c += b.count()
	}
	return c
}

func (t part) put(key, value interface{}) hamt {
	return t.putImpl(entry{key: key, value: value}, 0, newHasher(key, 0))
}

func (t part) putImpl(e entry, depth int, h *hasher) hamt {
	bit := uint64(1) << h.next()
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	if t.bitmap&bit == 0 {
		if bitmap := t.bitmap | bit; bitmap < 1<<hamtSize-1 {
			return part{
				bitmap: bitmap,
				base:   insert(t.base, offset, e),
			}
		}
		var f full
		copy(f.base[:offset], t.base[:offset])
		copy(f.base[offset+1:], t.base[offset:])
		f.base[offset] = e
		return &f
	}
	return t.update(offset, t.base[offset].putImpl(e, depth+1, h))
}

func (t part) get(key interface{}) (interface{}, bool) {
	return t.getImpl(key, newHasher(key, 0))
}

func (t part) getImpl(key interface{}, h *hasher) (interface{}, bool) {
	bit := uint64(1) << h.next()
	if t.bitmap&bit == 0 {
		return nil, false
	}
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	return t.base[offset].getImpl(key, h)
}

func (t part) delete(key interface{}) hamt {
	h, _ := t.deleteImpl(key, newHasher(key, 0))
	return h
}

func (t part) deleteImpl(key interface{}, h *hasher) (hamt, bool) {
	bit := uint64(1) << h.next()
	offset := bits.OnesCount64(t.bitmap & (bit - 1))
	if t.bitmap&bit != 0 {
		if child, deleted := t.base[offset].deleteImpl(key, h); deleted {
			if !child.isEmpty() {
				return t.update(offset, child), true
			}
			return t.remove(bit, offset), true
		}
	}
	return t, false
}

func (t part) update(offset int, n hamt) part {
	return part{bitmap: t.bitmap, base: update(t.base, offset, n)}
}

func (t part) remove(bit uint64, offset int) hamt {
	if bitmap := t.bitmap & ^bit; bitmap != 0 {
		return part{bitmap: bitmap, base: remove(t.base, offset)}
	}
	return empty{}
}

func (t part) validate() {
	if bits.OnesCount64(t.bitmap) != len(t.base) {
		panic(fmt.Sprintf("part=%v", t))
	}
	for _, v := range t.base {
		v.validate()
	}
}

func (t part) String() string {
	var b strings.Builder
	b.WriteString("{")
	for i := 0; i < hamtSize; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		if bit := uint64(1) << i; t.bitmap&bit != 0 {
			offset := bits.OnesCount64(t.bitmap & (bit - 1))
			b.WriteString(t.base[offset].String())
		}
	}
	b.WriteString("}")
	return b.String()
}

func (t part) iterator() *hamtIter {
	return newHamtIter(t.base)
}

type hamtIter struct {
	stk [][]hamt
	e   entry
}

type full struct {
	base [hamtSize]hamt
}

func (f *full) isEmpty() bool {
	return false
}

func (f *full) count() int {
	c := 0
	for _, b := range f.base {
		c += b.count()
	}
	return c
}

func (f *full) put(key, value interface{}) hamt {
	return f.putImpl(entry{key: key, value: value}, 0, newHasher(key, 0))
}

func (f *full) putImpl(e entry, depth int, h *hasher) hamt {
	offset := h.next()
	return f.update(offset, f.base[offset].putImpl(e, depth+1, h))
}

func (f *full) get(key interface{}) (interface{}, bool) {
	return f.getImpl(key, newHasher(key, 0))
}

func (f *full) getImpl(key interface{}, h *hasher) (interface{}, bool) {
	return f.base[h.next()].getImpl(key, h)
}

func (f *full) delete(key interface{}) hamt {
	h, _ := f.deleteImpl(key, newHasher(key, 0))
	return h
}

func (f *full) deleteImpl(key interface{}, h *hasher) (hamt, bool) {
	offset := h.next()
	if child, deleted := f.base[offset].deleteImpl(key, h); deleted {
		if !child.isEmpty() {
			return f.update(offset, child), true
		}
		return f.remove(offset), true
	}
	return f, false
}

func (f *full) update(offset int, t hamt) *full {
	base := f.base
	base[offset] = t
	return &full{base: base}
}

func (f *full) remove(offset int) part {
	return part{
		bitmap: uint64((1<<hamtSize - 1) & ^(uint64(1) << offset)),
		base:   remove(f.base[:], offset),
	}
}

func (f *full) validate() {
	for _, v := range f.base {
		v.validate()
	}
}

func (f *full) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range f.base {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(v.String())
	}
	b.WriteString("]")
	return b.String()
}

func (f *full) iterator() *hamtIter {
	return newHamtIter(f.base[:])
}

type entry struct {
	key, value interface{}
}

func (e entry) isEmpty() bool {
	return false
}

func (e entry) count() int {
	return 1
}

func (e entry) put(key, value interface{}) hamt {
	return e.putImpl(entry{key: key, value: value}, 0, newHasher(key, 0))
}

func (e entry) putImpl(e2 entry, depth int, h *hasher) hamt {
	if equal(e.key, e2.key) {
		return e2
	}
	return part{}.
		putImpl(e, depth, newHasher(e.key, depth)).(part).
		putImpl(e2, depth, h)
}

func (e entry) get(key interface{}) (interface{}, bool) {
	return e.getImpl(key, nil)
}

func (e entry) getImpl(key interface{}, _ *hasher) (interface{}, bool) {
	if equal(key, e.key) {
		return e.value, true
	}
	return nil, false
}

func (e entry) delete(key interface{}) hamt {
	h, _ := e.deleteImpl(key, nil)
	return h
}

func (e entry) deleteImpl(key interface{}, _ *hasher) (hamt, bool) {
	if equal(key, e.key) {
		return empty{}, true
	}
	return e, false
}

func (e entry) validate() {}

func (e entry) String() string {
	return fmt.Sprintf("%v:%v", e.key, e.value)
}

func (e entry) iterator() *hamtIter {
	return newHamtIter([]hamt{e})
}

func insert(tbase []hamt, offset int, n hamt) []hamt {
	base := make([]hamt, len(tbase)+1)
	copy(base, tbase[:offset])
	copy(base[offset+1:], tbase[offset:])
	base[offset] = n
	return base
}

func update(tbase []hamt, offset int, n hamt) []hamt {
	base := make([]hamt, len(tbase))
	copy(base, tbase)
	base[offset] = n
	return base
}

func remove(tbase []hamt, offset int) []hamt {
	base := make([]hamt, len(tbase)-1)
	copy(base, tbase[:offset])
	copy(base[offset:], tbase[offset+1:])
	return base
}

func newHamtIter(base []hamt) *hamtIter {
	return &hamtIter{stk: [][]hamt{base}}
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
			case part:
				i.stk = append(i.stk, b.base)
			case *full:
				i.stk = append(i.stk, b.base[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}
