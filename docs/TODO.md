# TODO

## Bugs

- [ ] `lazy/set_diff.go:21` — `differenceSet.Has` uses `||` (union logic) instead of `&& !`
- [ ] `lazy/set_union.go:38` — `unionSet.FastHas` queries `s.a` twice, never `s.b`
- [ ] `lazy/set_diff.go:28` — `differenceSet.FastHas` queries `s.a` twice, never `s.b`

## Performance

- [ ] Analyse parallel ops for optimisation opportunities

## Simplification

- [ ] Remove the hash generator — assume 128-bit h128 will never produce false equality
