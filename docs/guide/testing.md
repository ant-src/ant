<!-- status: reviewed -->

# Testing

Tests live next to the code they test, in test files whose names end in `_test.ant`. A `test`
declaration names a test and holds its body; `assert` checks what the body expects; a `test`
statement inside a body runs a subtest. The `Test` namespace holds the rest: temporary directories,
parallel runs, skipping, a context, and the test's name. `ant test` finds the test files and runs
them.

## Test files

A library, `stack.ant`, with a private function `Stack.peek`:

```ant [stack.ant]
export type Stack = { items: Array[int] }

export func Stack.new(): Stack { Stack{ items: [] } }

export func Stack.push(s: Stack, value: int): void { s.items.push(value) }

export func Stack.pop(s: Stack): int {
  s.items.pop() ?? panic("pop on an empty stack")
}

func Stack.peek(s: Stack): int | null { s.items.last() }
```

Its tests, in `stack_test.ant`:

<!-- test: cmd ant test stack_test.ant -->

```ant [stack_test.ant]
import "./stack.ant" { Stack }

test "pop returns the last value pushed" {
  let s = Stack.new()
  s.push(1)
  s.push(2)
  assert(s.pop() == 2)
  assert(s.peek() == 1)
}

test "pop on an empty stack panics" {
  let r = Panic.catch({ Stack.new().pop() })
  assert(r is Err)
  assert(r.error.message == "pop on an empty stack")
}
```

```text
ok    2 passed (0.00s)
```

A test file is a library that may also hold tests. It holds imports, `func`, `type`, `var`, and
`test` declarations, and no top-level statement. `ant run` refuses it; `ant test` runs it. With no
path, `ant test` runs every test file under the current directory:

```sh
ant test                    # every test file under .
ant test stack_test.ant     # one file
ant test -run 'pop on'      # only the tests whose names match
ant test -list 'pop'        # list the matching top-level tests, run none
```

A `-run` or `-skip` pattern that leaves no test to run is not an error: `ant test` warns
`ant test: warning: no tests to run` and exits with status 0. To see which top-level tests a pattern
names before running them, use `-list`.

Each test file runs as a program of its own, so module state never leaks from one file into another.
Top-level tests of one file run one at a time, in source order, after the file's modules are
initialized, unless they call `Test.parallel()` (Parallel tests, below).

## Private imports

The test above calls `s.peek()`, though `stack.ant` does not export `Stack.peek`. A test file's
imports reach every top-level declaration of a module, exported or not, and the private functions
declared on an imported type come with it. So you test a helper directly, without exporting it for
the tests' sake. Nothing private leaks into a program, since only a test file can import a test
file.

A script cannot be imported, by a test file or any other. Code in a script's top-level statements
cannot be tested where it is:

```ant [greet.ant]
let name = Os.args().get(0) ?? "world"
println("Hello, ${name}")
```

```ant [greet_test.ant]
import "./greet.ant" { name }
// error: cannot import "./greet.ant": it is a script, made one by the statement at greet.ant:2:1, and a script cannot be imported
```

Move the logic worth testing into a library, have the script import it, and test the library. The
script keeps only the reading of arguments and the printing.

## Tests, assertions, and failures

A test passes unless something fails it:

- **A panic leaves its body.** A failed `assert` is a panic, and so is any other.
- **Its body returns an `Err`.** `try` returns one, so a test can call fallible code without
  unwrapping it.
- **A subtest fails**, or a goroutine the test owns panics or outlives it.

A panic or a returned `Err` ends the body at once, after its deferred blocks run. A failed subtest
or a goroutine's panic does not stop the body: the test goes on and fails when its body ends. There
is no failure that records and goes on. The report shows the condition, the values it compared, and
the names it read:

<!-- test: cmd ant test total_test.ant -->
<!-- test: exit 1 -->

```ant [total_test.ant]
func parsePrice(text: string): Result[int, Error] {
  let cents = text.toInt() ?? return Result.fail("not a price: ${text}")
  Result.ok(cents)
}

test "total" {
  let prices = ["120", "80"]
  var sum = 0
  for p in prices {
    sum += try parsePrice(p)
  }
  assert(sum == 210, "the prices add up")
}

test "bad price" {
  let cents = try parsePrice("1.20")
  assert(cents == 120)
}
```

```text
--- FAIL: total_test.ant: total (0.00s)
    assertion failed: sum == 210 (200 == 210): the prices add up
      at test "total" (total_test.ant:12:9)
--- FAIL: total_test.ant: bad price (0.00s)
    not a price: 1.20
FAIL  2 failed, 0 passed (0.00s)
```

