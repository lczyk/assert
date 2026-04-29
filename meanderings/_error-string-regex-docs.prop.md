---
status: implemented
date: 2026-04-29
description: document `Error`'s string = regex-substring match semantics
---

# Proposal: document `assert.Error`'s string = regex-substring semantics

## Gap

When `expected` is a `string`, `assert.Error` does
`regexp.MustCompile(expected).MatchString(err.Error())` — a regex
*substring* match, not equality.

Two consequences:

- A test that expects exactly `"oops"` passes on `"oops happened, then more oops"`.
- Special characters in messages (`(`, `)`, `.`, `?`, etc.) silently
  change semantics.

`utils.AssertError` had the same behaviour, so this is not a regression
— but the docstring on `Error` did not state it.

## Options considered

- Document the behaviour explicitly on `Error`.
- Add `assert.ErrorMessage(t, err, exact string)` for exact equality.

## Outcome

Documented on `Error`'s doc comment in [interface.go](../interface.go),
including a note on regex metacharacters and the `^...$` anchor advice
for exact-equality. No `ErrorMessage` helper added — anchors cover it.
