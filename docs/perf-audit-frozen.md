# Performance Audit: frozen Generic HAMT Library

## Overview

The migration from `interface{}`-based to Go-generics-based collections introduced
a 1.4–1.7x regression in arrai's join benchmarks, with 50–60% more heap allocations.
Profiling shows 54–58% of CPU time in GC. The root cause is closure and boxing
allocations in frozen's hash/equality dispatch that didn't exist pre-generics.

This report covers frozen-level findings and optimization strategies.

## The Core Problem: Per-Operation Closure Allocations

Every HAMT operation (Get, Has, With, Equal, Hash) now allocates closures for
hash and equality dispatch. Pre-generics, these were simple inline operations.

| Operation | Expected allocs | Actual allocs | Source |
|---|---|---|---|
| `Map[string, Value].Get` | 0 | 6 (144B) | `value.Equal` + `resolveHashFunc` closures |
| `Set[string].Has` | 0 | 4 (64B) | Same |
| `Set[Value].Has` | 0 | 1 (8B) | Reduced but non-zero |
| `Map[string, Value].Hash` | 0 | 4 (160B) | `hash.Any` boxing |
| `Set[Value].Equal` (1000) | 0 | 74 | Per-node equality closures |
| `Set[Value].Range` (1000) | ~5 | 657 | Iterator + node allocs |

### Profiling Data (RelationSetJoin1000)

Top new allocation sources not present in pre-generics code:

| Function | Alloc count | % of total | Category |
|---|---|---|---|
| `value.Equal[KV{any, Set}]` | 69M | 26.2% | Closure allocation |
| `iterator.NewSliceIterator[any]` | 28M | 10.7% | Iterator objects |
| `tree.resolveHashFunc[KV...].func1` | 20M | 7.6% | Hash closure allocation |

Pre-generics, the top allocators were all iterator buffers and leaf nodes — no
closure allocations appeared in the profile at all.

## Root Cause Analysis

### 1. `value.Equal[T]` allocates on every call

```go
func Equal[T any](a, b T) bool {
    var i any = a          // boxes T to any (may allocate)
    switch a := i.(type) {
    case Equaler[T]:
        return a.Equal(b)  // type assertion dispatch
    case Samer:
        return a.Same(b)
    }
    return i == any(b)
}
```

When `T = Value` (an interface), `var i any = a` is cheap (interface-to-interface).
But the `i.(Equaler[T])` type assertion involves a runtime itab lookup. More
critically, this function is called on every HAMT node visit during Get, Has,
and Equal operations — millions of times per join.

**The CombineArgs pattern**: `CombineArgs[T]` already caches a pre-resolved `eq`
function. But only mutation paths (With, Builder.Add) use it. Read paths (Get,
Has) call `value.Equal` directly.

### 2. `EqualFuncFor[Value]` returns `equalSlow` for interfaces

```go
func EqualFuncFor[T any]() func(a, b T) bool {
    var t T
    switch any(t).(type) {
    case Equaler[T]:
        return func(a, b T) bool { return any(a).(Equaler[T]).Equal(b) }
    // ...
    case nil:  // ← interface types hit this (zero value is nil)
        return equalSlow[T]  // falls back to per-call type switch
    }
}
```

When `T` is an interface type like `Value`, the zero value is `nil`, so it hits
the `case nil` branch and returns `equalSlow[T]` — which does the full boxing +
type-switch on every call. This defeats the purpose of pre-resolution.

### 3. `resolveHashFunc[T]` returns boxing closures

For `mapEntry[string, Value]`, the resolved hash function does:
```go
func(key T, seed uintptr) uintptr {
    return any(key).(hash.Hashable).Hash(seed)  // boxes key to any every call
}
```

This creates a heap allocation per hash computation. For `Map[string, Value].Get`,
this is called at each HAMT level (typically 2-4 levels).

### 4. Go GCShape stenciling causes double dispatch

Go's generics implementation (GCShape stenciling) means all pointer types share
one compiled code shape. The `node[T]` interface — the hot path for every tree
operation — pays:
1. Interface vtable dispatch (to find concrete type: branch, leaf1, leaf2, leaf)
2. Generic dictionary dispatch (to resolve T-specific methods)

