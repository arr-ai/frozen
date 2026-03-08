# Targets

<!-- last-evaluated: 0ab92d9 -->

## Active

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
- **Status**: converging
- **Discovered**: 2026-03-05

## Achieved
