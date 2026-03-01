package hash

type Seed = uintptr

// Hashable represents a type that can evaluate its own hash.
type Hashable interface {
	Hash(seed Seed) Seed
}
