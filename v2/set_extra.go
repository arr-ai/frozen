package frozen

import (
	"encoding/json"

	"github.com/arr-ai/frozen/errors"
)

// Iota returns Iota3(0, stop, 1).
func Iota[T ~int](stop T) Set[T] {
	return Iota3(0, stop, 1)
}

// Iota2 returns Iota3(start, stop, 1).
func Iota2[T ~int](start, stop T) Set[T] {
	return Iota3(start, stop, 1)
}

// Iota3 returns a Set with elements {start, start+step, start+2*step, ...} up
// to but not including stop. Negative steps are allowed.
func Iota3[T ~int](start, stop, step T) Set[T] {
	if step == 0 {
		if start == stop {
			return Set[T]{}
		}
		panic("zero step size")
	}
	var b SetBuilder[T]
	if step > 0 {
		for i := start; i < stop; i += step {
			b.Add(i)
		}
	} else {
		for i := start; i > stop; i += step {
			b.Add(i)
		}
	}
	return b.Finish()
}

// NewSetFromMask64 returns a Set containing all elements 2**i such that bit i
// of mask is set.
func NewSetFromMask64(mask uint64) Set[int] {
	var b SetBuilder[int]
	for mask := BitIterator(mask); mask != 0; mask = mask.Next() {
		i := mask.Index()
		b.Add(i)
	}
	return b.Finish()
}

// MarshalJSON implements json.Marshaler.
func (s Set[T]) MarshalJSON() ([]byte, error) {
	proxy := make([]T, 0, s.Count())
	for i := s.Range(); i.Next(); {
		proxy = append(proxy, i.Value())
	}
	data, err := json.Marshal(proxy)
	return data, errors.Wrap(err, 0)
}

// Ensure that Set implements json.Marshaler.
var _ json.Marshaler = Set[int]{}
