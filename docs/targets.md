# Targets

<!-- last-evaluated: bf5e07b -->

## Active

(none)

## Achieved

### 🎯T4 Set and Map operations are correct across diverse key/value types
- **Weight**: 4 (value 8 / cost 2)
- **Estimated-cost**: 2
- **Acceptance**:
  - A parametric test suite exercises all core operations (With, Without, Has/Get, Union, Intersection, Difference, Equal) for Set and Map across a matrix of types: `int`, `string`, `float64`, `type ID int`, `type Name string`, `type Score float64`, `struct` types, pointer types
  - `Set[int]` and `Set[ID]` (where `type ID int`) never produce hash collisions — elements with the same underlying value but different types are independent
  - `Map[int, V]` and `Map[ID, V]` behave independently for same underlying key values
  - Derived-type sets/maps round-trip correctly through all operations (insert N elements, verify Has for all, verify Count, verify Without removes correctly)
  - All tests pass with `-race`
- **Context**: 17 parametric tests in `typematrix_test.go` cover Set and Map across int, string, float64, derived types (ID, Name, Score), and structs. Independence tests verify Set[int]/Set[ID] and Map[int,V]/Map[ID,V] function independently. Large round-trip test exercises 1000-element insert/verify/remove cycle for derived types. All pass with -race.
- **Status**: achieved
- **Discovered**: 2026-03-09
- **Achieved**: 2026-03-09

### 🎯T3 No-op write operations are zero-alloc
- **Weight**: 3 (value 5 / cost 2)
- **Estimated-cost**: 2
- **Acceptance**:
  - `Map.With` on existing key (same value) reports 0 allocs/op
  - `Map.Without` on absent key reports 0 allocs/op
  - `Set.With` on existing element reports 0 allocs/op
  - `Set.Without` on absent element reports 0 allocs/op
  - No regression in throughput
  - All existing tests pass
- **Context**: Added `same` field to `CombineArgs` for no-op detection. `Set.With` and `Map.With` now use `Tree.WithWith` with cached `CombineArgs` (non-boxing equality/hash via `EqArgs`) and identity checks via `same`. When an element is already present with the same value, leaf nodes return themselves unchanged, short-circuiting all spine allocation. Bonus: mutating inserts also improved (Map.Insert 1k: 12→3 allocs) because the cached EqArgs path avoids boxing through `value.Equal` and `hash.Any`.
- **Status**: achieved
- **Discovered**: 2026-03-08
- **Achieved**: 2026-03-08

### 🎯T2 All read operations are zero-alloc
- **Weight**: 5 (value 8 / cost 2)
- **Estimated-cost**: 2
- **Acceptance**:
  - `Map.Get`, `Map.Has`, `Set.Has` report 0 B/op, 0 allocs/op in benchmarks at 1k and 1M sizes
  - No regression in throughput (ns/op must not increase)
  - All existing tests pass
- **Context**: Used reflect.Kind dispatch in resolveHashFunc, resolveSeededHashFunc, and EqualFuncFor to catch derived types and route through direct hash/equality functions. Added size-based integer fast-paths to resolveSeededHashFunc.
- **Status**: achieved
- **Discovered**: 2026-03-08
- **Achieved**: 2026-03-08

### 🎯T1 Map[K,V] hot-path operations avoid per-call allocations
- **Weight**: 5 (value 8 / cost 2)
- **Estimated-cost**: 2
- **Acceptance**:
  - `Map.Get` and `Map.Without` use cached `EqArgs` instead of allocating per call
  - `mapEntryHashFunc` hashes keys directly without boxing through `Hashable` or `hash.Any`
  - `string` type has fast-path specializations in `resolveHashFunc`, `resolveSeededHashFunc`, and `EqualFuncFor`
  - All existing tests pass (`make test`)
  - Benchmarks show improvement for `Map[string,V]` Get/Without operations
- **Context**: Map.Get and Map.Without are high-frequency operations. Each call was allocating a new EqArgs and boxing keys through interfaces. Caching EqArgs per type and using direct hash functions eliminates this overhead.
- **Status**: achieved
- **Discovered**: 2026-03-05
- **Achieved**: 2026-03-08
