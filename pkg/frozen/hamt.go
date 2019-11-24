package frozen

import (
	"fmt"
	"strings"
)

const (
	hamtBits = 2
	hamtSize = 1 << hamtBits
	hamtMask = hamtSize - 1
)

type hasher uint64

// TODO: make lazy if depth == 0.
func newHasher(key interface{}, depth int) hasher {
	// Use the high four bits as the seed.
	h := hasher(0b1111<<60 | hash(key))
	for i := 0; i < depth; i++ {
		h = h.next(key)
	}
	return h
}

func (h hasher) next(key interface{}) hasher {
	if h >>= hamtBits; h < 0b1_0000 {
		return (h-1)<<60 | hasher(hash([2]interface{}{int(h), key})>>4)
	}
	return h
}

func (h hasher) hash() int {
	return int(h & hamtMask)
}

type element interface{}

type hamt interface {
	isEmpty() bool
	put(elem element, pool *buffer) (result hamt, old element)
	putImpl(elem element, pool *buffer, depth int, h hasher) (result hamt, old element)
	get(elem element) (element, bool)
	getImpl(elem element, h hasher) (element, bool)
	delete(elem element, pool *buffer) (result hamt, old element)
	deleteImpl(elem element, pool *buffer, h hasher) (result hamt, old element)
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

func (e empty) put(elem element, pool *buffer) (result hamt, old element) {
	return e.putImpl(elem, pool, 0, 0)
}

func (empty) putImpl(elem element, pool *buffer, _ int, _ hasher) (result hamt, old element) {
	return entry{elem: elem}, nil
}

func (empty) get(elem element) (element, bool) {
	return nil, false
}

func (empty) getImpl(elem element, _ hasher) (element, bool) {
	return nil, false
}

func (e empty) delete(elem element, pool *buffer) (result hamt, old element) {
	return e.deleteImpl(elem, pool, 0)
}

func (e empty) deleteImpl(elem element, pool *buffer, _ hasher) (result hamt, old element) {
	return e, nil
}

func (empty) String() string {
	return "∅"
}

func (empty) iterator() *hamtIter {
	return newHamtIter(nil)
}

type full [hamtSize]hamt

func (f *full) isEmpty() bool {
	return false
}

func (f *full) put(elem element, pool *buffer) (result hamt, old element) {
	return f.putImpl(elem, pool, 0, newHasher(elem, 0))
}

func (f *full) putImpl(elem element, pool *buffer, depth int, h hasher) (result hamt, old element) {
	offset := h.hash()
	t, old := f[offset].putImpl(elem, pool, depth+1, h.next(elem))
	return f.update(offset, t, elem, pool), old
}

func (f *full) get(elem element) (element, bool) {
	return f.getImpl(elem, newHasher(elem, 0))
}

func (f *full) getImpl(elem element, h hasher) (element, bool) {
	return f[h.hash()].getImpl(elem, h.next(elem))
}

func (f *full) delete(elem element, pool *buffer) (result hamt, old element) {
	return f.deleteImpl(elem, pool, newHasher(elem, 0))
}

func (f *full) deleteImpl(elem element, pool *buffer, h hasher) (result hamt, old element) {
	offset := h.hash()
	if child, old := f[offset].deleteImpl(elem, pool, h.next(elem)); old != nil {
		return f.update(offset, child, elem, pool), old
	}
	return f, nil
}

func (f *full) update(offset int, t hamt, elem element, pool *buffer) hamt {
	if t.isEmpty() {
		for i, b := range f {
			if i != offset && !b.isEmpty() {
				goto notempty
			}
		}
		return empty{}
	}
notempty:
	h := pool.copy(f)
	h[offset] = t
	return h
}

func (f *full) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range f {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(v.String())
	}
	b.WriteString("]")
	return b.String()
}

func (f *full) iterator() *hamtIter {
	return newHamtIter(f[:])
}

type entry struct {
	elem interface{}
}

func (entry) isEmpty() bool {
	return false
}

func (entry) count() int {
	return 1
}

func (e entry) put(elem element, pool *buffer) (result hamt, old element) {
	return e.putImpl(elem, pool, 0, newHasher(elem, 0))
}

var empties = func() *full {
	f := &full{}
	for i := range f {
		f[i] = empty{}
	}
	return f
}()

func (e entry) putImpl(elem element, pool *buffer, depth int, h hasher) (result hamt, old element) {
	if equal(elem, e.elem) {
		return entry{elem: elem}, e.elem
	}

	result, _ = pool.copy(empties).putImpl(e.elem, pool, depth, newHasher(e.elem, depth))
	result, _ = result.(*full).putImpl(elem, pool, depth, h)
	return result, nil
}

func (e entry) get(elem element) (element, bool) {
	return e.getImpl(elem, 0)
}

func (e entry) getImpl(elem element, _ hasher) (element, bool) {
	if equal(elem, e.elem) {
		return e.elem, true
	}
	return nil, false
}

func (e entry) delete(elem element, pool *buffer) (result hamt, old element) {
	return e.deleteImpl(elem, pool, 0)
}

func (e entry) deleteImpl(elem element, pool *buffer, _ hasher) (result hamt, old element) {
	if equal(elem, e.elem) {
		return empty{}, e.elem
	}
	return e, nil
}

func (e entry) String() string {
	return fmt.Sprintf("%v", e.elem)
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
				i.stk = append(i.stk, b[:])
			}
		} else if i.stk = i.stk[:len(i.stk)-1]; len(i.stk) == 0 {
			return false
		}
	}
}
