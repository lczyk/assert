// Package require is the hard-fail sibling of assert. Every primitive
// has the same signature as its assert counterpart but calls
// [testing.TB.Fatalf] on failure rather than [testing.TB.Errorf]. A
// failed require.* call aborts the test goroutine immediately, so the
// caller can safely use any value the assertion returned (relevant for
// [Type]).
//
// # Goroutine hazard
//
// testing.T.Fatalf calls [runtime.Goexit], which only unwinds the
// goroutine it is called from. A failed require.* call from a goroutine
// other than the one running the test will not stop the test, will not
// be reported, and will leave the goroutine half-dead. The test may then
// pass silently despite the failed assertion. ALWAYS call require.*
// from the test goroutine. From background goroutines use assert.* (the
// failure is reported via t.Errorf, which is goroutine-safe), capture
// the error and surface it back through a channel, or do the check on
// the test goroutine after the background work completes.
//
// # When to use which
//
//   - assert.* -- the failure is informative but the rest of the test
//     can still run usefully. Multiple unrelated assertions in one test
//     body.
//   - require.* -- the assertion is a precondition for the rest of the
//     test. Failing it makes everything after meaningless (or unsafe to
//     run, e.g. dereferencing a value the assertion returned).
//
// Pattern: require.NoError on setup, assert.* on the actual behaviour
// under test. Pre-condition / behaviour split.
package require
