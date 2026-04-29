# Rust testing comparison

Notes on whether porting this package to Rust is worthwhile, and what
Rust-side ideas could improve the Go package.

## Q1: Port to Rust?

**No real point.** Rust's testing ecosystem already covers this surface
area more idiomatically:

- `assert!`, `assert_eq!`, `assert_ne!` — file/line free via `panic!`,
  no runtime caller-walk needed.
- `pretty_assertions` — colored diffs (supersedes `EqualLineByLine`).
- `assert_matches!` / `matches!` — pattern match including errors.
- `googletest-rust`, `assert2`, `claim` — `assert_that!`, `check_that!`,
  source-spanned failures.
- `#[should_panic(expected = "...")]` replaces `Panic`.

Go-specific motivations don't transfer:

- `runtime.Caller` walk → Rust macros capture span at compile time.
- `reflect`-based `isNil` / `Len` → Rust types make these compile-time
  (`Option::is_none()`, `.len()` on concrete types, no `any`).
- `errors.Is` wrap chain → `Result<T, E>` + `thiserror`/`anyhow` +
  `assert_matches!(res, Err(MyErr::Foo(_)))`.
- `EqualCmpAny` / `Type[T]` → Rust generics + traits make redundant.

Niche where it could matter: unified cross-language style in a polyglot
project, or a specific helper like regex-substring error match as one
macro. That's ~30 LOC of macro, not a crate.

**Recommendation:** skip the port. If needed, write a thin
`assert_err!(res, "regex")` macro inline (~20 LOC) and lean on
`pretty_assertions` + `assert_matches`.

## Q2: Rust ideas worth borrowing into the Go package

Ranked by payoff.

### 1. Source-span snippet in failure output
`assert2` / `googletest` print the failing expression source:
`assert!(x > 5)` → `assertion failed: x > 5`. Go can't get expression
text from `runtime.Caller`, but the source file is readable at
`file:line` — print the calling line. ~15 LOC. Big UX win for
`That(t, complicated && expr)` which currently only prints
`assertion failed`.

### 2. Diff output for collection mismatches
`EqualLineByLine` currently prints lines one-by-one. Better: unified
diff with `+`/`-` (color gated on `isatty`). Same treatment for
`EqualArrays` / `EqualMaps` on mismatch — show a diff, not just
`expected X got Y`.

Options:
- `github.com/pmezard/go-difflib` — line diff.
- `github.com/google/go-cmp` `cmp.Diff` — idiomatic Go answer; handles
  structs/maps with field-level diff. Probably the biggest single
  ergonomic jump. Consider `EqualDiff(t, a, b)`.

### 3. Hard-fail variants
Rust `googletest` splits `assert_*` (hard fail) vs `expect_*` (soft
fail, continue). Go's `t.Errorf` is already soft; there is no
fail-fast path in this package. Add `MustEqual`, `MustNoError`, etc.
using `t.Fatalf` for hard-stop semantics.

### 4. Panic-message regex shorthand
Current `Panic(t, f, f_recover)` requires a callback. Add
`Panics(t, f, "regex")` that matches the panic message directly,
mirroring Rust's `#[should_panic(expected = "...")]`. Trivial wrapper
over existing `Panic`.

### Considered but not worth it

- **Pattern-match assertion** (`assert_matches!`): Go has no patterns;
  existing `EqualCmp` with closure is close enough.
- **Parameterized tests** (`rstest`, `proptest`): out of scope —
  `t.Run` subtests already cover it.
- **Option ergonomics**: N/A in Go; reflect-based `isNil` is the
  necessary cost.
- **Parallel isolation**: both languages handle it (`t.Parallel()`).

## Concrete shortlist

1. Source-line snippet at `file:line` in failures.
2. `cmp.Diff`-based output for `EqualArrays` / `EqualMaps` /
   `EqualLineByLine`, plus a new `EqualDiff`.
3. `Must*` variants using `t.Fatalf`.
4. `Panics(t, f, "regex")` shorthand.

Items 1 and 2 deliver most of the perceived gap with Rust tooling.
