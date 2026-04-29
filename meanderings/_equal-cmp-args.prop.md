---
status: implemented
date: 2026-04-29
description: `EqualCmp` accepting `args ...any` for custom failure messages
---

# Proposal: `EqualCmp` accepting message args

## Gap

`utils.AssertEqualCmp` and `assert.EqualCmp` both lacked a variadic
`args ...any` message tail (unlike `Equal`, `That`, etc.). When a
custom comparator failed, the default
`expected '%v' (%T) == '%v' (%T)` message was often unhelpful for
domain types.

## Outcome

`EqualCmp` and `EqualCmpAny` now take `args ...any` ([interface.go](../interface.go)).
API is unified with the rest of the package.
