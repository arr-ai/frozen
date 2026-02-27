# Optimizing a Go HAMT: Research Findings

## Abstract

This paper documents a series of optimizations to the hash array mapped trie
(HAMT) implementation in `github.com/arr-ai/frozen`, a Go library for immutable
data structures. Starting from a node-type simplification that introduced
performance regressions of up to 56%, we investigated and implemented three
optimizations that recovered and exceeded baseline performance on most
benchmarks. We also explored fanout tuning (4, 8, 16, 32-way branching),
variable-length node arrays, and cache line alignment, finding that the existing
8-way fanout is near-optimal for Go's memory model. The key insight is that Go's
interface-based polymorphism, garbage collector behavior, and cache architecture
create a unique optimization landscape where reducing allocation count matters
far more than reducing allocation size.

## 1. Background

### 1.1 The frozen Library

`frozen` provides immutable `Set[T]`, `Map[K,V]`, and `IntSet[I]` types backed
by a HAMT with structural sharing. All mutations produce new values; unmodified
subtrees are shared between old and new versions. The library uses Go 1.18+
generics.

### 1.2 HAMT Structure

The trie uses 3-bit hash chunks (fanout 8) with multi-round hashing: when all
bits of a 64-bit hash are exhausted after 21 levels, a new hash is computed with
a different seed. Node types:

- **`branch[T]`**: Internal node with a 16-bit occupancy mask and `[8]node[T]`
  child array, plus an element count.
- **`leaf1[T]`**: Single-element leaf.
- **`leaf[T]`**: Multi-element leaf (2-8 elements) with linear scan, deferring
  branch creation until a 9th element arrives.

The `node[T]` interface (Go interface = 16-byte type+data pointer pair) provides
polymorphic dispatch across node types.

### 1.3 The Triggering Change

A simplification removed two node types (`leaf2`, `twig`) in favor of
multi-round hashing and the unified `leaf` type. This eliminated ~500 lines of
code and many type-switch cases but introduced regressions:

| Benchmark | Regression |
|-----------|-----------|
| MergeFrozenMap 1M | +56% |
| SetBuilder-Add 1Mi | +25% |
| Set-With 1ki | +47% |

## 2. Implemented Optimizations

### 2.1 Boxing Elimination

**Problem.** Go's generic type parameters are compiled via dictionary-based
dispatch. Functions like `value.Equal[T](a, b)` and `hash.Any(key, seed)`
receive values as `any` (empty interface), causing heap allocation ("boxing")
for every comparison and hash computation.

**Solution.** Per-type function resolution cached in `sync.Map`, keyed by
`reflect.TypeOf(&t)`. On first call for a given type `T`, the resolver inspects
`T` and returns a specialized non-boxing function:

```go
func resolveEqualFunc[T any]() func(a, b T) bool {
    var t T
    switch any(t).(type) {
    case float32: return func(a, b T) bool { /* direct float comparison */ }
    // ... size-based specializations using unsafe.Pointer
    }
    return func(a, b T) bool { return value.Equal(a, b) } // fallback
}
```

Hash functions use the same pattern, dispatching to `hash.Uint32`, `hash.Uint64`,
etc. based on `unsafe.Sizeof(t)`.

**Impact.** At 1Mi elements, SetBuilder allocations dropped 75% and bytes/op
dropped 46%. MergeFrozenMap allocations dropped 80%.

### 2.2 sync.Map Lookup Elimination on Builder Hot Path

**Problem.** After 2.1, CPU profiling of `SetBuilder.Add` at 1ki elements
revealed `sync.Map.Load` consuming 16.75% of CPU. Each `Add` call performed two
`sync.Map` lookups: one for the hash function (via `getHashFunc[T]()`) and one
for the equality function (via `value.Equal`).

**Solution.** Pre-resolve both functions at Builder construction time into the
`CombineArgs[T]` struct, then route all Builder operations through the
args-based path:

```go
type Builder[T any] struct {
    t    Tree[T]
    args *CombineArgs[T]  // pre-resolved eq + hash functions
}

func (b *Builder[T]) Add(v T) {
    if b.args == nil {
        b.args = DefaultNPKeyCombineArgs[T]()  // cached per type
    }
    b.add(b.args, v)
}
```

