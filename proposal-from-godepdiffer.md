# proposal: gaps in `lczyk/assert` surfaced by godep_differ migration

context: migrated ~46 test files / ~27k lines from std-lib `t.Errorf` patterns to `github.com/lczyk/assert@v0.5.0`. these are the concrete gaps that forced fallback to manual style or noisy boilerplate.

ranked roughly by frequency-of-pain in this repo.

---

## 1. no `Fatal`-variant of any assertion

`assert.*` only calls `t.Errorf`. test continues after failure. every site that asserts then dereferences needs a manual guard:

```go
assert.NoError(t, err)
if err != nil { t.FailNow() }       // every. single. time.

assert.NotNil(t, client)
if client == nil { return }
```

across this repo that's ~200+ sites. boilerplate dwarfs the assertion.

**proposal:** either
- `assert.MustNoError`, `assert.MustNotNil`, ... -- parallel `Must*` family that calls `t.Fatalf` (named after stdlib `Must*` convention).
- or a per-call sentinel: `assert.NoError(t, err, assert.Fatal)` -- single extra arg flips the failure mode.

`Must*` is cleaner imo -- name signals stop-on-fail; calling convention stays positional. fits the `Must*` mental model from `regexp.MustCompile`, `template.Must`.

## 2. `Equal` is `comparable`-only -- forces variant-juggling

current api makes the caller pick by type:

- scalar / string / pointer-equality -> `Equal`
- `[]T` of comparable -> `EqualArrays`
- `map[K]V` of comparable -> `EqualMaps`
- `[]T` unordered comparable -> `EqualArraysUnordered`
- struct / nested / non-comparable -> `EqualCmp(t, a, b, func(x, y T) bool { return reflect.DeepEqual(x, y) })`

last one shows up dozens of times. pure noise -- caller already knows they want deep-equal.

**proposal:** add `assert.DeepEqual(t, a, b)` (reflect-based, works for any type). keep the typed variants for the perf-sensitive case but make the deep-equal path one call.

bonus: a single primitive collapses the decision tree -- "is this comparable? is the inner type comparable? is the slice ordered?" -- which is real cognitive load when migrating in bulk.

## 3. no ordering primitives (`Greater` / `Less` / `Between` / `GreaterOrEqual`)

threshold checks are common. score / count / length / duration assertions all fall back to `assert.That`:

```go
assert.That(t, score >= 0.4, "score below threshold: got %v", score)
assert.That(t, count > 4, "...")
assert.That(t, edgesAfter <= edgesBefore, "...")
```

`That` works but loses structured diagnostic ("got %v, want >= %v") and the call site has to hand-format the message.

**proposal:** add at least
- `assert.Greater(t, got, threshold)`
- `assert.GreaterOrEqual(t, got, threshold)`
- `assert.Less(t, got, threshold)`
- `assert.LessOrEqual(t, got, threshold)`

generic over `cmp.Ordered`. the auto-generated message ("got %v, want > %v") is the win.

`Between(t, x, lo, hi)` is nice-to-have for range checks.

## 4. no `True` / `False`