Pre-generics, only the vtable dispatch existed. PlanetScale benchmarking shows
42–91% overhead for this double dispatch pattern.

### 5. Iterator scans all slots

`packer_iter.go` has a TODO noting that the mask should be used to skip empty
slots. Currently it linearly scans all 8 (or fanout) children in each branch,
even when only 2-3 are occupied.

## HAMT Structure Overview

```
Fanout:    8 (FanoutBits=3)
Hash:      64-bit, 21 levels per round, multi-round with seed increment
Node types: branch[T] (internal), leaf1[T] (1 elem), leaf2[T] (2 elem), leaf[T] (3-8 collision)
Branch:    packer[T] { mask uint16, data [8]node[T] } + count int  (~152 bytes)
```

Key optimization already applied: `buildSpine` in `withFastBatched` batch-allocates
all branch copies in a single `[]branch` slice instead of one `*branch` per level.
This reduced With allocations by ~60% at 1M elements.

## Optimization Strategies

### Tier 1: Quick Wins (Low Risk, High Impact)

#### Strategy 1: Thread cached hash/equal through all code paths

**Problem**: `CombineArgs[T]` caches `eq` and `hash` functions, but only mutation
paths use them. Read paths (`Get`, `Has`, `leaf.Get`, `leaf.Remove`) call
`value.Equal` directly.

**Fix**: Add `args CombineArgs[T]` parameter to `Get`, `Has`, and all leaf/branch
methods that call `value.Equal` or compute hashes. Compute args once at the
`Set`/`Map` method level and thread through.

**Expected impact**: Eliminate 69M+ closure allocations in RelationSetJoin (26% of
total). `Map.Get` should go from 6 allocs to 0.

**Complexity**: Medium — requires signature changes through the tree layer.

#### Strategy 2: Fix `EqualFuncFor` for interface types

**Problem**: Interface types fall into `case nil`, returning `equalSlow`.

**Fix**: Add an explicit check before the nil case:
```go
// Check if the type constraint itself is Equaler
var probe T
if _, ok := any(&probe).(interface{ Equal(T) bool }); ok {
    return func(a, b T) bool { return any(a).(Equaler[T]).Equal(b) }
}
```
Or use `reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Interface` to detect
interface types and return a direct dispatch closure.

**Expected impact**: 30–50% reduction in equality dispatch cost when using cached
functions.

**Complexity**: Low — single function change in `value.go`.

#### Strategy 3: Eliminate hash closure boxing

**Problem**: `resolveHashFunc` returns closures that box keys via `any(key)`.

**Fix**: For types implementing `hash.Hashable`, capture the concrete `Hash` method
at resolution time. For `mapEntry[K,V]`, generate a specialized hash that hashes
the key field directly without boxing.

**Expected impact**: Eliminate 20–44M closure allocations per join benchmark.

**Complexity**: Low-medium.

#### Strategy 4: Iterator mask optimization

**Problem**: `packer_iter.go` scans all fanout slots. Source has a TODO for this.

**Fix**: Use `mask.FirstIndex()` + `mask.Next()` to skip empty slots. Also
pre-allocate a single iteration buffer instead of per-node `[]node[T]` slices.

**Expected impact**: 20–40% iteration improvement. Fewer allocations from
`packedIteratorBuf`.

**Complexity**: Low.

### Tier 2: Architectural (Medium Risk, High Impact)

#### Strategy 5: CHAMP compact arrays

**Problem**: Each `branch[T]` allocates a full `[8]node[T]` array (128 bytes of
interface pointers) regardless of occupancy. Values and sub-nodes are intermixed
behind the `node[T]` interface, requiring type assertions during iteration.

**Fix**: Adopt the CHAMP layout (Steindorfer & Vinju, OOPSLA 2015):
```go
type branch[T any] struct {
    datamap  uint16   // bitmap: which slots have inline values
    nodemap  uint16   // bitmap: which slots have sub-nodes
    content  []any    // [values (T)..., sub-nodes (*branch[T])...]
    count    int
}
```

