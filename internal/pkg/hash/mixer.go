package hash

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

// Seedless scalar hashing.
//
// Scalars and byte sequences of at most shortBytes are hashed with a
// single 64x64→128-bit multiply whose two halves are folded together: the
// "mum" primitive behind wyhash and rapidhash. It is a handful of
// instructions that inline at every call site, which makes it several
// times faster than a call into the AES assembly for keys that fit in a
// register. The high half of the product depends on every input bit, so
// the top bits of the result — the ones a hashed array trie addresses
// from first — are well mixed for keys whose entropy sits in their low
// bits (sequential integers) and for keys whose entropy sits in their
// high bits alike.
//
// Longer strings and []byte values still go through the AES memhash path,
// where SIMD over the remaining bytes pays for the call. mixKey is drawn
// from crypto/rand at process start, so hash values are neither stable
// across runs nor predictable without the key; the AES path is keyed the
// same way.

const mixMul = 0x9E3779B97F4A7C15 // 2^64 / φ, odd

var mixKey = func() uint64 {
	var b [8]byte
	getRandomData(b[:])
	return binary.LittleEndian.Uint64(b[:])
}()

func mix64(x uint64) uint64 {
	hi, lo := bits.Mul64(x^mixKey, mixMul)
	return hi ^ lo
}

func asUintptr(x uint64) uintptr {
	return uintptr(x & uint64(^uintptr(0)))
}

// Combine folds o into h asymmetrically. Use it to hash ordered or keyed
// structures: Combine(a, b) != Combine(b, a) in general, and
// Combine(Combine(s, a), b) differs from Combine(Combine(s, b), a).
// Unordered collections should XOR their element hashes instead and pass
// the result through a scalar hash once at the end.
func Combine(h, o uintptr) uintptr {
	return uintptr(mix64(uint64(h) ^ (uint64(o)*mixMul + mixSalt)))
}

const mixSalt = 0x2545F4914F6CDD1D

// shortBytes is the largest length at which two overlapping 8-byte reads
// cover every byte (the wyhash 9–16 trick). Above it, AES is the better
// tool. Distinct salts keep "" from hashing like []byte{} and keep both
// from hashing like a scalar zero.
const (
	shortBytes = 16
	strSalt    = 0xBE54646C23A5B365
	byteSalt   = 0x71D67FFFEDA9D9D5
)

func mixBytes(p unsafe.Pointer, n uintptr, salt uint64) uintptr {
	switch {
	case n == 0:
		return uintptr(mix64(salt))
	case n < 4:
		a := uint64(*(*byte)(p))
		b := uint64(*(*byte)(add(p, n>>1)))
		c := uint64(*(*byte)(add(p, n-1)))
		return uintptr(mix64((a | b<<8 | c<<16 | uint64(n)<<32) ^ salt))
	case n <= 8:
		lo := uint64(readUnaligned32(p))
		hi := uint64(readUnaligned32(add(p, n-4)))
		return uintptr(mix64((lo | hi<<32) ^ uint64(n)*mixMul ^ salt))
	default: // 9–16; caller guarantees n <= shortBytes
		a := readUnaligned64(p)
		b := readUnaligned64(add(p, n-8))
		hi, lo := bits.Mul64(a^mixKey, b^uint64(n)^salt)
		return uintptr(hi ^ lo)
	}
}
