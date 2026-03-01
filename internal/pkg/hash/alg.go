// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	c0 = (8-unsafe.Sizeof(uintptr(0)))/4*2860486313 +
		(unsafe.Sizeof(uintptr(0))-4)/4*(33054211828000289&(1<<unsafe.Sizeof(uintptr(0))-1))
	c1 = (8-unsafe.Sizeof(uintptr(0)))/4*3267000013 +
		(unsafe.Sizeof(uintptr(0))-4)/4*(23344194077549503&(1<<unsafe.Sizeof(uintptr(0))-1))
)

// type algorithms - known to compiler
const (
	algNOEQ = iota
	algMEM0
	algMEM8
	algMEM16
	algMEM32
	algMEM64
	algMEM128
	algSTRING
	algFLOAT32
	algFLOAT64
	algCPLX64
	algCPLX128
	algMAX
	algPTR  = algMEM32 + (algMEM64-algMEM32)*(unsafe.Sizeof(uintptr(0))/8)
	algINT  = algMEM32 + (algMEM64-algMEM32)*(unsafe.Sizeof(0)/8)
	algUINT = algMEM32 + (algMEM64-algMEM32)*(unsafe.Sizeof(uint(0))/8)
)

type hashFunc func(unsafe.Pointer, uintptr) uintptr

func memhash0(p unsafe.Pointer, h uintptr) uintptr {
	return h
}

func memhash8(p unsafe.Pointer, h uintptr) uintptr {
	return memhash(p, h, 1)
}

func memhash16(p unsafe.Pointer, h uintptr) uintptr {
	return memhash(p, h, 2)
}

func memhash128(p unsafe.Pointer, h uintptr) uintptr {
	return memhash(p, h, 16)
}

var algarray = [algMAX]hashFunc{
	algNOEQ:    nil,
	algMEM0:    memhash0,
	algMEM8:    memhash8,
	algMEM16:   memhash16,
	algMEM32:   memhash32,
	algMEM64:   memhash64,
	algMEM128:  memhash128,
	algSTRING:  strhash,
	algFLOAT32: f32hash,
	algFLOAT64: f64hash,
	algCPLX64:  c64hash,
	algCPLX128: c128hash,
}

var useAeshash bool

// in asm_*.s
func aeshash(p unsafe.Pointer, h, s uintptr) uintptr
func aeshash32(p unsafe.Pointer, h uintptr) uintptr
func aeshash64(p unsafe.Pointer, h uintptr) uintptr
func aeshashstr(p unsafe.Pointer, h uintptr) uintptr

type stringStruct struct {
	str unsafe.Pointer
	len int
}

func strhash(a unsafe.Pointer, h uintptr) uintptr {
	x := (*stringStruct)(a)
	return memhash(x.str, h, uintptr(x.len))
}

// NOTE: Because NaN != NaN, a map can contain any
// number of (mostly useless) entries keyed with NaNs.
// To avoid long hash chains, we assign a random number
// as the hash value for a NaN.

func f32hash(p unsafe.Pointer, h uintptr) uintptr {
	f := *(*float32)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(fastrand())) // any kind of NaN
	default:
		return memhash(p, h, 4)
	}
}

func f64hash(p unsafe.Pointer, h uintptr) uintptr {
	f := *(*float64)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(fastrand())) // any kind of NaN
	default:
		return memhash(p, h, 8)
	}
}

func c64hash(p unsafe.Pointer, h uintptr) uintptr {
	x := (*[2]float32)(p)
	return f32hash(unsafe.Pointer(&x[1]), f32hash(unsafe.Pointer(&x[0]), h))
}

func c128hash(p unsafe.Pointer, h uintptr) uintptr {
	x := (*[2]float64)(p)
	return f64hash(unsafe.Pointer(&x[1]), f64hash(unsafe.Pointer(&x[0]), h))
}

const hashRandomBytes = unsafe.Sizeof(uintptr(0)) / 4 * 64

//go:nosplit
func getRandomData(r []byte) {
	if _, err := rand.Read(r); err != nil {
		panic(err)
	}
}

// used in asm_{386,amd64,arm64}.s to seed the hash function
var aeskeysched [hashRandomBytes]byte

// used in hash{32,64}.go to seed the hash function
var hashkey [4]uintptr

// GetSeeds returns the internal seeds used for computing hashes.
func GetSeeds() (a []byte, h []uintptr) {
	if useAeshash {
		return aeskeysched[:], nil
	}
	return nil, hashkey[:]
}

// SetSeeds sets the internal seeds used for computing hashes.
func SetSeeds(a []byte, h []uintptr) error {
	if useAeshash {
		if len(a) != len(aeskeysched) {
			return fmt.Errorf("SetSeeds: aeskeysched: %d != %d", len(a), len(aeskeysched)) //nolint:golint
		}
		copy(aeskeysched[:], a)
		return nil
	}
	if len(h) != len(hashkey) {
		return fmt.Errorf("SetSeeds: hashkey: %d != %d", len(h), len(hashkey)) //nolint:golint
	}
	copy(hashkey[:], h)
	return nil
}

var _ = func() (_ struct{}) {
	// Install AES hash algorithms if the instructions needed are present.
	if (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") && runtime.GOOS != "nacl" &&
		cpu.X86.HasAES && // AESENC
		cpu.X86.HasSSSE3 && // PSHUFB
		cpu.X86.HasSSE41 { // PINSR{D,Q}
		initAlgAES()
	} else if runtime.GOARCH == "arm64" && cpu.ARM64.HasAES {
		initAlgAES()
	} else {
		getRandomData((*[len(hashkey) * int(unsafe.Sizeof(uintptr(0)))]byte)(unsafe.Pointer(&hashkey))[:])
		hashkey[0] |= 1 // make sure these numbers are odd
		hashkey[1] |= 1
		hashkey[2] |= 1
		hashkey[3] |= 1
		initH128Software()
	}
	return
}()

func initAlgAES() {
	if runtime.GOOS == "aix" {
		// runtime.algarray is immutable on AIX: see cmd/link/internal/ld/xcoff.go
		return
	}
	useAeshash = true
	algarray[algMEM32] = aeshash32
	algarray[algMEM64] = aeshash64
	algarray[algSTRING] = aeshashstr
	// Initialize with random data so hash collisions will be hard to engineer.
	getRandomData(aeskeysched[:])
	// H128 assembly is only implemented for amd64 and arm64.
	// On 386, fall back to software H128 (two seeded calls).
	if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
		initH128AES()
	} else {
		initH128Software()
	}
}

// Note: These routines perform the read with an native endianness.
func readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	if bigEndian {
		return uint32(q[3]) | uint32(q[2])<<8 | uint32(q[1])<<16 | uint32(q[0])<<24
	}
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

func readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	if bigEndian {
		return uint64(q[7]) | uint64(q[6])<<8 | uint64(q[5])<<16 | uint64(q[4])<<24 |
			uint64(q[3])<<32 | uint64(q[2])<<40 | uint64(q[1])<<48 | uint64(q[0])<<56
	}
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 |
		uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}