Values are stored at the front, sub-nodes at the back. Size is `popcount(datamap) +
popcount(nodemap)`. No `node[T]` interface needed for stored values.

**Expected impact (from literature)**:
- Lookup: 23–72% faster
- Iteration: 39–83% faster
- Equality: 81–96% faster (bitmap short-circuit)
- Memory: 16–68% reduction

**Complexity**: High — fundamental restructure of tree internals.

#### Strategy 6: Expand `buildSpine` to all mutation paths

**Problem**: Only `With`/`WithFast` use batched spine allocation. `Without`,
`Combine`, `Difference`, `Intersection` still allocate one branch per level.

**Fix**: Apply the same pattern: iterative descent recording the path, then
single-allocation batch construction.

**Expected impact**: 30–60% allocation reduction for non-With mutations.

**Complexity**: Medium — each operation needs its own batching logic.

#### Strategy 7: Node pooling for Builder

**Problem**: Builder allocates new branch/leaf nodes on every mutation.

**Fix**: Use `sync.Pool` for `branch[T]` nodes, resetting and reusing them. Most
beneficial during batch builds where many temporary nodes are created.

**Expected impact**: Moderate GC reduction during bulk builds.

**Complexity**: Low-medium.

### Tier 3: Longer-Term (Higher Risk)

#### Strategy 8: Type-erased internal tree with generic wrapper

**Problem**: Go's GCShape stenciling means `node[T]` interface calls pay double
dispatch. This is fundamental to Go's generics implementation and won't change.

**Fix**: Make the internal tree operate on `any` with hash/equal function pointers
(like the pre-generics code). Provide type-safe generic wrappers at the public API
boundary only:
```go
// Internal (fast, no generics overhead)
type tree struct { root node }  // node operates on any

// Public (type-safe wrapper)
type Set[T any] struct { t tree; args combineArgs }
func (s Set[T]) Has(v T) bool { return s.t.has(s.args, v) }
```

**Expected impact**: Eliminates all generics dispatch overhead on hot paths.
Recovers essentially the full pre-generics performance.

**Complexity**: High — requires restructuring the API layering.

**Risk**: Medium — more complex internal code, but the pattern is well-established
(many Go generic libraries use type-erased internals).

## Implementation Results

### Completed: b3ccecb — Reduce allocations, batch spine operations

Implemented Strategies 1, 3, and 6:

- **Cached hash/eq threading** (Strategy 1): `EqArgs[T]` now carries pre-resolved
  `eq` and `hash` functions through all read paths (Get, Has, Equal, SubsetOf,
  Difference, Intersection). `Tree[T].hf` field caches the hash function,
  lazy-initialized via `hashFunc()`.

- **Hash closure elimination** (Strategy 3): `getHashFunc[T]()` pre-resolves the
  hash function once per type and caches it with `sync.Map` keyed by
  `reflect.Type`. All tree operations use the cached function instead of
  per-call `hash.Any` boxing.

- **Batched spine for Without** (Strategy 6): `withoutBatched` mirrors
  `withFastBatched` — iterative descent recording the path, single-allocation
  batch construction.

### Completed: h0 — Recursive XOR hash for fast equality rejection

Not in the original roadmap. Each node caches `h0`, the XOR of seed-0 hashes of
all contained elements:

| Node | h0 storage |
|---|---|
| `leaf1` | `h0 = hash(elem, 0)` |
| `leaf2` | `h0 = ha ^ hb`, stores `ha` for O(1) recovery |
| `leaf` | `h0 = XOR(hash(e, 0) for e in data)` |
| `branch` | `h0 = XOR(child.H0() for child in children)` |

`Tree.Equal` and `branch.Equal` short-circuit on h0 mismatch: if the XOR hashes
differ, the trees cannot be equal. Builder paths leave h0 stale; `computeH0`
bottom-up pass in `Finish()` recomputes all values.

### Benchmark Results (Apple M4 Max, `-count=6`)

Three snapshots: pre-optimization (`0dad07d`), post-optimization (`b3ccecb`),
and h0 (`wip`). Full results in `benchmarks/`.

#### Equal

