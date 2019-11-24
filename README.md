# Frozen

![Go build status](https://github.com/marcelocantos/frozen/workflows/Go/badge.svg)

Efficient immutable data types.

## Types

Map and Set both use a [hashed array map trie
(HAMT)][1].

- Map: Associates keys with values.
- Set: Stores sets of values.

## Performance

The following benchmarks test the base HAMT implementation against several other
key-value map implementations. All implementations are tested for insertions
against an empty map and a map prepopulated with one million elements, except
for github.com/mediocregopher/seq, which only has 10k elements prepopulated to
avoid timeouts. In order of appearance in the benchmark, the implementations are
as follows:

| Benchmark       | Type                           |
| --------------- | ------------------------------ |
| FrozenHamt      | frozen.hamt                    |
| FrozenMap       | frozen.Map                     |
| FrozenSet       | frozen.Set                     |
| MapInt          | map[int]int                    |
| MapInterface    | map[interface{}]interface{}    |
| MediocreHashMap | seq.HashMap                    |
| MediocreSet     | seq.Set                        |
| SetInt          | set = map[int]struct{}         |
| SetInterface    | set = map[interface{}]struct{} |

In all cases, ints are mapped to ints.

```bash
$ go test -run ^$ -cpuprofile cpu.prof -memprofile mem.prof -benchmem -bench ^BenchmarkInsert ./...
goos: darwin
goarch: amd64
pkg: github.com/marcelocantos/frozen/pkg/frozen
BenchmarkInsertFrozenHamt0-8          	 2175505	       667 ns/op	    1182 B/op	       5 allocs/op
BenchmarkInsertFrozenHamt1k-8         	 1342575	      1475 ns/op	    1154 B/op	       6 allocs/op
BenchmarkInsertFrozenHamt1M-8         	  653156	      1681 ns/op	    1266 B/op	       6 allocs/op
BenchmarkInsertMapInt0-8              	 4456830	       242 ns/op	      79 B/op	       0 allocs/op
BenchmarkInsertMapInt1k-8             	 5691613	       205 ns/op	      63 B/op	       0 allocs/op
BenchmarkInsertMapInt1M-8             	 6212853	       234 ns/op	      99 B/op	       0 allocs/op
BenchmarkInsertMapInterface0-8        	 2402386	       477 ns/op	     151 B/op	       2 allocs/op
BenchmarkInsertMapInterface1k-8       	 2218449	       526 ns/op	     162 B/op	       2 allocs/op
BenchmarkInsertMapInterface1M-8       	 2805478	       603 ns/op	     189 B/op	       2 allocs/op
BenchmarkInsertFrozenMap0-8           	 2766694	       381 ns/op	     480 B/op	       6 allocs/op
BenchmarkInsertFrozenMap1k-8          	 1259384	       941 ns/op	     882 B/op	       6 allocs/op
BenchmarkInsertFrozenMap1M-8          	  430938	      2549 ns/op	    1266 B/op	       6 allocs/op
BenchmarkInsertMediocreHashMap0-8     	 3554623	       391 ns/op	     120 B/op	       4 allocs/op
BenchmarkInsertMediocreHashMap10-8    	  943761	      1127 ns/op	     530 B/op	       7 allocs/op
BenchmarkInsertMediocreHashMap10k-8   	    5233	    365611 ns/op	  102840 B/op	     646 allocs/op
BenchmarkInsertSetInt0-8              	 2300034	       527 ns/op	      43 B/op	       0 allocs/op
BenchmarkInsertSetInt1k-8             	 3203769	       338 ns/op	      33 B/op	       0 allocs/op
BenchmarkInsertSetInt1M-8             	 5868884	       225 ns/op	      59 B/op	       0 allocs/op
BenchmarkInsertSetInterface0-8        	 2779015	       443 ns/op	      70 B/op	       1 allocs/op
BenchmarkInsertSetInterface1k-8       	 2192677	       504 ns/op	      86 B/op	       1 allocs/op
BenchmarkInsertSetInterface1M-8       	 3472974	       476 ns/op	      82 B/op	       1 allocs/op
BenchmarkInsertFrozenSet0-8           	 3246156	       360 ns/op	     440 B/op	       3 allocs/op
BenchmarkInsertFrozenSet1k-8          	 1310988	       842 ns/op	     842 B/op	       4 allocs/op
BenchmarkInsertFrozenSet1M-8          	  651750	      1750 ns/op	    1226 B/op	       4 allocs/op
BenchmarkInsertMediocreSet0-8          	11555031	        96.9 ns/op	      72 B/op	       2 allocs/op
BenchmarkInsertMediocreSet10-8         	 2205480	       508 ns/op	     482 B/op	       4 allocs/op
BenchmarkInsertMediocreSet10k-8        	    9410	    113726 ns/op	  102784 B/op	     643 allocs/op
PASS
ok  	github.com/marcelocantos/frozen/pkg/frozen	417.918s
```


[![](assets/benchmarks.png)](https://docs.google.com/spreadsheets/d/1Sq48pT4sKLHx2uY_nSljfbFpEJijXhNAeoB-BbDlrsI/edit?usp=sharing)

## Bugs

Test coverage is sparse.


[1]: https://en.wikipedia.org/wiki/Hash_array_mapped_trie
