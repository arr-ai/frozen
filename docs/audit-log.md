# Audit Log

Chronological record of audits, releases, documentation passes, and other
maintenance activities. Append-only — newest entries at the bottom.

## 2026-02-28 — /release v1.8.0

- **Commit**: `3d0b250`
- **Outcome**: Released v1.8.0. Added STABILITY.md with full interaction surface catalogue for future breaking-change audits. No breaking changes since v1.7.0 — all changes were internal optimizations (h0 recursive XOR hash, HAMT allocation reduction, CI lint upgrade).

## 2026-02-28 — /audit

- **Commit**: `f7e494c` (branch `readme-perf-v1.8.0`)
- **Outcome**: 35 findings (2 high, 4 medium, 24 low, 5 info). Report: `docs/audit-2026-02-28.md`.
- **High**: 2 bugs in `lazy/` package (differenceSet.Has wrong logic, unionSet.FastHas copy-paste error).
- **Medium**: 3 CI/policy (no PR trigger, no required checks, rebase merge allowed).
- **Deferred**: All low/info items pending triage.
