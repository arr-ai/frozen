# Targets

<!-- last-evaluated: bf5e07b -->

## Active

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
- **Context**: No-op writes currently allocate because they take the same code path as mutating writes before discovering no change is needed. The tree already returns early (same node pointer), but allocations may occur in the hasher/EqArgs path before the early return.
- **Status**: identified
- **Discovered**: 2026-03-08

## Achieved

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
