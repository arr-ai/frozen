// Generated by gen-kv.pl kvt kv.KeyValue. DO NOT EDIT.
package kvt

import (
	"fmt"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/fu"
)

type twig struct {
	data []elementT
}

func newTwig(data ...elementT) *twig {
	return &twig{data: data}
}

// fmt.Formatter

func (l *twig) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	for i, e := range l.data {
		if i > 0 {
			fu.WriteString(f, ",")
		}
		fu.Format(e, f, verb)
	}
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *twig) String() string {
	return fmt.Sprintf("%s", l)
}

// node

func (l *twig) Add(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			l.data[i] = args.f(e, v)
			return l, matches
		}
	}
	if len(l.data) < cap(l.data) || depth >= maxTreeDepth {
		l.data = append(l.data, v)
		return l, matches
	}

	b := &branch{}
	for _, e := range l.data {
		_, m := b.Add(args, e, depth, newHasher(e, depth))
		matches += m
	}
	_, m := b.Add(args, v, depth, h)
	matches += m

	return b, matches
}

func (l *twig) Canonical(depth int) node {
	switch n := len(l.data); n {
	case 0:
		return nil
	case 1:
		return newLeaf1(l.data[0])
	case 2:
		return newLeaf2(l.data[0], l.data[1])
	default:
		if n <= maxLeafLen || depth*fanoutBits >= 64 {
			return l
		}
		return newBranchFrom(depth, l.data...)
	}
}

func (l *twig) Combine(args *CombineArgs, n node, depth int) (_ node, matches int) { //nolint:cyclop
	var ndata []elementT
	switch n := n.(type) {
	case *twig:
		ndata = n.data
	default:
		panic(errors.WTF)
	}

	cloned := false
scanning:
	for i, e := range ndata {
		for j, f := range l.data {
			if args.eq(f, e) {
				if !cloned {
					l = l.cloneWithExtra(0)
					cloned = true
				}
				l.data[j] = args.f(f, e)
				matches++
				continue scanning
			}
		}
		if len(l.data) < maxLeafLen {
			l = newTwig(append(l.data, e)...)
		} else {
			b := &branch{}
			for _, e := range l.data {
				_, m := b.Add(args, e, depth, newHasher(e, depth))
				matches += m
			}
			for _, e := range ndata[i:] {
				_, m := b.Add(args, e, depth, newHasher(e, depth))
				matches += m
			}
		}
	}
	if len(l.data) > maxLeafLen {
		panic(errors.WTF)
	}
	return l.Canonical(depth), matches
}

func (l *twig) AppendTo(dest []elementT) []elementT {
	if len(dest)+len(l.data) > cap(dest) {
		return nil
	}
	return append(dest, l.data...)
}

func (l *twig) Difference(args *EqArgs, n node, depth int) (_ node, matches int) {
	ret := newTwig()
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args.Flip(), e, h) == nil {
			ret.data = append(ret.data, e)
		} else {
			matches++
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig) Empty() bool {
	return len(l.data) == 0
}

func (l *twig) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*twig); is {
		if len(l.data) != len(n.data) {
			return false
		}
		for _, e := range l.data {
			if n.Get(args, e, 0) == nil {
				return false
			}
		}
		return true
	}
	return false
}

func (l *twig) Get(args *EqArgs, v elementT, _ hasher) *elementT {
	for i, e := range l.data {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *twig) Intersection(args *EqArgs, n node, depth int) (_ node, matches int) {
	ret := newTwig()
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args, e, h) != nil {
			matches++
			ret.data = append(ret.data, e)
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig) Iterator([][]node) Iterator {
	return newSliceIterator(l.data)
}

func (l *twig) Reduce(_ NodeArgs, _ int, r func(values ...elementT) elementT) elementT {
	return r(l.data...)
}

func (l *twig) Remove(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			last := len(l.data) - 1
			if last == 0 {
				return nil, matches
			}
			if i < last {
				l.data[i] = l.data[last]
			}
			l.data = l.data[:last]
			return l.Canonical(depth), matches
		}
	}
	return l, matches
}

func (l *twig) SubsetOf(args *EqArgs, n node, depth int) bool {
	for _, e := range l.data {
		h := newHasher(e, depth)
		if n.Get(args, e, h) == nil {
			return false
		}
	}
	return true
}

func (l *twig) Map(args *CombineArgs, _ int, f func(e elementT) elementT) (_ node, matches int) {
	var nb Builder
	for _, e := range l.data {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	return t.root, t.count
}

func (l *twig) Vet() int {
	if len(l.data) <= 2 {
		panic(errors.Errorf("twig too small (%d)", len(l.data)))
	}
	h0 := newHasher(l.data[0], 0)
	for _, e := range l.data[1:] {
		if newHasher(e, 0) != h0 {
			panic(errors.Errorf("twig with multiple hashes"))
		}
	}
	return len(l.data)
}

func (l *twig) Where(args *WhereArgs, depth int) (_ node, matches int) {
	ret := newTwig()
	for _, e := range l.data {
		if args.Pred(e) {
			ret.data = append(ret.data, e)
			matches++
		}
	}
	return ret.Canonical(depth), matches
}

func (l *twig) With(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			ret := l.cloneWithExtra(0)
			ret.data[i] = args.f(ret.data[i], v)
			return ret, matches
		}
	}
	return newTwig(append(l.data, v)...).Canonical(depth), matches
}

func (l *twig) Without(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	for i, e := range l.data {
		if args.eq(e, v) {
			matches++
			ret := newTwig(l.data[:len(l.data)-1]...).cloneWithExtra(0)
			if i != len(ret.data) {
				ret.data[i] = l.data[len(ret.data)]
			}
			return ret.Canonical(depth), matches
		}
	}
	return l, matches
}

func (l *twig) clone() node {
	return &twig{data: append(([]elementT)(nil), l.data...)}
}

func (l *twig) cloneWithExtra(extra int) *twig {
	return newTwig(append(make([]elementT, 0, len(l.data)+extra), l.data...)...)
}