A `newHasherWith` variant accepts the pre-resolved hash function directly,
bypassing `getHashFunc[T]()`. The `CombineArgs` themselves are cached per type
in a `sync.Map` to avoid per-Builder allocation overhead.

**Impact.** SetBuilder-Add at 1ki improved from +45% regressed to -6% vs
baseline. The IntSet/New benchmark, which briefly regressed +157% due to
per-Builder allocation of args, was fixed by caching args per type.

### 2.3 Inline Branch Copies

**Problem.** CPU profiling of `Set.With` showed 46% of time in GC. The
immutable `WithFast` method used:

```go
ret := newBranch(b.p.WithChild(i, x2))
```

`packer.WithChild` allocates a copy of the packer on the heap (returned as
`*packer[T]`), then `newBranch` allocates a branch and copies the packer into
it. Two heap allocations per branch level.

**Solution.** Copy the branch directly and mutate the copy:

```go
ret := *b              // single allocation (escapes to heap)
ret.p.data[i] = x2     // mask bit already set for existing children
ret.count = b.count + 1 - matches
return &ret, matches
```

For new children (nil slot), `SetNonNilChild` sets the mask bit. For removals
(`Without`), `SetChild` handles mask clearing when the child becomes nil.

**Impact.** Set-With at 32 elements improved 12%, SetBuilder-Add at 32 improved
17%. The `Remove` method already used this pattern; we extended it to `with`,
`WithFast`, and `Without`.

### 2.4 Cumulative Results

After all three optimizations, measured against the pre-simplification baseline:

| Benchmark | Time | B/op | Allocs |
|-----------|------|------|--------|
| SetBuilder-Add 32 | **-31%** | -38% | -34% |
| SetBuilder-Add 1ki | **-4.5%** | -24% | -46% |
| MergeFrozenMap 1M | **-38%** | -62% | -75% |
| SetUnion 1M | **-46%** | -37% | ~ |
| SetParallelWith 1M | **-17%** | ~ | ~ |
| Set-With 32 | +30% | -4% | -9% |
| Set-With 1ki | +22% | -4% | -1% |
| Set-With 1Mi | +14% | ~ | ~ |
| IntSet/With 100 | +62% | +21% | ~ |

The merge and builder paths significantly exceed the original baseline. Set-With
remains regressed, driven by structural differences in the simplified node
types (primarily the `splitLeaf` path in the unified leaf).

## 3. Fanout Exploration

### 3.1 Hypothesis

Wider branch nodes reduce tree depth, meaning fewer branch copies per `With`
operation. Narrower nodes reduce copy size. The optimal fanout balances these
factors.

### 3.2 Branch Occupancy Analysis

We instrumented the tree to measure branch occupancy (populated slots / total
slots) by depth:

**Fanout 8, N=1,000,000:**

| Depth | Branches | Avg Occupancy | Occupancy % |
|-------|----------|---------------|-------------|
| 0 | 1 | 8.00 | 100% |
| 1 | 8 | 8.00 | 100% |
| 2 | 64 | 8.00 | 100% |
| 3 | 512 | 8.00 | 100% |
| 4 | 4,096 | 8.00 | 100% |
| 5 | 32,768 | 7.82 | 98% |
| 6 | 4,273 | 5.73 | 72% |
| **Total** | **41,722** | **7.63** | **95%** |

Only the deepest level shows significant sparsity (72%). All other levels are
95-100% full. The tree is remarkably dense.

**Fanout 32, N=1,000,000:**

| Depth | Branches | Avg Occupancy | Occupancy % |
|-------|----------|---------------|-------------|
| 0 | 1 | 16.00 | 50% |
| 1 | 32 | 16.00 | 50% |
| 2 | 1,024 | 16.00 | 50% |
| 3 | 32,768 | 9.85 | 31% |
| **Total** | **33,825** | **10.04** | **31%** |

At fanout 32, branches are only 31% occupied. Each 528-byte branch copy carries
~360 bytes of nil pointers.

### 3.3 Benchmark Results

All fanouts benchmarked against fanout 8 on Apple M4 Max:

**Set-With (the primary regression target):**

