package frozen

import (
	"math"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestHash64(t *testing.T) {
	if hash(uint64(0)) == 0 {
		t.Error()
	}
}

func TestHash64String(t *testing.T) {
	if hash("hello") == 0 {
		t.Error()
	}
}

func TestHashMatchesEquality(t *testing.T) {
	t.Logf("%d unique elements", len(cornucopia))
	total := 0
	falsePositives := 0
	for _, a := range cornucopia {
		for _, b := range cornucopia {
			if a == b {
				assert.Equal(t, hash(a), hash(b), "a=%v b=%v hash(a)=%v hash(b)=%v", a, b, hash(a), hash(b))
			} else if hash(a) == hash(b) {
				t.Logf("hash(%#v %[1]T) == hash(%#v %[2]T) == %d", a, b, hash(a))
				falsePositives++
			}
			total++
		}
	}
	assert.LessOrEqual(t, falsePositives, total/1_000_000, total)
}

func BenchmarkHash(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < b.N; i++ {
		hash(cornucopia[r.Int()%len(cornucopia)])
	}
}

var cornucopia = func() []interface{} {
	x := 42
	result := []interface{}{
		false,
		true,
		&x,
		&[]int{43}[0],
		&[]string{"hello"}[0],
		uintptr(unsafe.Pointer(&x)),
		unsafe.Pointer(nil),
		unsafe.Pointer(&x),
		unsafe.Pointer(uintptr(unsafe.Pointer(&x))),
		[...]int{},
		[...]int{1, 2, 3, 4, 5},
		[...]int{5, 4, 3, 2, 1},
	}

	// The following number lists are massive overkill, but it can't hurt.

	type intFoo int
	type int8Foo int8
	type int16Foo int16
	type int32Foo int32
	type int64Foo int64
	type intBar int
	type int8Bar int8
	type int16Bar int16
	type int32Bar int32
	type int64Bar int64

	for _, i := range []int64{
		-43, -42, -10, -1, 0, 1, 10, 42,
		math.MaxInt64, math.MaxInt64 - 1,
		math.MinInt64, math.MinInt64 + 1,
	} {
		// result = append(result, int(i), int8(i), int16(i), int32(i), int64(i))
		result = append(result, intFoo(i), int8Foo(i), int16Foo(i), int32Foo(i), int64Foo(i))
		result = append(result, intBar(i), int8Bar(i), int16Bar(i), int32Bar(i), int64Bar(i))
	}

	type uintFoo int
	type uint8Foo int8
	type uint16Foo int16
	type uint32Foo int32
	type uint64Foo int64
	type uintptrFoo int64
	type uintBar int
	type uint8Bar int8
	type uint16Bar int16
	type uint32Bar int32
	type uint64Bar int64
	type uintptrBar int64

	for _, i := range []uint64{0, 42} {
		result = append(result, uint(i), uint8(i), uint16(i), uint32(i), i)
		result = append(result, uintFoo(i), uint8Foo(i), uint16Foo(i), uint32Foo(i), uint64Foo(i))
		result = append(result, uintptrFoo(i))
		result = append(result, uintBar(i), uint8Bar(i), uint16Bar(i), uint32Bar(i), uint64Bar(i))
		result = append(result, uintptrBar(i))
	}

	type float32Foo float32
	type float64Foo float64
	type float32Bar float32
	type float64Bar float64

	floats := []float64{
		0, 42, math.Pi,
		math.MaxFloat32, math.SmallestNonzeroFloat32,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
	}

	for _, f := range floats {
		result = append(result, float32(f), f)
		result = append(result, float32Foo(f), float64Foo(f))
		result = append(result, float32Bar(f), float64Bar(f))
	}

	type complex64Foo complex64
	type complex128Foo complex128
	type complex64Bar complex64
	type complex128Bar complex128

	c64 := func(re, im float64) complex64 { return complex(float32(re), float32(im)) }
	c128 := func(re, im float64) complex128 { return complex(re, im) }
	for _, re := range floats {
		for _, im := range floats {
			result = append(result, c64(re, im), c128(re, im))
			result = append(result, complex64Foo(c64(re, im)), complex128Foo(c128(re, im)))
			result = append(result, complex64Bar(c64(re, im)), complex128Bar(c128(re, im)))
		}
	}

	type stringFoo string
	type stringBar string

	for _, s := range []string{
		"",
		"a",
		"b",
		"hello",
		"-------------------------------------------------------",
		"--------------------------------------------------------",
		"--------------------------------------------------------\000",
	} {
		result = append(result, s)
		result = append(result, stringFoo(s))
		result = append(result, stringBar(s))
	}

	// Dedupe
	m := map[interface{}]struct{}{}
	for _, i := range result {
		m[i] = struct{}{}
	}

	for i := range m {
		result = append(result, i)
	}

	return result
}()
