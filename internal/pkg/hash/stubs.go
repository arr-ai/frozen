// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"unsafe"

	"sync/atomic"
)

// Should be a built-in for unsafe.Pointer?
//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

var fastrandState uint64

//go:nosplit
func fastrand() uint32 {
	var s1, s0 uint32
	for i := 0; i < 3; i++ {
		state := atomic.LoadUint64(&fastrandState)
		frand := *(*[]uint32)(noescape(unsafe.Pointer(&state)))
		// Implement xorshift64+: 2 32-bit xorshift sequences added together.
		// Shift triplet [17,7,16] was calculated as indicated in Marsaglia's
		// Xorshift paper: https://www.jstatsoft.org/article/view/v008i14/xorshift.pdf
		// This generator passes the SmallCrush suite, part of TestU01 framework:
		// http://simul.iro.umontreal.ca/testu01/tu01.html
		s1, s0 = frand[0], frand[1]
		s1 ^= s1 << 17
		s1 = s1 ^ s0 ^ s1>>7 ^ s0>>16
		frand[0], frand[1] = s0, s1
		newState := *(*uint64)(noescape(unsafe.Pointer(&frand)))
		if atomic.CompareAndSwapUint64(&fastrandState, state, newState) {
			break
		}
	}
	return s0 + s1
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0) //nolint:staticcheck
}