| Fanout | 32 elements | 1ki elements | 1Mi elements |
|--------|-------------|--------------|--------------|
| 4 | +10% | **-19%** | +17% |
| 8 | baseline | baseline | baseline |
| 16 | **-16%** | +62% | ~ |
| 32 | **-20%** | **-35%** | **-17%** |

**SetBuilder-Add:**

| Fanout | 32 elements | 1ki elements | 1Mi elements |
|--------|-------------|--------------|--------------|
| 4 | +50% | ~ | +47% |
| 8 | baseline | baseline | baseline |
| 16 | +12% | **-33%** | +20% |
| 32 | +49% | **-34%** | **-26%** |

**MergeFrozenMap 1M:**

| Fanout | Time vs F=8 |
|--------|------------|
| 4 | ~ |
| 8 | baseline |
| 16 | ~ |
| 32 | **+154%** |

**Key observations:**

1. **No fanout wins universally.** Wider fanout helps sequential depth-bound
   operations (Set-With, Builder at scale) but catastrophically hurts
   merge-heavy operations at fanout 32 (+154% on MergeFrozenMap).

2. **The occupancy data explains why.** At fanout 8, branches are 95% full —
   the fixed `[8]node` array wastes almost nothing. At fanout 32, 69% of every
   branch copy is nil pointers.

3. **Fanout 16 showed erratic behavior,** with a +62% regression on Set-With at
   1ki that contradicts both neighbors, likely due to cache/alignment effects at
   the 272-byte branch size.

### 3.4 Cache Line Analysis

Modern CPUs operate on cache lines, not individual bytes. Branch node sizes
relative to cache lines:

| Fanout | Branch size | 64B cache lines | 128B cache lines (Apple M-series) |
|--------|------------|-----------------|----------------------------------|
| 4 | 80 B | 2 | **1** |
| 8 | 144 B | 3 | 2 |
| 16 | 272 B | 5 | 3 |
| 32 | 528 B | 9 | 5 |

Total cache line touches per `With` operation at 1M elements:

| Fanout | Tree depth | Cache lines per copy | Total cache line touches |
|--------|-----------|---------------------|------------------------|
| 4 | ~10 | 1 | ~10 |
| 8 | ~6.7 | 2 | ~13 |
| 16 | ~5 | 3 | ~15 |
| 32 | ~4 | 5 | ~20 |

Fanout 4 minimizes cache line touches but maximizes allocation count. The
-19% Set-With improvement at 1ki (where depth is only ~5 levels, not ~10)
confirms the cache line effect, but at 1Mi the allocation count dominates
(+17% regression).

### 3.5 Variable-Length Node Arrays

Traditional HAMTs use compact arrays indexed by popcount, storing only populated
children. We analyzed whether this would help given the occupancy data.

At 95% average occupancy (fanout 8), compact arrays would save ~0.4 slots per
branch on average. In Go, the compact approach requires a slice (`[]node[T]`)
adding 24 bytes of header overhead and a separate heap allocation for the
backing array. Every immutable `With` becomes two allocations (branch + slice)
instead of one.

**Conclusion:** Compact arrays would add allocation overhead to save negligible
space. The fixed `[8]node` array is well-matched to the actual occupancy
profile. This optimization is a dead end for this implementation.

## 4. Discussion

### 4.1 Go-Specific Constraints

Several findings are specific to Go's runtime characteristics:

1. **Interface overhead.** Each `node[T]` in the child array is a Go interface
   (16 bytes: type pointer + data pointer). This is 2x larger than a raw pointer
   in C/Rust, making branch nodes inherently larger and favoring lower fanout.

2. **GC sensitivity.** Go's garbage collector scans all pointer-containing
   objects. Branch nodes with interface arrays are expensive to scan. Reducing
   allocation count (fewer branches per operation) reduces GC pressure more
   effectively than reducing allocation size.

3. **Escape analysis limitations.** Go's escape analysis often forces stack
   variables to the heap when they are returned by pointer or stored in
   interfaces. The `ret := *b; return &ret` pattern always heap-allocates
   because the compiler cannot prove `ret` doesn't escape.

4. **`sync.Map` as a generic cache.** Go lacks static dispatch for generics.
   Per-type function caches using `sync.Map` are necessary but add overhead on
   hot paths. Pre-resolving at construction time eliminates this cost.

