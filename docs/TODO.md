# TODO

## Bugs

### lazy/ (deferred — package is parked, fix when revisited)

- [ ] `lazy/set_diff.go:21` — `differenceSet.Has` uses `||` (union logic) instead of `&& !`
- [ ] `lazy/set_union.go:38` — `unionSet.FastHas` queries `s.a` twice, never `s.b`
- [ ] `lazy/set_diff.go:28` — `differenceSet.FastHas` queries `s.a` twice, lacks difference semantics
- [ ] `lazy/set_symmdiff.go:4` — `symmetricDifference` inherits differenceSet.Has bug
- [ ] `lazy/set_empty.go:41` — `EmptySet.Equal` compares argument to itself (always true)
- [ ] `lazy/set_base.go:78` — `baseSet.Hash` uses order-dependent chaining instead of XOR
- [ ] Missing `Has`/`FastHas`/`Equal` tests on composite lazy sets

## Performance

- [ ] Analyse parallel ops for optimisation opportunities

## Simplification

- [ ] Remove the hash generator — assume 128-bit h128 will never produce false equality
