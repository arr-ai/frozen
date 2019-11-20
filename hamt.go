package frozen

import (
	"fmt"
	"strings"
)

type hamt interface {
	isEmpty() bool
	count() int
	put(key, value interface{}) (result hamt, old *entry)
	putImpl(e entry, depth int, h *hasher) (result hamt, old *entry)
	get(key interface{}) (interface{}, bool)
	getImpl(key interface{}, h *hasher) (interface{}, bool)
	delete(key interface{}) (result hamt, old *entry)
	deleteImpl(key interface{}, h *hasher) (result hamt, old *entry)
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

func (e empty) put(key, value interface{}) (result hamt, old *entry) {
	return e.putImpl(entry{key: key, value: value}, 0, nil)
}

func (empty) putImpl(e entry, _ int, _ *hasher) (result hamt, old *entry) {
	return e, nil
}

func (empty) get(key interface{}) (interface{}, bool) {
	return nil, false
}

func (empty) getImpl(key interface{}, _ *hasher) (interface{}, bool) {
	return nil, false
}

func (e empty) delete(key interface{}) (result hamt, old *entry) {
	return e.deleteImpl(key, nil)
}

func (e empty) deleteImpl(key interface{}, _ *hasher) (result hamt, old *entry) {
	return e, nil
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
	return &full{
		base: [hamtSize]hamt{
			empty{}, empty{}, empty{}, empty{},
			empty{}, empty{}, empty{}, empty{},
		},
	}
}

func (f *full) isEmpty() bool {
	return f.base[0].isEmpty() && f.base[1].isEmpty() && f.base[2].isEmpty() && f.base[3].isEmpty() &&
		f.base[4].isEmpty() && f.base[5].isEmpty() && f.base[6].isEmpty() && f.base[7].isEmpty()
}

func (f *full) count() int {
	c := 0
	for _, b := range f.base {
		c += b.count()
	}
	return c
}

func (f *full) put(key, value interface{}) (result hamt, old *entry) {
	return f.putImpl(entry{key: key, value: value}, 0, newHasher(key, 0))
}

func (f *full) putImpl(e entry, depth int, h *hasher) (result hamt, old *entry) {
	offset := h.next(e.key)
	r, old := f.base[offset].putImpl(e, depth+1, h)
	return f.update(offset, r), old
}

func (f *full) get(key interface{}) (interface{}, bool) {
	return f.getImpl(key, newHasher(key, 0))
}

func (f *full) getImpl(key interface{}, h *hasher) (interface{}, bool) {
	return f.base[h.next(key)].getImpl(key, h)
}

func (f *full) delete(key interface{}) (result hamt, old *entry) {
	return f.deleteImpl(key, newHasher(key, 0))
}

func (f *full) deleteImpl(key interface{}, h *hasher) (result hamt, old *entry) {
	offset := h.next(key)
	if child, old := f.base[offset].deleteImpl(key, h); old != nil {
		return f.update(offset, child), old
	}
	return f, nil
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

func (e entry) put(key, value interface{}) (result hamt, old *entry) {
	return e.putImpl(entry{key: key, value: value}, 0, newHasher(key, 0))
}

func (e entry) putImpl(e2 entry, depth int, h *hasher) (result hamt, old *entry) {
	if equal(e.key, e2.key) {
		return e2, &e
	}
	result, _ = newFull().putImpl(e, depth, newHasher(e.key, depth))
	result, _ = result.(*full).putImpl(e2, depth, h)
	return result, nil
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

func (e entry) delete(key interface{}) (result hamt, old *entry) {
	return e.deleteImpl(key, nil)
}

func (e entry) deleteImpl(key interface{}, _ *hasher) (result hamt, old *entry) {
	if equal(key, e.key) {
		return empty{}, &e
	}
	return e, nil
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
	i   int
}

func newHamtIter(base []hamt) *hamtIter {
	return &hamtIter{stk: [][]hamt{base}, i: -1}
}

func (i *hamtIter) next() bool {
	for {
		if basep := &i.stk[len(i.stk)-1]; len(*basep) > 0 {
			b := (*basep)[0]
			*basep = (*basep)[1:]
			switch b := b.(type) {
			case entry:
				i.e = b
				i.i++
				return true
			case *full:
				i.stk = append(i.stk, b.base[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}