### 4.2 Allocation Count vs Allocation Size

The dominant finding is that **allocation count matters far more than allocation
size** for this workload:

- Boxing elimination removed allocations entirely (not just resized them):
  -75% allocs, -46% bytes, -25% time on merge paths.
- Inline branch copies halved allocations per branch level (from 2 to 1):
  -12% time on Set-With despite identical total bytes.
- Fanout 32 reduced allocation count by ~40% but increased size by 3.7x:
  mixed results, with merge paths catastrophically worse due to GC scanning
  of larger objects.

This suggests that future optimization efforts should focus on reducing
allocation count rather than compacting data structures.

### 4.3 The Fanout Sweet Spot

The 8-way fanout (3-bit hash chunks) emerges as a Goldilocks value for Go HAMTs:

- **Dense enough** that fixed arrays waste little space (95% occupancy).
- **Shallow enough** that allocation count per operation is manageable (~7
  levels at 1M elements).
- **Small enough** that branch copies (144 bytes, 2 Apple cache lines) are
  efficient.
- **Power-of-2** so hash extraction is a simple bit shift.

Neither narrower (fanout 4: too deep, too many allocations) nor wider (fanout
16/32: too sparse, too much copy bandwidth, GC-hostile) improves the overall
picture.

## 5. Summary of Outcomes

### Implemented (committed)

| Optimization | Key Mechanism | Best Improvement |
|-------------|---------------|-----------------|
| Boxing elimination | sync.Map-cached per-type function dispatch | -75% allocs on merge |
| sync.Map bypass | Pre-resolved functions in Builder args | -45% allocs on builder |
| Inline branch copies | `ret := *b` instead of `newBranch(p.WithChild(...))` | -12% time on Set-With |

### Investigated (not adopted)

| Technique | Finding |
|-----------|---------|
| Fanout 32 | +154% MergeFrozenMap; 31% occupancy wastes copy bandwidth |
| Fanout 16 | Erratic; +62% Set-With at 1ki, no clear improvement |
| Fanout 4 | +47% SetBuilder at 1Mi; depth penalty exceeds cache benefit |
| Compact arrays | 95% occupancy makes savings negligible; extra alloc hurts |

### Remaining Regressions

Set-With (+14-30%) and IntSet/With (+60%) remain regressed relative to the
pre-simplification baseline. These are driven by structural differences in the
simplified node types (unified leaf's `splitLeaf` creates deeper branch chains
than the old `leaf2` which stored two elements flat). Further improvement would
require either restoring a specialized two-element node type or finding a way to
reduce per-level allocation cost below one allocation per branch.

## 6. Phase 2 Optimizations: Read/Remove Paths

The optimizations in section 2 targeted the mutation paths (With, Add, Combine).
A second round of work targeted the read and remove paths (Get, Has, Without,
Difference, Intersection, SubsetOf) and iteration, which were still paying
per-call dispatch costs.

### 6.1 EqArgs Threading Through Read/Remove Paths

**Problem.** Every `Tree.Get` and `Tree.Without` call went through
`newHasher(v, 0)` which performs a `sync.Map` lookup via `getHashFunc[T]()`.
At every round boundary during descent, `nextHasher` called `newHasher` again.
Leaf equality checks used `value.Equal()` — a full type-switch per comparison.
`Difference`, `Intersection`, and `SubsetOf` compounded both costs across all
elements.

**Solution.** Thread `*EqArgs[T]` (which carries pre-resolved `eq` and `hash`
functions plus the parallelism `depth.Gauge`) through all six node interface
methods: `Get`, `Without`, `Remove`, `Difference`, `Intersection`, `SubsetOf`.
At the `Tree` level, `Get` and `Without` create a `DefaultNPEqArgs[T]()` once
at entry and use `newHasherWith` + `nextHasherWith` throughout descent.
`Difference`, `Intersection`, and `SubsetOf` accept `*EqArgs[T]` directly
(replacing `depth.Gauge`), matching the existing convention of `Combine` and
`Equal`.

### 6.2 Mask-Based Iterator

**Problem.** `packerIterator.Next()` linearly scanned all `fanout` (8) slots,
checking each for nil. With typical 2–4 occupied slots per branch, most
iterations were wasted nil checks.