`assert.That(t, ok)` works but reads oddly for boolean predicates. muscle memory from other libs reaches for `True`. and `That` requires at least a message arg in spirit (it's variadic but bare `That(t, ok)` produces a sparse failure message).

**proposal:** trivial wrappers `assert.True(t, b)` / `assert.False(t, b)`. zero implementation cost, big readability win in test code.

## 5. no `Slice/Map` containment primitives

repeating shapes:

```go
slices.Contains(s, v)                  // slice contains value
_, ok := m[k]; ok                      // map has key
m[k] == want                           // map key has value
```

all fall back to `assert.That` + manual message. especially common in fixer tests (`result["error"].(string)` + substring match).

**proposal:**
- `assert.Contains[T comparable](t, []T, T)` -- slice membership
- `assert.HasKey[K comparable, V any](t, map[K]V, K)` -- map key presence
- `assert.MapEqual` already covered by `EqualMaps`, but `MapContains(t, m, k, v)` for single-pair check would help.

## 6. `Error` regex-vs-substring overload is a foot-gun

current: `assert.Error(t, err, "some pattern")` -- the string is treated as a regex. `.()?+` are metacharacters. caller passing a literal substring with parens / dots will get either false positives or `regexp` panics.

surprised me twice during migration. had to switch to `*regexp.Regexp` form explicitly to avoid ambiguity.

**proposal:** split into two named primitives:
- `assert.ErrorContains(t, err, substring)` -- literal `strings.Contains(err.Error(), s)`
- `assert.ErrorMatches(t, err, pattern)` -- explicit regex (`string` or `*regexp.Regexp`)

drop the implicit-regex branch from `Error`. structural-match (`error` arg) and `AnyError` stay.

## 7. asymmetric `args ...any` support

`NoError`, `Error`, `Nil`, `NotNil`, `Len`, `ContainsString`, `That` accept trailing `args ...any` for custom messages. `Equal`, `NotEqual`, `EqualArrays`, `EqualMaps`, `EqualCmp` do **not**.

inconsistent. on equality checks the caller often wants extra context ("for case %s") but has to drop back to `t.Errorf`.

**proposal:** add `args ...any` to every public assertion. uniform api == less api to remember.

(would also make `assert.Equalf`-style separate variants unnecessary -- variadic covers it.)

## 8. no float-tolerance comparison

score tests have epsilon comparisons:

```go
assert.That(t, math.Abs(got-want) < 1e-6, "...")
```

**proposal:** `assert.NearlyEqual(t, got, want, tolerance)` or `assert.Within(t, got, want, delta)`. covers floats and any `cmp.Ordered` numeric.

## 9. `EqualArraysUnordered` is `comparable`-only

works for `[]string` / `[]int`. doesn't work for `[]struct{...}` / `[]MyType` even when the struct is itself comparable in some sense (deep-equal). these slices stayed manual.

**proposal:** `assert.ElementsMatch(t, a, b)` -- reflect-based unordered slice equality. mirrors testify's name.

## 10. `Type[T]` extraction has same Fatal-gap as #1

```go
client := assert.Type[*GeminiClient](t, c)  // returns *GeminiClient
client.DoThing()                            // nil-deref if assert failed
```

since assert doesn't stop the test, `client` may be the zero value. caller must guard manually.

**proposal:** `Type[T]` is one of those primitives that almost always wants Fatal semantics -- consider making it Fatal-by-default (or providing `MustType[T]` per #1).

## 11. `EqualLineByLine` doesn't show a diff

current output: per-line "expected 'X' got 'Y'". for big-string compares (generated code, large config), no surrounding context, no unified-diff format. hard to spot the actual delta when tens of lines mismatch.

**proposal:** unified-diff output (`--- expected / +++ got`, with context lines). either replace `EqualLineByLine` body or add `assert.EqualDiff`. could vendor a small diff impl or use `github.com/pmezard/go-difflib`.

## 12. no `Eventually` / `WithinDuration` for timing-flaky checks

didn't bite this repo hard, but `internal/launchpad` and `integration_test.go` have a few "wait for state" loops. would clean those up.

**proposal:** `assert.Eventually(t, predicate func() bool, timeout, interval)` -- polls. `assert.WithinDuration(t, t1, t2, delta)` for timestamp checks.

low priority -- only relevant for integration-style tests.

---

## summary by priority

high-impact, low-cost:
- `Must*` family (#1) -- biggest boilerplate eliminator
- `True` / `False` (#4)
- ordering primitives (#3)
- uniform `args ...any` (#7)

high-impact, moderate-cost:
- `DeepEqual` (#2) -- one primitive replaces the typed variants in 90% of call sites
- `ErrorContains` / `ErrorMatches` split (#6) -- correctness-relevant

nice-to-have:
- `Contains` / `HasKey` (#5)
- `NearlyEqual` (#8)
- `ElementsMatch` (#9)
- `Type[T]` Fatal-by-default (#10)
- diff output for `EqualLineByLine` (#11)
- `Eventually` (#12)

if forced to pick three: `Must*`, `DeepEqual`, ordering primitives. those three together would have removed ~60% of the manual `t.FailNow()` / `assert.That` / `EqualCmp(reflect.DeepEqual)` boilerplate from the godep_differ migration.
