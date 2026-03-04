package tree

import (
	"fmt"
	"reflect"
	"sync"
)

type hashFuncRef struct {
	typ uintptr
	ptr uintptr
}

var customHashFuncRegistry sync.Map

func newHashFuncRef[T any](hf func(T) H128) hashFuncRef {
	if hf == nil {
		return hashFuncRef{}
	}
	ptr := reflect.ValueOf(hf).Pointer()
	if ptr == reflect.ValueOf(getHashFunc[T]()).Pointer() {
		return hashFuncRef{}
	}
	ref := hashFuncRef{typ: typeKey[T](), ptr: ptr}
	customHashFuncRegistry.Store(ref, hf)
	return ref
}

func mergeHashFuncRef(a, b hashFuncRef) hashFuncRef {
	if a.ptr != 0 {
		return a
	}
	if b.ptr != 0 {
		return b
	}
	return hashFuncRef{}
}

func resolveHashFuncRef[T any](r hashFuncRef) func(T) H128 {
	if r.ptr == 0 {
		return getHashFunc[T]()
	}
	hf, ok := customHashFuncRegistry.Load(r)
	if !ok {
		panic(fmt.Errorf("missing hash function registry entry: type=%d ptr=%d", r.typ, r.ptr))
	}
	return hf.(func(T) H128) //nolint:forcetypeassert
}