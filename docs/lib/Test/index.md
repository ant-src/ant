<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test

Test holds what a test body needs beyond assert, Panic.catch, and defer (ant test). Each function
acts on the current test: the one whose body the calling goroutine runs, or that owns the goroutine.
Called outside any test, each panics.

<!-- end generated -->

A test checks its results with `assert` and cleans up with `defer`. The `Test` namespace adds the
rest: skip a test, run it alongside others, give it a scratch directory or a context that is done
when the test's body ends, and ask for its name. Each function finds the test it acts on by itself,
so a helper function calls them without being given a handle.

## Details

- The current test is the test whose body runs the call, or the test that owns the goroutine that
  makes it. A goroutine started while a test runs belongs to that test, and so do the goroutines it
  starts. In a subtest, each function acts on the subtest.
- Called outside any test, each function panics. That includes a script and the initialization of a
  test file's top-level variables, which runs before any test.
- `Test.skip` ends the test, but it is not a panic: `Panic.catch` does not stop it.

## Basic examples

Give a test a scratch directory with `Test.tempDir`, and skip a test that cannot run here with
`Test.skip`:

<!-- test: cmd ant test basic_test.ant -->

```ant [basic_test.ant]
test "saves a note" {
  let path = Path.join([Test.tempDir(), "note.txt"])
  try File.write(path, "buy milk")
  assert(try File.read(path) == "buy milk")
}

test "uploads a note" {
  if Os.env("NOTES_API_TOKEN") == null { Test.skip("NOTES_API_TOKEN is not set") }
}
```

```text
ok    1 passed, 1 skipped (0.00s)
```

`Test.name` returns the full name of the current test, its names joined with `/`. Under
`ant test -v`, what a test prints shows below its `=== RUN` line:

<!-- test: cmd ant test -v name_test.ant -->

```ant [name_test.ant]
test "parse" {
  for input in ["1", "0x10"] {
    test "${input}" {
      println(Test.name())
    }
  }
}
```

```text
=== RUN   name_test.ant: parse
=== RUN   name_test.ant: parse/1
    parse/1
--- PASS: name_test.ant: parse/1 (0.00s)
=== RUN   name_test.ant: parse/0x10
    parse/0x10
--- PASS: name_test.ant: parse/0x10 (0.00s)
--- PASS: name_test.ant: parse (0.00s)
ok    2 passed (0.00s)
```

## Functions by job

### Ending a test early

- [Test.skip](skip.md): end the current test as skipped, with a reason

### Running tests at once

- [Test.parallel](parallel.md): run the current test alongside the other parallel tests of its
  parent

### Resources that end with the test

- [Test.tempDir](tempDir.md): a new directory, removed with its contents when the test ends
- [Test.context](context.md): a context that is done when the test's body ends

### Asking about the test

- [Test.name](name.md): the current test's full name, such as `"parse/0x10"`

## Applications

### A fixture helper

A helper makes its own scratch directory, so each test that calls it gets a fresh file and nothing
has to remove it:

<!-- test: cmd ant test fixture_test.ant -->

```ant [fixture_test.ant]
func fixture(text: string): Result[string, Error] {
  let path = Path.join([Test.tempDir(), "input.txt"])
  try File.write(path, text)
  Result.ok(path)
}

func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.trim().split("\n").length())
}

test "counts lines" {
  let path = try fixture("a\nb\nc\n")
  assert(try countLines(path) == 3)
}

test "a missing file is an error" {
  let path = Path.join([Test.tempDir(), "missing.txt"])
  assert(countLines(path) is Err)
}
```

```text
ok    2 passed (0.00s)
```

### A table whose rows run at once

Each row of a table is a subtest. A row that calls `Test.parallel` runs alongside the other rows,
and its `Test.tempDir` is its own:

<!-- test: cmd ant test rows_test.ant -->

