---
status: implemented
date: 2026-04-29
description: `AnyError` as distinct sentinel type, not the surprising `""` regex
---

# Proposal: `AnyError` as a distinct sentinel type

## Gap

`assert.Error(t, err, "")` matched any non-nil error because the empty
regex matches everything. Reads at the call site as "expected error
message is the empty string" — opposite of what it does.

## Options considered

- A separate `AnyError` sentinel of distinct type, e.g. `var AnyError = errAny{}`.
- Or `assert.AnyError(t, err)` as a dedicated helper.

## Outcome

Shipped the sentinel: `var AnyError error = anyErr{}` in [interface.go](../interface.go).
`assert_error` checks `expected == AnyError` explicitly; empty string is
no longer the recommended spelling for "any error".