The first test expects the wrong sum, and `assert` shows both sides. The second gets an `Err` from
`try`, which returns it from the body, and the report shows the error.

Pick the check by what you expect:

| You expect                              | Write                                                      |
| --------------------------------------- | ---------------------------------------------------------- |
| a condition holds                       | `assert(got == want)`                                      |
| a call succeeds, and you need its value | `let v = try call()`                                       |
| a call fails with an error              | `let r = call()`, then `assert(r is Err)`                  |
| a call panics                           | `let r = Panic.catch({ call() })`, then `assert(r is Err)` |

A passing test prints nothing. What a test prints is collected and shown with its failure, or as it
comes under `ant test -v`.

## Subtests and table tests

A `test` statement inside a test's body runs a subtest. In a loop over a table it runs one subtest
per row, named after the row:

```ant [words.ant]
export func countWords(text: string): int {
  text.fields().length()
}
```

<!-- test: cmd ant test words_test.ant -->
<!-- test: exit 1 -->

```ant [words_test.ant]
import "./words.ant" { countWords }

type Case = { name: string, text: string, want: int }

test "countWords" {
  let cases = [
    Case{ name: "empty", text: "", want: 0 },
    Case{ name: "one", text: "one", want: 1 },
    Case{ name: "spaces", text: "  two  words ", want: 2 },
    Case{ name: "tabs", text: "a\tb\nc", want: 4 },
  ]
  for c in cases {
    test "${c.name}" {
      assert(countWords(c.text) == c.want)
    }
  }
}
```

```text
--- FAIL: words_test.ant: countWords/tabs (0.00s)
    assertion failed: countWords(c.text) == c.want (3 == 4)
      c.text = "a\tb\nc"
      at test "tabs" (words_test.ant:14:13)
FAIL  1 failed, 3 passed (0.00s)
```

The `tabs` row expects the wrong count. Its failure ends that subtest only: the parent goes on to
the next row and fails when its body ends. The run counts the four rows, not the test that holds
them.

- A subtest's full name joins the names with `/`: `countWords/tabs`.
  `ant test -run 'countWords/tabs'` runs that row alone.
- A subtest's name may interpolate; a top-level test's name is a plain string, so tools list the
  tests without running anything, as `ant test -list` does.
- A subtest sees the variables around it, as a lambda does, and subtests nest to any depth.
- Two subtests of one test with the same name are told apart by ` #2`, ` #3`, and so on; an empty
  subtest name panics.
- A subtest runs only in a test's body, not inside a function, lambda, `go` block, or `defer` block.

Use a subtest for each case of one behavior. Use separate top-level tests for separate behaviors.

## The `Test` namespace

There is no test handle to pass around. Each `Test` function acts on the current test: the test
whose body the calling goroutine runs, or the test that owns the goroutine. A helper function can
call one without being given anything. Called outside any test, a `Test` function panics.
`Test.name()` returns the current test's full name, such as `countWords/tabs`, for a helper that
labels files or log lines.

Share helpers in a test file of their own. Only test files can import it:

```ant [helpers_test.ant]
export func writeTemp(name: string, text: string): Result[string, Error] {
  let path = Path.join([Test.tempDir(), name])
  try File.write(path, text)
  Result.ok(path)
}
```

### Temporary directories

`Test.tempDir()` returns a new directory at every call. `ant test` removes it with its contents when
the test ends, so a test needs no cleanup and never sees another test's files:

```ant [lines.ant]
export func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.lines().length())
}
```

<!-- test: cmd ant test lines_test.ant -->

```ant [lines_test.ant]
import "./lines.ant" { countLines }
import "./helpers_test.ant" { writeTemp }

test "countLines counts each line" {
  let path = try writeTemp("notes.txt", "a\nb\nc\n")
  let lines = try countLines(path)
  assert(lines == 3)
}

test "countLines reports a missing file" {
  let path = Path.join([Test.tempDir(), "missing.txt"])
  assert(countLines(path) is Err)
}
```

```text
ok    2 passed (0.00s)
```

### Skipping

`Test.skip(reason)` ends the current test as skipped. Call it in a guard, before the main path, when
the environment lacks what the test needs. A skip is not a panic, so `Panic.catch` does not stop it:

<!-- test: cmd ant test db_test.ant -->

```ant [db_test.ant]
test "connects to the database" {
  let url = Os.env("TEST_DATABASE_URL") ?? Test.skip("TEST_DATABASE_URL is not set")
  assert(url.startsWith("postgres://"))
}
```

```text
ok    0 passed, 1 skipped (0.00s)
```

## Parallel tests

Tests run one at a time unless they call `Test.parallel()`. A test that calls it waits until every
test of the same parent that did not call it has ended, then runs alongside the parent's other
parallel tests, at most `ant test -parallel n` at a time; `n` defaults to
[`Os.cpus()`](../lib/Os/cpus.md), the CPUs the program may use.

