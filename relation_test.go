package frozen_test

import (
	"fmt"
	"math/bits"
	"testing"

	. "github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/internal/pkg/test"
)

func TestJoinSimple(t *testing.T) {
	t.Parallel()

	a := NewRelation(
		[]interface{}{"x", "y"},
		[]interface{}{1, 2},
	)
	b := NewRelation(
		[]interface{}{"y", "z"},
		[]interface{}{2, 3},
	)
	expected := NewRelation(
		[]interface{}{"x", "y", "z"},
		[]interface{}{1, 2, 3},
	)
	actual := a.Join(b)
	test.AssertSetEqual(t, expected, actual)
}

// We use numbers as follows to represent tuples:
// 0bZZYYXX = (x:XX, y:YY, z:ZZ).
// When XX, YY or ZZ are zero, they are considered to be not in the map.
// Relations are numbers with multiple sequences of the above pattern.
//   - 0b001001001110 = {(x:1, y:2), (x:2, y:3)}
//   - 0b101000_110100 = {(y:1, z:3), (y:2, z:2)}
type bitRelation uint64

// Since we work with a{x, y} and b{y, z} as inputs, we will initially
// work with the 0bYYXX pattern and expandBinaryToTernaryBitRelation it to 0b00YYXX or 0bZZYY00.
func expandBinaryToTernaryBitRelation(a, offset uint64) bitRelation {
	result := uint64(0)
	for ; a != 0; a >>= 4 {
		// If XX or YY is zero, discard the whole relation.
		//   0011        1100          1111        1111
		if a&3 == 0 || a&0xc == 0 || a&15 < result&15 {
			return 0
		}
		result = result<<6 | a&15 /*1111*/
	}
	return bitRelation(result << offset)
}

func (a bitRelation) toRelation() Set {
	header := []interface{}{"x", "y", "z"}

	lo, hi := 0, len(header)
	if a&3 == 0 {
		lo++
	}
	if a>>4&3 == 0 {
		hi--
	}
	h := header[lo:hi]
	rows := make([][]interface{}, 0, (bits.Len(uint(a))+5)/6)
	for ; a != 0; a >>= 6 {
		if a&12 == 0 { // 001100
			panic(fmt.Sprintf("a=%b", a))
		}
		row := make([]interface{}, 0, hi-lo)
		for i := 2 * lo; i < 2*hi; i += 2 {
			row = append(row, a>>uint(i)&3)
		}
		rows = append(rows, row)
	}
	return NewRelation(h, rows...)
}

func (a bitRelation) join(b bitRelation) bitRelation {
	result := bitRelation(0)
	for i := a; i != 0; i >>= 6 {
		for j := b; j != 0; j >>= 6 {
			//   001100  001100
			if i&12 == j&12 {
				result = result<<6 | (i|j)&0x3f // 111111
			}
		}
	}
	return result
}

func TestJoinExhaustive(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	outerTotal := 0
	for i0 := uint64(1); i0 < 0x100; /*1_0000_0000*/ i0++ {
		i0 := i0
		//    0011         1100           0011_0000       1100_0000               1111
		if i0&3 == 0 || i0&0xc == 0 || i0&0x30 == 0 || i0&0xc0 == 0 || i0>>4 < i0&0xf {
			continue
		}
		t.Run(fmt.Sprintf("0b%04b_%04b", i0>>4, i0&0xf), func(t *testing.T) {
			testJoinExhaustiveCase(t, i0)
		})
		outerTotal++
	}
}

func testJoinExhaustiveCase(t *testing.T, i0 uint64) {
	t.Helper()
	t.Parallel()

	innerTotal := 0
	for i := i0; i < 0x1000; i += 0x100 {
		a := expandBinaryToTernaryBitRelation(i, 0)
		if a == 0 {
			continue
		}
		for j := uint64(1); j < 0x1000; j++ {
			b := expandBinaryToTernaryBitRelation(j, 2)
			if b == 0 {
				continue
			}
			c := a.join(b)
			setA := a.toRelation()
			setB := b.toRelation()
			setC := c.toRelation()
			if !test.AssertSetEqual(t, setC, setA.Join(setB), "a=%b=%v b=%b=%v", a, setA, b, setB) {
				_ = a.join(b)
				setA.Join(setB)
				t.FailNow()
			}
			innerTotal++
		}
	}
	t.Logf("%d scenarios tested", innerTotal)
}

func TestNest(t *testing.T) {
	t.Parallel()

	ca := NewRelation(
		[]interface{}{"c", "a"},
		[]interface{}{1, 10},
		[]interface{}{1, 11},
		[]interface{}{2, 13},
		[]interface{}{3, 11},
		[]interface{}{4, 14},
		[]interface{}{3, 10},
		[]interface{}{4, 13},
	)
	sharing := ca.
		Nest(NewMap(KV("aa", NewSet("a")))).
		Nest(NewMap(KV("cc", NewSet("c")))).
		Where(func(tuple interface{}) bool {
			return tuple.(Map).MustGet("cc").(Set).Count() > 1
		})
	expected := NewRelation(
		[]interface{}{"aa", "cc"},
		[]interface{}{
			NewRelation(
				[]interface{}{"a"},
				[]interface{}{10},
				[]interface{}{11},
			),
			NewRelation(
				[]interface{}{"c"},
				[]interface{}{1},
				[]interface{}{3},
			),
		},
	)
	test.AssertSetEqual(t, expected, sharing)
}

func TestUnnest(t *testing.T) {
	t.Parallel()

	sharing := NewRelation(
		[]interface{}{"aa", "cc"},
		[]interface{}{
			NewRelation(
				[]interface{}{"a"},
				[]interface{}{10},
				[]interface{}{11},
			),
			NewRelation(
				[]interface{}{"c"},
				[]interface{}{1},
				[]interface{}{3},
			),
		},
	)
	expected := NewRelation(
		[]interface{}{"c", "a"},
		[]interface{}{1, 10},
		[]interface{}{1, 11},
		[]interface{}{3, 11},
		[]interface{}{3, 10},
	)

	actual := sharing.Unnest(NewSet("cc", "aa"))
	test.AssertSetEqual(t, expected, actual)

	actual = sharing.Unnest(NewSet("aa", "cc"))
	test.AssertSetEqual(t, expected, actual)
}