**Solution.** Replace the linear scan with `masker.Masker` bit iteration. The
iterator stores the branch's occupancy mask and advances via `FirstIndex()` /
`Next()`, jumping directly to populated slots. The stack buffer is passed
through to child iterators unchanged (no longer appended with `p.data[:]`).

### 6.3 Batched Allocation for Without (buildSpineWithout)

**Problem.** `branch.Without` allocated a new `branch[T]` per tree level via
`ret := *b`. For depth-7 trees (1M elements, fanout 8), that's 7 separate heap
allocations per `Without` call.

**Solution.** `withoutBatched` iteratively descends through branches, recording
the path in a stack-allocated `[levelsPerRound*2]spineEntry` buffer. On
reaching a leaf, it calls `leaf.Without`, and if the element was found,
`buildSpineWithout` batch-allocates all branch copies in a single
`make([]branch[T], pathLen)` slice and wires them bottom-up. The bottom branch
is collapsed if its count drops to ≤ `maxLeafLen`. This mirrors the existing
`withFastBatched` pattern for `With`.

### 6.4 Benchmark Results

Benchmarks on Apple M4 Max, 6 iterations per measurement, comparing the commit
before these three optimizations (`c5e25c0`) against the final state
(`6d361b0`). The benchmark file exercises the exact operations optimized.

**Latency (sec/op):**

| Benchmark | Before | After | Change |
|-----------|--------|-------|--------|
| SetIsSubsetOf 1K | 22.2 µs | 5.97 µs | **−73%** |
| SetIsSubsetOf 1M | 9.65 ms | 4.02 ms | **−58%** |
| SetIntersection 1K | 38.4 µs | 22.5 µs | **−41%** |
| SetDifference 1K | 49.4 µs | 32.3 µs | **−35%** |
| SetDifference 1M | 11.6 ms | 7.9 ms | **−32%** |
| SetIntersection 1M | 9.4 ms | 7.1 ms | **−25%** |
| SetRange 1M | 58.5 ms | 47.4 ms | **−19%** |
| SetRange 1K | 26.4 µs | 22.1 µs | **−16%** |
| SetWithout 1M | 670 ns | 582 ns | **−13%** |
| SetWithout 1K | 203 ns | 188 ns | **−7%** |
| **geomean** | **48.2 µs** | **34.1 µs** | **−29%** |

**Allocations (allocs/op):**

| Benchmark | Before | After | Change |
|-----------|--------|-------|--------|
| SetIsSubsetOf 1M | 1,033K | 37K | **−96%** |
| SetIsSubsetOf 1K | 361 | 38 | **−89%** |
| SetDifference 1M | 2,684K | 731K | **−73%** |
| SetWithout 1M | 10 | 4 | **−60%** |
| SetIntersection 1M | 1,636K | 639K | **−61%** |
| SetIntersection 1K | 852 | 532 | **−38%** |
| SetDifference 1K | 1,383 | 890 | **−36%** |
| SetWithout 1K | 5 | 4 | **−20%** |
| **geomean** | **1,376** | **638** | **−54%** |

**Bytes/op:**

| Benchmark | Before | After | Change |
|-----------|--------|-------|--------|
| SetIsSubsetOf 1M | 9.3 MiB | 1.7 MiB | **−82%** |
| SetIsSubsetOf 1K | 4.3 KiB | 1.7 KiB | **−60%** |
| SetDifference 1M | 35.3 MiB | 20.3 MiB | **−43%** |
| SetIntersection 1M | 25.0 MiB | 17.4 MiB | **−30%** |
| SetIntersection 1K | 19.2 KiB | 16.9 KiB | **−12%** |
| **geomean** | **30.5 KiB** | **27.1 KiB** | **−11%** |

### 6.5 Analysis

The largest improvements appear on `IsSubsetOf` (−73% time, −96% allocs at 1M).
This operation was the most allocation-heavy before: every `SubsetOf` call
descended through `Get` at each leaf, and each `Get` performed a `sync.Map`
lookup per round boundary. With pre-resolved functions, the only remaining
allocations are the `EqArgs` struct at the entry point.

