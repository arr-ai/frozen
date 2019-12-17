package frozen

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/require"
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
	assertSetEqual(t, expected, actual)
}

func TestJoinExhaustive(t *testing.T) {
	t.Parallel()

	// We use numbers as follows to represent tuples:
	// 0bZZYYXX = (x:XX, y:YY, z:ZZ).
	// When XX, YY or ZZ are zero, they are considered to be not in the map.
	// Relations are numbers with multiple sequences of the above pattern.
	//   - 0b001001_001110 = {(x:1, y:2), (x:2, y:3)}
	//   - 0b101000_110100 = {(y:1, z:3), (y:2, z:2)}

	// Since we work with a{x, y} and b{y, z} as inputs, we will initially
	// work with the 0bYYXX pattern and expand it to 0b00YYXX or 0bZZYY00.
	expand := func(a uint64, offset uint64) uint64 {
		result := uint64(0)
		for ; a != 0; a >>= 4 {
			// If XX or YY is zero, discard the whole relation.
			if a&0b0011 == 0 || a&0b1100 == 0 || a&0b1111 < result&0b1111 {
				return 0
			}
			result = result<<6 | a&0b1111
		}
		return result << offset
	}

	header := []interface{}{"x", "y", "z"}

	relation := func(a uint64) Set {
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
			require.NotZero(t, a&0b001100, "a=%b", a)
			row := make([]interface{}, 0, hi-lo)
			for i := 2 * lo; i < 2*hi; i += 2 {
				row = append(row, a>>i&3)
			}
			rows = append(rows, row)
		}
		return NewRelation(h, rows...)
	}

	join := func(a, b uint64) uint64 {
		const yMask = 0b001100_001100_001100
		result := uint64(0)
		for i := a; i != 0; i >>= 6 {
			for j := b; j != 0; j >>= 6 {
				if i&0b001100 == j&0b001100 {
					result = result<<6 | (i|j)&0b111111
				}
			}
		}
		return result
	}

	outerTotal := 0
	for i0 := uint64(1); i0 < 0b1_0000_0000; i0++ {
		i0 := i0
		if i0&0b0011 == 0 || i0&0b1100 == 0 || i0&0b0011_0000 == 0 || i0&0b1100_0000 == 0 || i0>>4 < i0&0b1111 {
			continue
		}
		t.Run(fmt.Sprintf("0b%04b_%04b", i0>>4, i0&0b1111), func(t *testing.T) {
			t.Parallel()

			innerTotal := 0
			for i := i0; i < 0b1_0000_0000_0000; i += 0b1_0000_0000 {
				if a := expand(i, 0); a != 0 {
					for j := uint64(1); j < 0b1_0000_0000_0000; j++ {
						if b := expand(j, 2); b != 0 {
							c := join(a, b)
							setA := relation(a)
							setB := relation(b)
							setC := relation(c)
							if !assertSetEqual(t, setC, setA.Join(setB), "a=%b=%v b=%b=%v", a, setA, b, setB) {
								join(a, b)
								setA.Join(setB)
								t.FailNow()
							}
							innerTotal++
						}
					}
				}
			}
			t.Logf("%d scenarios tested", innerTotal)
		})
		outerTotal++
	}
	t.Logf("%d subtests created", outerTotal)
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
	assertSetEqual(t, expected, sharing)
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
	assertSetEqual(t, expected, actual)

	actual = sharing.Unnest(NewSet("aa", "cc"))
	assertSetEqual(t, expected, actual)
}