Call `Test.parallel()` once, from the test's body; a second call, or a call from a deferred block or
a goroutine, panics. Rows of a table are the usual candidates:

<!-- test: cmd ant test parallel_test.ant -->

```ant [parallel_test.ant]
import "./lines.ant" { countLines }
import "./helpers_test.ant" { writeTemp }

type Case = { name: string, text: string, want: int }

test "countLines" {
  let cases = [
    Case{ name: "empty", text: "", want: 0 },
    Case{ name: "one", text: "a", want: 1 },
    Case{ name: "three", text: "a\nb\nc\n", want: 3 },
  ]
  for c in cases {
    test "${c.name}" {
      Test.parallel()
      let path = try writeTemp("input.txt", c.text)
      let lines = try countLines(path)
      assert(lines == c.want)
    }
  }
}
```

```text
ok    3 passed (0.00s)
```

Each row writes `input.txt` into a directory of its own, so the rows cannot disturb each other. Run
in parallel only tests that share nothing they change. Tests in different files never run at once.

A parallel subtest resumes from its `Test.parallel()` call only after its parent's body and deferred
blocks have finished. A parent that sets something up and tears it down with `defer` tears it down
before its parallel subtests run:

<!-- test: cmd ant test order_test.ant -->
<!-- test: exit 1 -->

```ant [order_test.ant]
type Server = { open: bool }

test "server" {
  let server = Server{ open: true }
  defer { server.open = false }
  test "answers" {
    Test.parallel()
    assert(server.open)
  }
}
```

```text
--- FAIL: order_test.ant: server/answers (0.00s)
    assertion failed: server.open
      at test "answers" (order_test.ant:8:11)
FAIL  1 failed, 0 passed (0.00s)
```

Keep such subtests sequential, or give each one its own setup.

## Goroutines and contexts

A test owns the goroutines it starts, and every goroutine they start in turn. A panic in one fails
the test. One still running shortly after the test ends fails it too, with
`goroutine started at … outlived the test`, and `ant test` ends it, so a leak cannot disturb the
next test.

Give a goroutine's blocking calls `Test.context()`. It is done when the test's body ends, before its
deferred blocks run, so the goroutine stops with the test:

<!-- test: cmd ant test worker_test.ant -->

```ant [worker_test.ant]
test "worker keeps a running total" {
  let jobs: Channel[int] = Channel.new()
  let totals: Channel[int] = Channel.new()
  let ctx = Test.context()
  go {
    var total = 0
    while true {
      let job = jobs.receive(ctx: ctx) ?? break
      total += job
      totals.send(total)
    }
  }
  jobs.send(2)
  assert(totals.receive() == 2)
  jobs.send(3)
  assert(totals.receive() == 5)
}
```

```text
ok    1 passed (0.00s)
```

## Coming from other languages

- **No test handle.** A Go test takes a `*testing.T`; an Ant test takes nothing. `Test.tempDir()`,
  `Test.skip`, and the other `Test` functions find the current test themselves.
- **A failure stops the test.** There is no `t.Errorf` that records and goes on: a failed `assert`
  ends the body. To check many cases and see every failure, make each case a subtest.
- **No setup and teardown hooks.** Set up in the body, or in a helper it calls, and tear down with
  `defer` right after. A temporary directory needs no teardown at all.
- **Module state is per file.** Each test file runs as its own program. Within one file, all its
  tests share a top-level `var`, so a change one test makes is seen by the tests after it.
- **Private names are reachable.** Import them into the test file; there is no need to export a
  function for its tests.
- **`ant test` type-checks first.** It checks the test files with their imports and runs nothing
  when the checker reports an error; `-nocheck` skips the check. A script that no test imports is
  not checked, so run `ant check .` too.

## See also

- [Tests](../tour/20-tests.md): the tour step on testing
- [`test`](../language/keywords/test.md): the declaration and subtests in full
- [`ant test`](../tools/test.md): the command, its flags, and its report
- [Testing](../lib/guides/testing.md): `assert`, `Panic.catch`, and the `Test` functions by job
- [assert](../lib/prelude/assert.md) and [Panic.catch](../lib/Panic/catch.md)
- [Test.tempDir](../lib/Test/tempDir.md), [Test.parallel](../lib/Test/parallel.md),
  [Test.skip](../lib/Test/skip.md), [Test.context](../lib/Test/context.md),
  [Test.name](../lib/Test/name.md)
- [Modules](modules.md): libraries, scripts, and imports
- [Errors](errors.md): when to use `assert`, `Result`, and `panic`
