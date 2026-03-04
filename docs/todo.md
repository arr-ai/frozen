# TODO

- [ ] Ensure different types with the same underlying value hash differently (e.g., `int(1)` vs `uint(1)` vs custom `type MyInt int` should not collide)
- [ ] Add `reflect.DeepEqual` regression tests for `Set` and `Map` — verify that values built via different construction paths (e.g., `NewMap().With(...)` vs `NewMap().Update(m)`) remain `DeepEqual`. This regressed in v1.8.0–v1.9.0 due to an unexported func field in the tree struct.
