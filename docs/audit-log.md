# Audit Log

Chronological record of audits, releases, documentation passes, and other
maintenance activities. Append-only — newest entries at the bottom.

## 2026-02-28 — /release v1.8.0

- **Commit**: `3d0b250`
- **Outcome**: Released v1.8.0. Added STABILITY.md with full interaction surface catalogue for future breaking-change audits. No breaking changes since v1.7.0 — all changes were internal optimizations (h0 recursive XOR hash, HAMT allocation reduction, CI lint upgrade).

## 2026-03-02 — /release v1.9.0

- **Commit**: `6bc3bbe`
- **Outcome**: Released v1.9.0. H128 single-call 128-bit AES hash internalized from arr-ai/hash, eliminating external dependency. Added `frozen.Hashable` interface (additive, non-breaking). NOTICES file added for vendored code attribution. Breaking change audit passed — no removals or signature changes.