| Scenario | Pre-opt | Post-opt | + h0 | vs pre-opt |
|---|---|---|---|---|
| 1Mi/Half | 32µs | 21µs | **51ns** | **-99.84%** |
| 1Mi/Disjoint | 25µs | 23µs | **46ns** | **-99.81%** |
| 1Mi/Equal | 24ms | 5.2ms | **5.2ms** | -78% |
| 1Mi/Same | 11ms | 6ms | **6.4ms** | -43% |
| 1ki/Equal | 26µs | 6.7µs | **6.9µs** | -73% |

Non-equal sets (Half, Disjoint) go from O(n) traversal to O(1) h0 comparison.

#### Difference

| Scenario | Pre-opt | Post-opt | + h0 | vs pre-opt |
|---|---|---|---|---|
| 1Mi/Same | 119ms | 5ms | 8.9ms | **-92.5%** |
| 1Mi/Equal | 154ms | 8.5ms | 10.8ms | **-93.0%** |
| 1Mi/Half | 157ms | 13ms | 15ms | -90.4% |
| 1Mi/Disjoint | 109ms | 15ms | 12ms | -89.2% |

#### Intersection

| Scenario | Pre-opt | Post-opt | + h0 | vs pre-opt |
|---|---|---|---|---|
| 1Mi/Same | 31ms | 15ms | 21ms | -30% |
| 1Mi/Equal | 43ms | 21ms | 24ms | -45% |
| 1Mi/Disjoint | 27ms | 13ms | 8.7ms | **-68%** |

#### SubsetOf

| Scenario | Pre-opt | Post-opt | + h0 | vs pre-opt |
|---|---|---|---|---|
| 1Mi/Same | 21ms | 3.3ms | 9ms | -57% |
| 1Mi/Equal | 26ms | 9.4ms | 9.8ms | -63% |

#### Has (regression from larger node footprint)

| Scenario | Pre-opt | Post-opt | + h0 | vs pre-opt |
|---|---|---|---|---|
| 1ki/Hit | 41ns | 69ns | 66ns | +61% |
| 1Mi/Hit | 205ns | 203ns | 245ns | +20% |

The `Has` regression comes from the extra `uintptr` field per node increasing
memory footprint and cache pressure. This is a deliberate tradeoff: single-element
lookup is ~25ns slower, while set-level equality/difference/intersection operations
are orders of magnitude faster.

#### Memory (allocs/op)

| Benchmark | Pre-opt | + h0 | Change |
|---|---|---|---|
| Equal/1Mi/Same | 1.32M | 43.2k | **-96.7%** |
| Equal/1Mi/Half | 350 | 2 | **-99.4%** |
| Diff/1Mi/Same | 3.37M | 178k | **-94.7%** |
| Inter/1Mi/Disjoint | 1.71M | 167k | **-90.2%** |
| SubsetOf/1Mi/Same | 2.57M | 135k | **-94.7%** |

### Remaining Roadmap

**Still open** — Strategies 2, 4, 5, 7, 8:

- Strategy 2 (fix `EqualFuncFor` for interface types) — still relevant for
  `Set[Value]` / `Map[string, Value]` workloads
- Strategy 4 (iterator mask optimization) — low-hanging fruit
- Strategy 5 (CHAMP compact arrays) — high-impact architectural change
- Strategy 7 (node pooling for Builder) — moderate GC reduction
- Strategy 8 (type-erased internals) — last resort if generics overhead persists

## References

- Phil Bagwell, "Ideal Hash Trees" (2001)
  https://lampwww.epfl.ch/papers/idealhashtrees.pdf
- Michael J. Steindorfer & Jurgen J. Vinju, "Optimizing Hash-Array Mapped Tries
  for Fast and Lean Immutable JVM Collections" (OOPSLA 2015)
  https://michael.steindorfer.name/publications/oopsla15.pdf
- PlanetScale, "Generics can make your Go code slower"
  https://planetscale.com/blog/generics-can-make-your-go-code-slower
- Go GCShape stenciling proposal
  https://go.googlesource.com/proposal/+/refs/heads/master/design/generics-implementation-gcshape.md
- Clojure transient data structures
  https://clojure.org/reference/transients
