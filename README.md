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
BenchmarkInsertMapInt0-8             5247402           199 ns/op          51 B/op          0 allocs/op
BenchmarkInsertMapInt1k-8            5994780           193 ns/op          46 B/op          0 allocs/op
BenchmarkInsertMapInt1M-8            6461139           206 ns/op          51 B/op          0 allocs/op
BenchmarkInsertMapInterface0-8       5600690           364 ns/op          73 B/op          2 allocs/op
BenchmarkInsertMapInterface1k-8      5510244           378 ns/op          73 B/op          2 allocs/op
BenchmarkInsertMapInterface1M-8      5364372           389 ns/op          80 B/op          2 allocs/op
BenchmarkInsertFrozenMap0-8          5754868           200 ns/op         144 B/op          4 allocs/op
BenchmarkInsertFrozenMap1k-8         1923518           628 ns/op         584 B/op          8 allocs/op
BenchmarkInsertFrozenMap1M-8          902554          1757 ns/op         901 B/op         11 allocs/op
BenchmarkInsertFrozenNode0-8         6056293           194 ns/op         144 B/op          4 allocs/op
BenchmarkInsertFrozenNode1k-8        1918303           622 ns/op         584 B/op          8 allocs/op
BenchmarkInsertFrozenNode1M-8         910125          1752 ns/op         901 B/op         11 allocs/op
BenchmarkInsertSetInt0-8             5966091           178 ns/op          26 B/op          0 allocs/op
BenchmarkInsertSetInt1k-8            6793104           176 ns/op          25 B/op          0 allocs/op
BenchmarkInsertSetInt1M-8            6985861           184 ns/op          26 B/op          0 allocs/op
BenchmarkInsertSetInterface0-8       3895648           357 ns/op          51 B/op          1 allocs/op
BenchmarkInsertSetInterface1k-8      4016401           371 ns/op          49 B/op          1 allocs/op
BenchmarkInsertSetInterface1M-8      5609694           330 ns/op          40 B/op          1 allocs/op
BenchmarkInsertFrozenSet0-8          8715810           123 ns/op         104 B/op          2 allocs/op
BenchmarkInsertFrozenSet1k-8         2123174           516 ns/op         544 B/op          6 allocs/op
BenchmarkInsertFrozenSet1M-8          959947          1518 ns/op         861 B/op          9 allocs/op
⋮
```

[![](assets/benchmarks.png)](https://docs.google.com/spreadsheets/d/1Sq48pT4sKLHx2uY_nSljfbFpEJijXhNAeoB-BbDlrsI/edit?usp=sharing)

## Bugs

Test coverage is sparse.
