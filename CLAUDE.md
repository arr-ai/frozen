# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`github.com/arr-ai/frozen` is an immutable data structures library for Go, built on hashed array tries with Go 1.19+ generics. Core types: `Set[T]`, `Map[K, V]`, `IntSet[I]`. Go module version requires 1.19, CI tests against 1.19 and 1.23.

## Commands

```bash
# Build
go build ./...

# Test (runs both -short and full)
make test

# Single test
go test -run TestName ./...

# Test with coverage
TESTFLAGS='-cover' make ci

# Lint (requires Docker; uses golangci-lint v1.60.1)
make lint

# Benchmarks with profiling
make bench

# Race detection
go test -race -short ./...

# Structural vet (validates tree invariants via frozen_vet build tag)
make vet

# Single benchmark
go test -run=^$ -bench=BenchmarkName ./...
```

### Build Tags

- **`frozen_vet`** — enables structural tree verification after each operation (used by `make vet`)
- **`branch4`** / **`branch16`** — override default fanout (fanout 4 / fanout 16)

## Architecture

### Core Types (root package)

- **`Set[T]`** — immutable set backed by `tree.Tree[T]`; supports Union, Intersection, Difference, Powerset, Where, Reduce
- **`Map[K, V]`** — immutable map backed by `tree.Tree[mapEntry[K, V]]`
- **`IntSet[I]`** — specialized integer set using bitmap compression (`Map[I, cellMask]`)
- **`Key[T]`** — constraint interface combining `value.Equaler[T]` and `hash.Hashable`; required for Map keys and Set elements that need custom equality
- Builders (`SetBuilder[T]`, `MapBuilder[K, V]`) for efficient incremental construction → call `Finish()` to produce immutable result

### Internal Packages (`internal/pkg/`)

- **`tree`** — hashed array trie: `Tree[T]` with node types (`branch`, `leaf1`, `leaf2`, `twig`) and `packer` (sparse array with 256-fanout). Core operations: `Combine`, `Difference`, `Intersection`, `Equal`. Also `Builder[T]`.
- **`depth`** — `Gauge` type controlling parallelism depth; operations go parallel when tree is deep enough
- **`value`** — generic equality via `Equal[T]()` dispatching through `Equaler[T]`, `Samer`, or `==`
- **`iterator`** — bit iterator and slice iterator implementations

### Public Subpackages

- **`lazy/`** — lazy Set interface for deferred evaluation with memoization
- **`pkg/rel/`** — relational algebra (Tuple, Relation, Join, CartesianProduct, Project) built on frozen Map/Set

## Key Conventions

- Immutable-by-design: all mutations produce new values via structural sharing
- Hashing uses `github.com/arr-ai/hash` with seed-based `hash.Any(value, seed)`
- External test packages (`frozen_test`, etc.)
- Line length limit: 120 characters
- Imports grouped with `goimports` local prefix: `github.com/arr-ai/frozen`
- Linter config in `.golangci.yml` enables 60+ linters with strict settings
