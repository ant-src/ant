<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.skip

```ant signature
func Test.skip(reason: string): never
```

Test.skip ends the current test as skipped, after its deferred blocks run. It is not a panic, so
Panic.catch does not stop it. A test that failed before stays failed.

<!-- end generated -->

## Details

- `Test.skip` returns `never`, so it fits anywhere a value is expected, such as after `??`.
- It acts on the current test, so a helper function can call it without being given anything. In a
  subtest it skips only that subtest. Outside any test it panics.
- A test that failed before the call stays failed. In a deferred block that runs while a panic
  unwinds the test, the call is dropped and the test fails.
- Skipped tests do not fail the run: `ant test` exits with status 0 when no test failed.

## Basic examples

Skip a test when the environment lacks what it needs:

<!-- test: cmd ant test env_test.ant -->

```ant [env_test.ant]
test "connects to the sandbox" {
  let key = Os.env("PAYMENTS_SANDBOX_KEY") ?? Test.skip("PAYMENTS_SANDBOX_KEY is not set")
  assert(key.length() > 0)
}

test "adds" {
  assert(1 + 1 == 2)
}
```

The summary line counts the skipped test:

```text
ok    1 passed, 1 skipped (0.00s)
```

Under `ant test -v`, each skipped test shows its reason:

<!-- test: cmd ant test -v reason_test.ant -->

```ant [reason_test.ant]
test "slow import" {
  Test.skip("needs the full data set")
}
```

```text
=== RUN   reason_test.ant: slow import
--- SKIP: reason_test.ant: slow import (0.00s)
    needs the full data set
ok    1 skipped (0.00s)
```

## Scope

### Deferred blocks

The deferred blocks of the test run before it ends. Nothing after the call runs:

<!-- test: cmd ant test -v defer_test.ant -->

```ant [defer_test.ant]
test "cleans up" {
  defer { println("cleanup") }
  Test.skip("not ready")
  println("never printed")
}
```

```text
=== RUN   defer_test.ant: cleans up
    cleanup
--- SKIP: defer_test.ant: cleans up (0.00s)
    not ready
ok    1 skipped (0.00s)
```

### Subtests

In a subtest, `Test.skip` skips that subtest. The other rows of the table still run, and the parent
passes:

<!-- test: cmd ant test -v rows_test.ant -->

```ant [rows_test.ant]
test "square" {
  for n in [1, 2, 3] {
    test "${n}" {
      if n == 2 { Test.skip("two is covered elsewhere") }
      assert(n * n >= n)
    }
  }
}
```

```text
=== RUN   rows_test.ant: square
=== RUN   rows_test.ant: square/1
--- PASS: rows_test.ant: square/1 (0.00s)
=== RUN   rows_test.ant: square/2
--- SKIP: rows_test.ant: square/2 (0.00s)
    two is covered elsewhere
=== RUN   rows_test.ant: square/3
--- PASS: rows_test.ant: square/3 (0.00s)
--- PASS: rows_test.ant: square (0.00s)
ok    2 passed, 1 skipped (0.00s)
```

### Goroutines

Called on a goroutine the test owns, `Test.skip` ends that goroutine and marks the test skipped. The
test's body goes on:

<!-- test: cmd ant test -v go_test.ant -->

```ant [go_test.ant]
test "worker" {
  let wg = WaitGroup.new()
  wg.add(1)
  go {
    defer { wg.done() }
    Test.skip("the worker has nothing to do")
  }
  wg.wait()
  println("body goes on")
}
```

```text
=== RUN   go_test.ant: worker
    body goes on
--- SKIP: go_test.ant: worker (0.00s)
    the worker has nothing to do
ok    1 skipped (0.00s)
```

## Applications

### A helper that skips

A helper function can skip the test that calls it, with no handle passed in. Since `Test.skip`
returns `never`, the helper returns a plain `string`:

<!-- test: cmd ant test helper_test.ant -->

```ant [helper_test.ant]
func requireEnv(name: string): string {
  Os.env(name) ?? Test.skip("${name} is not set")
}

test "reads the bucket" {
  let bucket = requireEnv("STORAGE_TEST_BUCKET")
  assert(bucket.startsWith("test-"))
}
```

```text
ok    1 skipped (0.00s)
```

### Optional fixtures

Skip a test when a large fixture is not checked out, and run it when it is:

<!-- test: file testdata/small.csv
id,name
1,Ann
-->

<!-- test: cmd ant test fixture_test.ant -->

```ant [fixture_test.ant]
func lineCount(path: string): int {
  if !File.exists(path) { Test.skip("${path} is missing") }
  File.read(path).unwrap().trim().split("\n").length()
}

test "small fixture" {
  assert(lineCount("testdata/small.csv") == 2)
}

test "large fixture" {
  assert(lineCount("testdata/large.csv") > 1000)
}
```

```text
ok    1 passed, 1 skipped (0.00s)
```

## Possible issues

### Outside a test

`Test.skip` acts on the current test. Called in a script, or anywhere outside a test, it panics:

```ant
if !File.exists("config.json") {
  Test.skip("config.json is missing") // panic: Test.skip: not in a test
}
println("configured")
```

### `Panic.catch` does not stop it

`Test.skip` is not a panic, so `Panic.catch` neither returns nor lets the test go on:

<!-- test: cmd ant test -v catch_test.ant -->

```ant [catch_test.ant]
test "caught" {
  let r = Panic.catch({ Test.skip("skipped anyway") })
  println("never printed: ${r}")
}
```

```text
=== RUN   catch_test.ant: caught
--- SKIP: catch_test.ant: caught (0.00s)
    skipped anyway
ok    1 skipped (0.00s)
```

### A failed test stays failed

Skipping cannot undo a failure. Here a subtest has failed, so its parent fails even though it calls
`Test.skip` afterwards, and `ant test` exits with status 1:

<!-- test: cmd ant test failed_test.ant -->
<!-- test: exit 1 -->

```ant [failed_test.ant]
test "parent" {
  test "child" {
    assert(1 + 1 == 3)
  }
  Test.skip("too late")
}
```

```text
--- FAIL: failed_test.ant: parent/child (0.00s)
    assertion failed: 1 + 1 == 3 (2 == 3)
      at test "child" (failed_test.ant:3:11)
FAIL  1 failed, 0 passed (0.00s)
```

A deferred block cannot turn a failure into a skip either. While a panic unwinds the test, the call
is dropped and the panic goes on:

<!-- test: cmd ant test unwind_test.ant -->
<!-- test: exit 1 -->

```ant [unwind_test.ant]
test "flaky" {
  defer { Test.skip("ignore failures") }
  assert(2 * 2 == 5)
}
```

```text
--- FAIL: unwind_test.ant: flaky (0.00s)
    assertion failed: 2 * 2 == 5 (4 == 5)
      at test "flaky" (unwind_test.ant:3:9)
FAIL  1 failed, 0 passed (0.00s)
```

## See also

- [test](../../language/keywords/test.md): declare tests and subtests
- [Test.parallel](parallel.md), [Test.tempDir](tempDir.md), [Test.context](context.md),
  [Test.name](name.md)
- [Panic.catch](../Panic/catch.md), [defer](../../language/keywords/defer.md)
- [ant test](../../tools/test.md): `-skip regexp` leaves out tests by name; they are neither run nor
  counted
