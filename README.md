# Frozen

![Go build status](https://github.com/marcelocantos/frozen/workflows/Go/badge.svg)

Efficient immutable data types.

## Types

Map and Set both use a hashed array trie.

- Map: Associates keys with values.
- Set: Stores sets of values.

## Performance

The following benchmarks test the base node implementation against several other
key-value map implementations. All implementations are tested for insertions
against an empty map and a map prepopulated with one million elements, except
for github.com/mediocregopher/seq, which only has 10k elements prepopulated to
avoid timeouts. In order of appearance in the benchmark, the implementations are
as follows:

| Benchmark       | Type                           |
| --------------- | ------------------------------ |
| FrozenNode      | frozen.node                    |
| FrozenMap       | frozen.Map                     |
| FrozenSet       | frozen.Set                     |
| MapInt          | map[int]int                    |
| MapInterface    | map[interface{}]interface{}    |
| SetInt          | set = map[int]struct{}         |
| SetInterface    | set = map[interface{}]struct{} |

In all cases, ints are mapped to ints.

```bash
$ go test -run ^$ -cpuprofile cpu.prof -memprofile mem.prof -benchmem -bench ^BenchmarkInsert ./pkg/frozen/
⋮
BenchmarkInsertMapInt0-8             5247402           199 ns/op          51 B/op           0 allocs/op
BenchmarkInsertMapInt1k-8            5994780           193 ns/op          46 B/op           0 allocs/op
BenchmarkInsertMapInt1M-8            6461139           206 ns/op          51 B/op           0 allocs/op
BenchmarkInsertMapInterface0-8       5600690           364 ns/op          73 B/op           2 allocs/op
BenchmarkInsertMapInterface1k-8      5510244           378 ns/op          73 B/op           2 allocs/op
BenchmarkInsertMapInterface1M-8      5364372           389 ns/op          80 B/op           2 allocs/op
BenchmarkInsertFrozenMap0-8          5728576           206 ns/op         128 B/op           4 allocs/op
BenchmarkInsertFrozenMap1k-8         1948087           624 ns/op         495 B/op           8 allocs/op
BenchmarkInsertFrozenMap1M-8          883378          1760 ns/op         759 B/op          11 allocs/op
BenchmarkInsertFrozenNode0-8         5796955           198 ns/op         128 B/op           4 allocs/op
BenchmarkInsertFrozenNode1k-8        1938098           605 ns/op         495 B/op           8 allocs/op
BenchmarkInsertFrozenNode1M-8         947007          1794 ns/op         759 B/op          11 allocs/op
BenchmarkInsertSetInt0-8             5966091           178 ns/op          26 B/op           0 allocs/op
BenchmarkInsertSetInt1k-8            6793104           176 ns/op          25 B/op           0 allocs/op
BenchmarkInsertSetInt1M-8            6985861           184 ns/op          26 B/op           0 allocs/op
BenchmarkInsertSetInterface0-8       3895648           357 ns/op          51 B/op           1 allocs/op
BenchmarkInsertSetInterface1k-8      4016401           371 ns/op          49 B/op           1 allocs/op
BenchmarkInsertSetInterface1M-8      5609694           330 ns/op          40 B/op           1 allocs/op
BenchmarkInsertFrozenSet0-8         10061632           113 ns/op          88 B/op           2 allocs/op
BenchmarkInsertFrozenSet1k-8         2387151           493 ns/op         455 B/op           6 allocs/op
BenchmarkInsertFrozenSet1M-8          966304          1536 ns/op         719 B/op           9 allocs/op
⋮
```

![](assets/benchmarks.png)

## Bugs

Test coverage is sparse.
