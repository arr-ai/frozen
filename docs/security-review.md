# Security Review: `hamt-optimization` branch

**Date**: 2026-02-25
**Commit**: `cdd2511` (v1.7.0 + 12 commits)
**Scope**: All changes on `hamt-optimization` vs `master`

## Result: No vulnerabilities found

## Attack Surface Assessment

This is a pure in-memory data structure library with **zero external attack surface**:

- No network I/O
- No file I/O
- No user input handling
- No deserialization of untrusted data
- No command execution
- No authentication or authorization
- No cryptographic operations
- No database access

## `unsafe.Pointer` Review

The PR introduces `unsafe.Pointer` usage in two files for performance-critical type dispatch:

### `internal/pkg/tree/hasher.go`

- **`typeKey[T]()`** (lines 81-85): Extracts the type-descriptor pointer from an eface to use as a `sync.Map` key. Well-known Go runtime pattern. The pointer is immediately dereferenced to `uintptr` and used only as a map key.
- **`resolveHashFunc[T]()`** (lines 25-64): Reinterprets generic type parameters as concrete scalar types via `*(*uintNN)(unsafe.Pointer(&key))`. Guarded by type switches and `unsafe.Sizeof` checks that guarantee size matches before casts. Closures are cached once per type.

### `internal/pkg/value/value.go`

- **`equalScalar[T]()`** (lines 93-113): Same size-checked scalar reinterpretation pattern for equality comparison.
- **`EqualFuncFor[T]()`** (lines 47-77): Uses `defer recover()` to test comparability — documented Go pattern that only determines dispatch path.

**Verdict**: All `unsafe` usages are type-safe by construction (guarded by type switches or size checks), operate only on stack-local variables, and accept no external input.

## Categories Reviewed

| Category | Applicable | Findings |
|---|---|---|
| Input validation (SQLi, CMDi, path traversal) | No | N/A |
| Authentication / authorization | No | N/A |
| Crypto / secrets management | No | N/A |
| Injection / code execution | No | N/A |
| Data exposure | No | N/A |
| Unsafe pointer misuse | Yes | None — all usages properly guarded |