`Difference` and `Intersection` show 25–41% latency improvements and 36–73%
fewer allocations. These operations traverse both trees and call `Get` or
`Without` on leaf elements in loops — the per-element `sync.Map` and
`value.Equal` dispatch costs compounded across thousands of elements.

`Without` shows a modest 7–13% improvement. The batched spine allocation
(phase 3) reduces heap allocations from O(depth) to O(1), but `Without`
operations are already fast in absolute terms (sub-microsecond at 1K).

`Range` (iteration) improved 16–19%, attributable to the mask-based iterator
skipping nil slots via bit manipulation instead of linear scanning.

`Has` (Get) showed a regression in bytes/op (+300% at 1K) due to the
`DefaultNPEqArgs[T]()` allocation at the `Tree.Get` entry point. The
pre-optimization code allocated nothing at the entry point (hash/eq were
resolved inline). A per-type cache for default `EqArgs` — matching the existing
`defaultKeyCombineArgsCache` pattern — would eliminate this overhead.

## 7. Future Work

The following ideas have been identified but not yet explored.

### 7.1 splitLeaf Hot Path Optimization

The primary source of the remaining Set-With regression (+14-30%). When two
elements collide at a given hash prefix depth, `splitLeaf` creates a full branch
node plus two `leaf1` nodes — three allocations. The old `leaf2` stored two
elements flat with zero branch overhead. Options:

- **Restore a specialized two-element node** (`leaf2`): eliminates the branch
  creation for the common case of exactly 2 elements sharing a hash prefix.
  Trades code complexity (~200 lines of type-switch cases) for a ~20%
  improvement on insertion-heavy workloads.
- **Batch splitLeaf**: when splitting a `leaf[T]` with N elements, partition all
  N elements by their next hash chunk in a single pass rather than inserting
  one-by-one. This reduces intermediate allocations from O(N) to O(fanout).

### 7.2 IntSet/With Profiling

The steepest remaining regression (+60%). Needs dedicated CPU profiling to
determine whether the cause is the same `splitLeaf` overhead or something
specific to the bitmap-compressed integer encoding (`Map[I, cellMask]`). The
bitmap layer adds its own With operations on top of the tree's, so the
regression may compound differently than for plain `Set[T]`.

### 7.3 pkg/rel Generics Migration

The `pkg/rel/` subpackage (relational algebra: Tuple, Relation, Join,
CartesianProduct, Project) currently does not compile against the generics-based
API. It uses pre-generic types (`frozen.Map` without type parameters,
`frozen.StringMapBuilder`, `frozen.NewSetFromStrings`, etc.) and methods that no
longer exist (`.Equal`, `.Hash` on value types). A migration pass is needed to:

- Instantiate generic types (`frozen.Map[string, Value]`, `frozen.Set[Value]`,
  etc.)
- Replace removed helper constructors with their generic equivalents
- Update value types to satisfy the `Key[T]` constraint interface

### 7.4 Linter Upgrade

The branch name (`upgrade-linter`) suggests this was the original goal.
`.golangci.yml` targets golangci-lint v1.60.1. Newer versions may support
additional linters, improved analysis, and updated rule sets.

### 7.5 Leaf Slice Preallocation in splitLeaf

When `splitLeaf` distributes N elements into child nodes, the resulting `leaf`
slices grow dynamically. Preallocating at the expected partition size
(`N / fanout` or a small constant) could reduce GC pressure from intermediate
slice growth. Impact is likely small given that leaf sizes are bounded at 8
elements, but worth measuring.

### 7.6 Cache DefaultNPEqArgs Per Type

`Tree.Get` and `Tree.Without` currently allocate a `DefaultNPEqArgs[T]()` on
every call. Adding a `sync.Map`-based per-type cache — matching the existing
`defaultKeyCombineArgsCache` pattern in `builder.go` — would eliminate the
`Has` bytes/op regression (+300% at 1K) identified in section 6.5.

### 7.7 Lazy branch.count

The `branch.count` field is maintained eagerly on every `With`/`Without`
mutation, requiring arithmetic at each branch level. If `Count()` is called
infrequently relative to mutations, computing it lazily — either on demand by
summing children, or via a dirty-flag cache — could eliminate per-mutation
overhead. Requires profiling to determine whether `count` maintenance is
actually visible in profiles.
