package hash

import (
	"reflect"
	"testing"
)

type (
	id      int
	otherID int
	emptyA  struct{}
	emptyB  struct{}
	name    string
)

func TestAnyDistinguishesDefinedTypes(t *testing.T) {
	t.Parallel()
	if Any(int(5)) == Any(id(5)) {
		t.Error("int(5) and id(5) hashed alike")
	}
	if Any(id(5)) == Any(otherID(5)) {
		t.Error("id(5) and otherID(5) hashed alike")
	}
	a, b := Any(id(5)), Any(id(5))
	if a != b {
		t.Error("id(5) hashed inconsistently")
	}
	if Any("x") == Any(name("x")) {
		t.Error("string and name hashed alike")
	}
	if Any(emptyA{}) == Any(emptyB{}) {
		t.Error("distinct empty named structs hashed alike")
	}
}

func TestValueMixesTypeName(t *testing.T) {
	t.Parallel()
	h := Value(reflect.ValueOf(id(5)))
	// The defined type must not hash like the underlying bits alone.
	if h == Int(5) {
		t.Error("Value(id(5)) hashed like Int(5)")
	}
	if h != Value(reflect.ValueOf(id(5))) {
		t.Error("Value(id(5)) hashed inconsistently")
	}
}
