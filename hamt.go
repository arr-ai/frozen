package frozen

import (
	"fmt"
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

type full struct {
	base [hamtSize]hamt
}

func newFull() *full {
	return &full{base: [4]hamt{empty{}, empty{}, empty{}, empty{}}}
}

func (f *full) isEmpty() bool {
	return f.base[0].isEmpty() && f.base[1].isEmpty() && f.base[2].isEmpty() && f.base[3].isEmpty()
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
		return f.update(offset, child), true
	}
	return f, false
}

func (f *full) update(offset int, t hamt) *full {
	h := newFull()
	copy(h.base[:], f.base[:])
	h.base[offset] = t
	return h
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
	return newFull().
		putImpl(e, depth, newHasher(e.key, depth)).(*full).
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

type hamtIter struct {
	stk [][]hamt
	e   entry
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
			case *full:
				i.stk = append(i.stk, b.base[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}
