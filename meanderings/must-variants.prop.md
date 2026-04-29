---
status: open
date: 2026-04-29
description: `Must*` hard-fail variants using `t.Fatalf`
---

# Proposal: `Must*` hard-fail variants

## Gap

Rust `googletest` splits `assert_*` (hard fail, abort test) vs
`expect_*` (soft fail, continue). Go's `t.Errorf` is already soft;
there is no fail-fast path in this package.

Add `MustEqual`, `MustNoError`, etc. using `t.Fatalf` for hard-stop
semantics — useful when later assertions only make sense if an earlier
one passed (e.g. `MustNoError(err); use(result)`).

## Open questions

- Does every assertion get a `Must*` sibling, or only a curated subset
  (`MustNoError`, `MustNotNil`, `MustEqual`, `MustType`)?
- `Type[T]` already returns the asserted value; `MustType` is the
  natural "use the result" form.
- Naming: `Must*` matches Go stdlib (`regexp.MustCompile`); `Require*`
  matches testify.
