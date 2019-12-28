package frozen

import (
	"encoding/json"
)

// Iota returns Iota3(0, stop, 1).
func Iota(stop int) Set {
	return Iota3(0, stop, 1)
}

// Iota2 returns Iota3(start, stop, 1).
func Iota2(start, stop int) Set {
	return Iota3(start, stop, 1)
}

// Iota3 returns a Set with elements {start, start+step, start+2*step, ...} up
// to but not including stop. Negative steps are allowed.
func Iota3(start, stop, step int) Set {
	if step == 0 {
		if start == stop {
			return Set{}
		}
		panic("zero step size")
	}
	var b SetBuilder
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
func NewSetFromMask64(mask uint64) Set {
	var b SetBuilder
	for mask := BitIterator(mask); mask != 0; mask = mask.Next() {
		b.Add(mask.Index())
	}
	return b.Finish()
}

// MarshalJSON implements json.Marshaler.
func (s Set) MarshalJSON() ([]byte, error) {
	proxy := make([]interface{}, 0, s.Count())
	for i := s.Range(); i.Next(); {
		proxy = append(proxy, i.Value())
	}
	return json.Marshal(proxy)
}

// Ensure that Set implements json.Marshaler.
var _ json.Marshaler = Set{}
