---
status: implemented
date: 2026-05-05
description: flip `Error`'s string arg from regex to literal substring; close the silent-bite foot-gun
---

# Proposal: flip `Error`'s string arg from regex to literal substring

> Tactical fix for the foot-gun called out in the godep_differ migration
> proposal (#6). Sibling to the broader Error-family redesign discussion
> in [godep-differ-gaps.prop.md](godep-differ-gaps.prop.md); this is the
> minimal patch that closes the bug without redesigning the family.

## Gap

`assert.Error(t, err, "some pattern")` treats the string as a regex.
`.()?+` are metacharacters. Caller passing what looks like a literal
substring gets superset matching:

```go
// caller intent: err msg contains literal "want a+b"
assert.Error(t, err, "want a+b")
// impl returns err.Error() == "want aaab" (typo, doubled a)
// regex `a+` matches `aaa` -> test green -> typo ships
```

[_error-string-regex-docs.prop.md](_error-string-regex-docs.prop.md)
documented this in v0.5; downstream users still trip over it because
the docs aren't on screen at the call site -- the type/name has to
carry the meaning.

Empirical data across ~85 downstream files using `lczyk/assert`:

- 1 string-arg call (`goruby/trace_test.go: "not walkable"`).
- 0 regex-via-string calls in real downstream code.
- 4 `*regexp.Regexp`-arg calls, all inside `assert/`'s own tests.

Regex-via-string is not earning its keep. The footgun is paying
nothing for the risk it carries.

## Options

- **(a)** drop string branch entirely, runtime-panic on string arg.
  Caller picks `*regexp.Regexp` (regex) or new `ErrorContains`
  (literal). Loud break, forces explicit choice.
- **(b)** flip string -> literal substring. Keep `*regexp.Regexp`
  branch as the regex-only path. Each match style maps to a distinct
  arg type; no coercion. Breaking but the break direction is safe
  (subset semantics, see below).
- **(c)** silently accept both interpretations via fallback chain.
  Rejected -- doubles the ambiguity instead of removing it.

## Why (b) is safe to ship as a silent-looking flip

Tests that change behaviour under this flip can only turn red, not
green. The new interpretation is a strict subset of the old:

- old: `re.MatchString(err.Error())` -- matches any err msg that
  matches the regex.
- new: `strings.Contains(err.Error(), s)` -- matches any err msg
  that contains the literal string. Identical to a regex with all
  metachars escaped.

Every literal match is also a regex match (regex `s` w/ no metachars =
substring match). The reverse is not true: regex `a+` matches `aaa`,
literal `a+` does not. So:

- callers passing literal-intent strings (no metachars): identical
  behaviour pre/post flip. No bite.
- callers passing regex-intent strings: tests that previously matched
  via regex superset-behaviour turn red. Caller forced to investigate
  -- loud, not silent. Caller migrates to `*regexp.Regexp`.

The truly dangerous flip would be the other direction (literal ->
regex, expanding match set, tests turning green wrongly). This flip
is the safe direction.

## Modes after the flip

```go
Error(t, err, nil)                    // err == nil
Error(t, err, AnyError)               // err != nil
Error(t, err, errVal)                 // structural match + errors.Is
Error(t, err, "literal substring")    // strings.Contains  (FLIPPED)
Error(t, err, regexp.MustCompile(p))  // regex (unchanged)
```

Caller intent encoded in arg type. Type switch dispatches. No
ambiguity, no silent-bite path.

## What stays out of scope

- **`ErrorContains` named alias.** Skipped for now. `Error(t, err,
  "sub")` reads fine post-flip; a separate name doesn't earn its
  surface. Add later iff discoverability complaints emerge.
- **`ErrorMatches` named alias.** Skipped. 0 real downstream regex
  callers; existing `*regexp.Regexp` path covers the few that may
  appear.
- **Dropping `AnyError` sentinel / making `Error` single-purpose.**
  Out of scope here -- that's the broader redesign in
  [godep-differ-gaps.prop.md](godep-differ-gaps.prop.md). This
  proposal stays narrow.

## Open questions

- Update message wording: `"expected error to match 'X'"` ->
  `"expected error to contain 'X'"`? Reads more accurately for
  literal substring.
- `regexCache` + `compile_cached` become dead code after this flip
  (only the string branch used them; `*regexp.Regexp` callers compile
  upstream). Remove or keep for future regex caching needs?
- Demo / example tests reference the regex behaviour. Rewrite or
  remove?

## Migration cost

Across ~85 downstream files:

- 1 string-arg call -> works as-is post-flip (no metachars).
- 4 `*regexp.Regexp`-arg calls inside `assert/`'s own test suite:
  test the now-only-regex path, keep as-is.
- 0 real downstream regex-via-string callers: nothing to migrate.

Internal `assert/` tests / demos that exercise the flipped semantics:
rewrite to assert literal-substring behaviour.

## Outcome

Shipped in v0.6.0.

- `assert.go`: `case string:` in `assert_error` switched to
  `strings.Contains`. `regexCache` and `compile_cached` removed (only
  the string branch used them; `*regexp.Regexp` callers compile
  upstream).
- `interface.go`: `Error` doc updated to say "literal substring match"
  and to point at `*regexp.Regexp` for regex matching.
- Failure message wording: `"expected error to match 'X'"` ->
  `"expected error to contain 'X'"`.
- Tests: existing string-arg tests kept (all used non-metachar
  patterns and pass under both interpretations); added
  `TestErrorStringLiteralNotRegex` to lock in the new semantics
  (metachars treated literally, no superset matches).
- Demo: `TestDemoErrorRegex` renamed to `TestDemoErrorContains`.
- Decision left for later: `ErrorContains` named alias not shipped.
  `Error(t, err, "sub")` reads fine post-flip; revisit iff
  discoverability complaints emerge.

Open questions resolved:

- Message wording: switched to "contain".
- `regexCache` / `compile_cached`: removed (dead).
- Demo / examples: rewritten to reflect literal semantics.
