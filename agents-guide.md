# Frozen — agent guide

Immutable sets, maps, and integer sets for Go. Structural sharing; every
"mutation" returns a new value.

## Install

```
go get github.com/arr-ai/frozen/v2
```

```go
import "github.com/arr-ai/frozen/v2"
```

v1 remains `github.com/arr-ai/frozen` (last tag `v1.14.0`). Both modules can
appear in one build. New work should import `/v2`.

## Types

- `Set[T]`, `Map[K, V]`, `IntSet[I]` — hashed array tries
- Builders: `SetBuilder[T]`, `MapBuilder[K, V]`; call `Finish()`
- `lazy/` — deferred set ops with memoization
- `pkg/rel/` — relations on frozen maps/sets

Custom keys implement `frozen.Key[T]`: `Equal(T) bool` and `Hash() uintptr`
(no seed). `Set` and `Map` themselves implement `Key`, so they nest.

## Gotchas

- **v2 Hash is seedless.** `Hashable` is `Hash() uintptr`. v1 was
  `Hash(seed uintptr) uintptr`; `Hash128` is gone. Implement the new method;
  do not pass a seed into `Set.Hash` / `Map.Hash`.
- **Equal does not trust hashes.** Distinct sets that collide on `h0` still
  compare unequal. Do not short-circuit equality on a matching hash.
- **Defined types in `Set[any]`.** `int(5)` and `type ID int` (value 5) may
  share a hash slot; they remain distinct elements because `==` includes the
  dynamic type. Mixing `PkgPath` into the hash was measured and rejected
  (~18 ns extra on the reflection path).
- **Floats.** `+0` and `-0` hash alike; every NaN hashes differently, matching
  `NaN != NaN`.
- **Internal packages are not API.** Import `frozen/v2`, `frozen/v2/lazy`,
  or `frozen/v2/pkg/rel` only.

## Tests

```
go test -short ./...
go test -race -short ./...
go test -tags frozen_vet -short ./...
```
