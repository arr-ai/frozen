package rel_test

import (
	"fmt"
	"math/bits"
	"testing"

	. "github.com/arr-ai/frozen"
	testset "github.com/arr-ai/frozen/internal/pkg/test/set"
	. "github.com/arr-ai/frozen/pkg/rel"
)

func TestJoinSimple(t *testing.T) {
	t.Parallel()

	a := New(
		[]string{"x", "y"},
		[]any{1, 2},
	)
	b := New(
		[]string{"y", "z"},
		[]any{2, 3},
	)
	expected := New(
		[]string{"x", "y", "z"},
		[]any{1, 2, 3},
	)
	actual := Join(a, b)
	testset.AssertSetEqual(t, expected, actual)
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

func (a bitRelation) toRelation() Relation {
	header := []string{"x", "y", "z"}

	lo, hi := 0, len(header)
	if a&3 == 0 {
		lo++
	}
	if a>>4&3 == 0 {
		hi--
	}
	h := header[lo:hi]
	rows := make([][]any, 0, (bits.Len(uint(a))+5)/6)
	for ; a != 0; a >>= 6 {
		if a&12 == 0 { // 001100
			panic(fmt.Sprintf("a=%b", a))
		}
		row := make([]any, 0, hi-lo)
		for i := 2 * lo; i < 2*hi; i += 2 {
			row = append(row, int(a>>uint(i)&3))
		}
		rows = append(rows, row)
	}
	return New(h, rows...)
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
			if !testset.AssertSetEqual(t, setC, Join(setA, setB), "a=%b=%v b=%b=%v", a, setA, b, setB) {
				_ = a.join(b)
				Join(setA, setB)
				t.FailNow()
			}
			innerTotal++
		}
	}
	t.Logf("%d scenarios tested", innerTotal)
}

func TestNest(t *testing.T) {
	t.Parallel()

	ca := New(
		[]string{"c", "a"},
		[]any{1, 10},
		[]any{1, 11},
		// []any{2, 13},
		[]any{3, 11},
		// []any{4, 14},
		[]any{3, 10},
		// []any{4, 13},
	)
	sharing := Nest(ca, NewMap(KV("aa", NewSet("a"))))
	// t.Log(sharing)
	sharing = Nest(sharing, NewMap(KV("cc", NewSet("c"))))
	// t.Log(sharing)
	sharing = sharing.Where(func(t Tuple) bool {
		return t.MustGet("cc").(Relation).Count() > 1
	})
	// t.Log(sharing)
	expected := New(
		[]string{"aa", "cc"},
		[]any{
			New([]string{"a"}, []any{10}, []any{11}),
			New([]string{"c"}, []any{1}, []any{3}),
		},
	)
	testset.AssertSetEqual(t, expected, sharing)
}

func TestUnnest(t *testing.T) {
	t.Parallel()

	sharing := New(
		[]string{"aa", "cc"},
		[]any{
			New(
				[]string{"a"},
				[]any{10},
				[]any{11},
			),
			New(
				[]string{"c"},
				[]any{1},
				[]any{3},
			),
		},
	)
	expected := New(
		[]string{"c", "a"},
		[]any{1, 10},
		[]any{1, 11},
		[]any{3, 11},
		[]any{3, 10},
	)

	actual := Unnest(Unnest(sharing, "cc"), "aa")
	testset.AssertSetEqual(t, expected, actual)

	actual = Unnest(Unnest(sharing, "aa"), "cc")
	testset.AssertSetEqual(t, expected, actual)
}

// func TestNestImpl(t *testing.T) {
// 	t.Parallel()

// 	s := New(
// 		[]string{"c", "a"},
// 		[]any{1, 10},
// 		[]any{1, 11},
// 		[]any{3, 11},
// 		[]any{3, 10},
// 	)
// 	// attrAttrs := NewMap(KV("aa", NewSet("a")))
// 	keyAttrs := NewSet("a")

// 	grouped := SetGroupBy(s, func(el Tuple) Tuple {
// 		return el.Without(keyAttrs)
// 	})
// 	t.Log("grouped =", grouped)

// 	mapped := MapMap(grouped, func(key Tuple, group Relation) Tuple {
// 		return NewMap(KV[string, any]("aa", Project(group, keyAttrs)))
// 	})
// 	t.Log("mapped  =", mapped)

// 	result := mapped.Values()
// 	t.Logf("result  = %+v", result)

// 	s = result
// 	keyAttrs = NewSet("c")

// 	grouped = SetGroupBy(s, func(el Tuple) Tuple {
// 		return el.Without(keyAttrs)
// 	})
// 	t.Log("grouped =", grouped)

// 	mapped = MapMap(grouped, func(key Tuple, group Relation) Tuple {
// 		a := NewMap(KV[string, any]("cc", NewSet(NewMap[string, any]())))
// 		return a //.Update(key)
// 	})
// 	t.Log("mapped  =", mapped)
// 	keys := mapped.Keys().Elements()
// 	t.Log("keys    =", keys)
// 	// t.Log("keys[0] == keys[1] =", keys[0].Equal(keys[1]))
// 	// t.Log("keys[0][aa] == keys[1][aa] =", keys[0].MustGet("aa").(Relation).Equal(keys[1].MustGet("aa").(Relation)))

// 	result = mapped.Values()
// 	t.Log("result  =", result)

// 	t.Fail()
// }
