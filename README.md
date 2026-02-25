# Frozen

![Go build status](https://github.com/arr-ai/frozen/workflows/Go/badge.svg)

Efficient immutable data types.

## Overview

`frozen` is a Go 1.19+ library of immutable, persistent data structures built on hashed array tries (HAT). All mutation operations return a new value that shares structure with the original; no existing value is ever modified. The library uses Go generics throughout.

```
go get github.com/arr-ai/frozen
```

## Types

### `Set[T any]`

An immutable set backed by a hashed array trie.

Key operations: `With`, `Without`, `Has`, `Union`, `Intersection`, `Difference`, `SymmetricDifference`, `Where`, `Reduce`, `Reduce2`, `IsSubsetOf`.

Package-level functions: `Powerset[T]`, `SetMap[T, U]`, `SetGroupBy[T, K]`, `SetAs[U, T]`.

```go
import "github.com/arr-ai/frozen"

s := frozen.NewSet(1, 2, 3)
s2 := s.With(4).Without(2)       // {1, 3, 4}
fmt.Println(s2.Has(3))           // true
fmt.Println(s2.Count())          // 3

evens := s2.Where(func(n int) bool { return n%2 == 0 }) // {4}

u := s.Union(frozen.NewSet(3, 4, 5))         // {1, 2, 3, 4, 5}
d := s.Difference(frozen.NewSet(2, 3))       // {1}

sum, _ := s.Reduce2(func(a, b int) int { return a + b }) // 6
```

### `Map[K any, V any]`

An immutable key-value map backed by a hashed array trie.

Key operations: `With`, `Without`, `Get`, `MustGet`, `GetElse`, `Has`, `Keys`, `Values`, `Where`, `Merge`, `Update`, `Project`.

Package-level functions: `MapMap[K, V, U]`, `NewMapFromKeys`, `NewMapFromGoMap`, `MapToGoMap`.

```go
m := frozen.NewMap(frozen.KV("a", 1), frozen.KV("b", 2))
m2 := m.With("c", 3).Without("a")   // (b: 2, c: 3)

if v, ok := m2.Get("b"); ok {
    fmt.Println(v) // 2
}

doubled := frozen.MapMap(m, func(k string, v int) int { return v * 2 })

merged := m.Merge(m2, func(k string, a, b int) int { return a + b })
```

### `IntSet[I integer]`

A specialised immutable set for integer types, using 64-bit bitmap compression to pack values into cells. More memory-efficient than `Set[int]` for dense integer ranges.

The `integer` constraint covers `~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr`.

Key operations: `With`, `Without`, `Has`, `Union`, `Intersection`, `Where`, `Map`, `IsSubsetOf`.

```go
s := frozen.NewIntSet(1, 2, 3, 100, 101)
fmt.Println(s.Has(100))              // true
fmt.Println(s.Count())               // 5
u := s.Union(frozen.NewIntSet(3, 4)) // {1, 2, 3, 4, 100, 101}
```

### `Key[T any]`

A constraint interface that types must satisfy to be usable as custom map keys or set elements with value-based equality:

```go
type Key[T any] interface {
    value.Equaler[T]   // Equal(T) bool
    hash.Hashable      // Hash(seed uintptr) uintptr
}
```

`Set[T]` and `Map[K, V]` themselves implement `Key`, so they can be nested.

### Builders

Builders accumulate elements incrementally and are more efficient than repeated `With` calls when constructing large collections.

```go
// SetBuilder
b := frozen.NewSetBuilder[string](16)
b.Add("x")
b.Add("y")
s := b.Finish() // immutable Set[string]; builder is reset

// MapBuilder
mb := frozen.NewMapBuilder[string, int](16)
mb.Put("a", 1)
mb.Put("b", 2)
m := mb.Finish() // immutable Map[string, int]
```

`Finish()` returns the immutable result and resets the builder for reuse.

## Subpackages

### `lazy/`

A `Set` interface for deferred evaluation with memoization. Operations such as `Union`, `Intersection`, `Difference`, `Where`, and `Powerset` are composed lazily and only evaluated when elements are accessed. Useful for constructing large set expressions without paying the full construction cost up front.

### `pkg/rel/`

Relational algebra built on `frozen.Map` and `frozen.Set`:

```go
type Tuple    = frozen.Map[string, any]
type Relation = frozen.Set[Tuple]
```

Provided operations: `New`, `Join`, `CartesianProduct`, `Project`, `Nest`, `Unnest`.

```go
r := rel.New(
    []string{"name", "age"},
    []any{"alice", 30},
    []any{"bob", 25},
)
names := rel.Project(r, "name")
```

## Performance

The following benchmarks test the base node implementation against several other key-value map implementations. All implementations are tested for insertions against an empty map, a map prepopulated with 1k elements and one prepopulated with 1M elements. The implementations are as follows:

> Note: these benchmarks were recorded before the generics rewrite; relative ordering remains indicative but absolute numbers will differ on current hardware and Go versions.

| Benchmark       | Type                           |
| --------------- | ------------------------------ |
| MapInt          | map[int]int                    |
| MapInterface    | map[any]any                    |
| FrozenMap       | frozen.Map                     |
| FrozenNode      | frozen.node                    |
| SetInt          | set = map[int]struct{}         |
| SetInterface    | set = map[any]struct{}         |
| FrozenSet       | frozen.Set                     |

In all cases, ints are mapped to ints.

```bash
$ go test -run ^$ -cpuprofile cpu.prof -memprofile mem.prof -benchmem -bench ^BenchmarkInsert .
goos: linux
goarch: amd64
pkg: github.com/arr-ai/frozen
BenchmarkInsertMapInt0-24           	 8532830	       175 ns/op	      72 B/op	       0 allocs/op
BenchmarkInsertMapInt1k-24          	10379329	       164 ns/op	      60 B/op	       0 allocs/op
BenchmarkInsertMapInt1M-24          	 6760242	       185 ns/op	      78 B/op	       0 allocs/op
BenchmarkInsertMapInterface0-24     	 3579843	       348 ns/op	     152 B/op	       2 allocs/op
BenchmarkInsertMapInterface1k-24    	 3675631	       365 ns/op	     148 B/op	       2 allocs/op
BenchmarkInsertMapInterface1M-24    	 6517272	       354 ns/op	     115 B/op	       2 allocs/op
BenchmarkInsertFrozenMap0-24        	 5443401	       225 ns/op	     240 B/op	       6 allocs/op
BenchmarkInsertFrozenMap1k-24       	 2553954	       446 ns/op	     635 B/op	      10 allocs/op
BenchmarkInsertFrozenMap1M-24       	 1263691	       960 ns/op	     954 B/op	      13 allocs/op
BenchmarkInsertFrozenNode0-24       	 8220901	       141 ns/op	     144 B/op	       4 allocs/op
BenchmarkInsertFrozenNode1k-24      	 3294789	       388 ns/op	     539 B/op	       8 allocs/op
BenchmarkInsertFrozenNode1M-24      	 1316443	       871 ns/op	     858 B/op	      11 allocs/op
BenchmarkInsertSetInt0-24           	12816358	       155 ns/op	      29 B/op	       0 allocs/op
BenchmarkInsertSetInt1k-24          	12738687	       155 ns/op	      29 B/op	       0 allocs/op
BenchmarkInsertSetInt1M-24          	 7613054	       171 ns/op	      39 B/op	       0 allocs/op
BenchmarkInsertSetInterface0-24     	 5121948	       302 ns/op	      58 B/op	       1 allocs/op
BenchmarkInsertSetInterface1k-24    	 5051988	       303 ns/op	      58 B/op	       1 allocs/op
BenchmarkInsertSetInterface1M-24    	 3172472	       329 ns/op	      62 B/op	       1 allocs/op
BenchmarkInsertFrozenSet0-24        	 5400745	       236 ns/op	     296 B/op	       6 allocs/op
BenchmarkInsertFrozenSet1k-24       	 2460313	       512 ns/op	     787 B/op	      11 allocs/op
BenchmarkInsertFrozenSet1M-24       	 1132215	      1046 ns/op	    1106 B/op	      14 allocs/op
PASS
ok  	github.com/arr-ai/frozen	65.909s
```

![Benchmarks Graph](assets/benchmarks.png)
