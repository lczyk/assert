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
// other than the one running the test still records the failure, but
// only that goroutine stops: whatever it was going to do next,
// including signalling the test that it is done, never happens, so a
// test waiting on it hangs unless the signal was deferred. Once the test
// function has returned, any assert.* or require.* call from a leftover
// goroutine panics inside the testing package ("Fail in goroutine after
// Test... has completed") and takes the whole test binary down. ALWAYS
// call require.* from the test goroutine. From background goroutines
// use assert.* and make the test wait for them, send the error back
// through a channel, or do the check on the test goroutine after the
// background work completes.
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
