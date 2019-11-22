# Frozen

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
avoid timeouts. In order of appearance in the benchmark (fastest to
slowest) the implementations are as follows:

1. `map[int]int`
2. `map[interface{}]interface{}`
3. `github.com/marcelocantos/frozen/pkg/frozen` (this library's HAMT)
4. `github.com/mediocregopher/seq`

In all cases, ints are mapped to ints.

```
BenchmarkInsertMapInt0-8          	 5409126	       234 ns/op
BenchmarkInsertMapInt1M-8         	 5233239	       218 ns/op
BenchmarkInsertMapInterface0-8    	 2130547	       542 ns/op
BenchmarkInsertMapInterface1M-8   	 3082635	       642 ns/op
BenchmarkInsertFrozen0-8          	 1000000	      2085 ns/op
BenchmarkInsertFrozen1M-8         	  764451	      2232 ns/op
BenchmarkInsertMediocre0-8        	   14215	    127258 ns/op
BenchmarkInsertMediocre10k-8      	    7867	    257561 ns/op
```


![https://docs.google.com/spreadsheets/d/1Sq48pT4sKLHx2uY_nSljfbFpEJijXhNAeoB-BbDlrsI/edit?usp=sharing](https://docs.google.com/spreadsheets/d/e/2PACX-1vQHj9jtHqlQ_aHpTUTk9dQ_VfCw77_3QR4j5M76T-e6TtxhM77CfbvhDzB8IHKm29iP-L1TJqrPkOPa/pubchart?oid=301271241&format=image)

## Bugs

Test coverage is sparse.


[1]: https://en.wikipedia.org/wiki/Hash_array_mapped_trie
