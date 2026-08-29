package frozen

import (
	"sync"

	"github.com/arr-ai/hash/hash128"

	"github.com/arr-ai/frozen/internal/pkg/hash"
	"github.com/arr-ai/frozen/internal/pkg/tree"
	"github.com/arr-ai/frozen/internal/pkg/value"
)

type mapEntry[K, V any] struct {
	KeyValue[K, V]
}

func newMapEntry[K, V any](k K, v V) mapEntry[K, V] {
	return mapEntry[K, V]{KeyValue: KeyValue[K, V]{Key: k, Value: v}}
}

func newMapKey[K, V any](k K) mapEntry[K, V] {
	var v V
	return mapEntry[K, V]{KeyValue: KeyValue[K, V]{Key: k, Value: v}}
}

// Equal implements value.Equaler for key-only comparison in the default path
// (Tree.Get, Tree.With, Tree.Without).
func (e mapEntry[K, V]) Equal(e2 mapEntry[K, V]) bool {
	return value.Equal(e.Key, e2.Key)
}

// Hash implements hash.Hashable for key-only hashing.
func (e mapEntry[K, V]) Hash(seed uintptr) uintptr {
	return hash.Any(e.Key, seed)
}

// Hash128 implements hash128.Hashable for key-only hashing. It takes
// precedence over Hash wherever entries are hashed, so every path that hashes
// a mapEntry agrees with mapEntryHashFunc.
func (e mapEntry[K, V]) Hash128() hash128.H128 {
	return tree.GetHashFunc[K]()(e.Key).ToHash128()
}

// mapEntryEqHash provides full entry equality (key + value) for Map.Equal and similar.
type mapEntryEqHash[K, V any] struct {
	eqK  func(K, K) bool
	eqV  func(V, V) bool
	hash func(mapEntry[K, V]) tree.H128
}

func (m *mapEntryEqHash[K, V]) Equal(a, b mapEntry[K, V]) bool {
	return m.eqK(a.Key, b.Key) && m.eqV(a.Value, b.Value)
}

func (m *mapEntryEqHash[K, V]) Hash(a mapEntry[K, V]) tree.H128 {
	return m.hash(a)
}

func (m *mapEntryEqHash[K, V]) FullHash() bool { return false }

// mapKeyEqHash provides key-only equality for Map operations (With, Without, etc.).
type mapKeyEqHash[K, V any] struct {
	eqK  func(K, K) bool
	hash func(mapEntry[K, V]) tree.H128
}

func (m *mapKeyEqHash[K, V]) Equal(a, b mapEntry[K, V]) bool {
	return m.eqK(a.Key, b.Key)
}

func (m *mapKeyEqHash[K, V]) Hash(a mapEntry[K, V]) tree.H128 {
	return m.hash(a)
}

func (m *mapKeyEqHash[K, V]) FullHash() bool { return false }

// mapEntryHashFunc returns a non-boxing H128 hash function for mapEntry[K, V].
// It hashes only the key, consistent with mapEntry.Hash128, but avoids boxing
// the mapEntry struct through the Hashable interface. The key is hashed in a
// single pass by the key type's own H128 function.
func mapEntryHashFunc[K, V any]() func(mapEntry[K, V]) tree.H128 {
	kHash := tree.GetHashFunc[K]()
	return func(e mapEntry[K, V]) tree.H128 {
		return kHash(e.Key)
	}
}

var mapEntryEqHashCache sync.Map

func getMapEntryEqHash[K, V any]() *mapEntryEqHash[K, V] {
	key := tree.TypeKeyOf[mapEntry[K, V]]()
	if f, ok := mapEntryEqHashCache.Load(key); ok {
		return f.(*mapEntryEqHash[K, V]) //nolint:forcetypeassert
	}
	ops := &mapEntryEqHash[K, V]{
		eqK:  value.EqualFuncFor[K](),
		eqV:  value.EqualFuncFor[V](),
		hash: mapEntryHashFunc[K, V](),
	}
	mapEntryEqHashCache.Store(key, ops)
	return ops
}

var mapKeyEqHashCache sync.Map

func getMapKeyEqHash[K, V any]() *mapKeyEqHash[K, V] {
	key := tree.TypeKeyOf[mapEntry[K, V]]()
	if f, ok := mapKeyEqHashCache.Load(key); ok {
		return f.(*mapKeyEqHash[K, V]) //nolint:forcetypeassert
	}
	ops := &mapKeyEqHash[K, V]{
		eqK:  value.EqualFuncFor[K](),
		hash: mapEntryHashFunc[K, V](),
	}
	mapKeyEqHashCache.Store(key, ops)
	return ops
}