```ant [rows_test.ant]
test "upper" {
  for word in ["ant", "bee", "cat"] {
    test "${word}" {
      Test.parallel()
      let path = Path.join([Test.tempDir(), "${word}.txt"])
      try File.write(path, word.toUpper())
      assert(try File.read(path) == word.toUpper())
    }
  }
}
```

```text
ok    3 passed (0.00s)
```

### A worker that ends with the test

A goroutine the test starts gives `Test.context()` to its blocking calls, so it ends when the test's
body does instead of outliving the test:

<!-- test: cmd ant test worker_test.ant -->

```ant [worker_test.ant]
test "the worker answers" {
  let requests: Channel[string] = Channel.new()
  let replies: Channel[string] = Channel.new()
  go {
    let ctx = Test.context()
    while true {
      let request = requests.receive(ctx) ?? break
      replies.send(request.toUpper())
    }
  }
  requests.send("ping")
  assert(replies.receive() == "PING")
}
```

```text
ok    1 passed (0.00s)
```

## Properties and relations

A goroutine the test starts acts on the same test, so `Test.name` and `Test.context` give the same
answer there as in the body:

<!-- test: cmd ant test owned_test.ant -->

```ant [owned_test.ant]
test "worker" {
  let ctx = Test.context()
  let name = Test.name()
  let seen: Channel[bool] = Channel.new(1)
  go { seen.send(Test.context() == ctx && Test.name() == name) }
  assert(seen.receive() == true)
}
```

```text
ok    1 passed (0.00s)
```

## Possible issues

### Outside a test

In a script, no test is current, and each function panics:

```ant
let dir = Test.tempDir() // panic: Test.tempDir: not in a test
```

### Top-level variables of a test file

A test file initializes its top-level variables before any test runs, so a `Test` call there panics.
The file's tests do not run, the file counts as one failed test, and `ant test` exits with status 1:

<!-- test: cmd ant test init_test.ant -->

```ant [init_test.ant]
let scratch = Test.tempDir() // panic: Test.tempDir: not in a test

test "uses the scratch directory" {
  assert(File.exists(scratch))
}
```

```text
FAIL  1 failed, 0 passed (0.00s)
```

Call `Test.tempDir` in each test instead, or in a helper the tests call.

### Test.context in a deferred block

The test's context is done when its body ends, before its deferred blocks run, so a wait in a
deferred block that takes `Test.context()` returns at once. Cleanup that has to wait makes a context
of its own:

<!-- test: cmd ant test cleanup_test.ant -->

```ant [cleanup_test.ant]
test "the worker flushes its log" {
  let log: Channel[string] = Channel.new(1)
  go {
    Time.sleep(20ms)
    log.send("written")
  }
  defer {
    let ctx = Context.new(timeout: 1s)
    assert(log.receive(ctx) == "written")
  }
}
```

```text
ok    1 passed (0.00s)
```

See [Test.context](context.md) for the wait that gives up.

### Test.parallel twice or off the body

`Test.parallel` may be called once, from the test's body, and panics otherwise, as from a second
call or from a goroutine the test starts:

<!-- test: cmd ant test twice_test.ant -->
<!-- test: exit 1 -->

```ant [twice_test.ant]
test "twice" {
  Test.parallel()
  Test.parallel()
}
```

```text
--- FAIL: twice_test.ant: twice (0.00s)
    Test.parallel: called twice
      at test "twice" (twice_test.ant:3:16)
FAIL  1 failed, 0 passed (0.00s)
```

See [Test.parallel](parallel.md) for the call from a goroutine.

### Test.skip while a panic unwinds

`Test.skip` in a deferred block that runs while a panic unwinds is dropped and the panic goes on, so
the test stays failed. See [Test.skip](skip.md).

## See also

- [test](../../language/keywords/test.md): declare tests and subtests
- [assert](../prelude/assert.md), [Panic.catch](../Panic/catch.md),
  [defer](../../language/keywords/defer.md)
- [ant test](../../tools/test.md): run tests, filter them, and read the report
- [Testing](../../guide/testing.md): how tests, subtests, and helpers fit together
