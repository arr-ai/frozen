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

## 2026-03-01 — /audit

- **Commit**: `6bab8ab` (branch `readme-perf-v1.8.0`)
- **Outcome**: 34 findings (7 high, 2 medium, 19 low, 6 info). Report: `docs/audit-2026-03-01.md`.
- **Fixed since prior**: 5 items (CI triggers, rebase merge policy, CLAUDE.md inaccuracies, action versions).
- **High**: 3 lazy/ bugs still unfixed (#1-3: differenceSet.Has/FastHas, unionSet.FastHas) + 4 new: symmetricDifference cascading bug (#4), EmptySet.Equal self-comparison (#5), missing Has/FastHas test coverage (#6), baseSet.Hash non-deterministic (#7).
- **Medium**: Branch protection lacks required status checks (#8). Debug log.Print in Nest() (#9).
- **Deferred**: All low/info items pending triage.
