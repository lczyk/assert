---
status: open
date: 2026-04-29
description: `DeepEqual` for non-comparable types via `reflect.DeepEqual`
---

# Proposal: `Equal` for non-comparable types via `reflect.DeepEqual`

## Gap

`assert.Equal` requires `comparable`, which excludes slices, maps, and
structs containing them. Today the workaround is `EqualArrays` /
`EqualMaps` / `EqualCmp`.

A generic `assert.DeepEqual[T any](t, a, b T)` (or letting `Equal` fall
back to `reflect.DeepEqual` for non-comparable `T` via constraints)
would remove per-shape helpers.

The original `utils.AssertDeepEqual` was commented out, suggesting the
same itch.

## Open questions

- Separate function (`DeepEqual`, explicit) or constraint relaxation on
  `Equal` (implicit)? The latter would change generic-instantiation
  behaviour for existing callers.
- If `DeepEqual` lands, do `EqualArrays` / `EqualMaps` become aliases or
  stay as type-checked-at-compile-time variants?
