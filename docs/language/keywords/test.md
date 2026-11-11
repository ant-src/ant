<!-- status: reviewed -->

# `test`

`test` declares a test: a name and a block that `ant test` runs. Tests live in test files, whose
names end in `_test.ant`. Inside a test, a `test` statement runs a subtest.

## Details

- `test` is a keyword only in test files. Anywhere else it is an ordinary name. In a test file it is
  still a name after `.` or `?.` and before `:`, so `Regex.test(re, s)`, a field named `test`, and a
  named argument `test: x` still work.
- A top-level test's name is a string literal without interpolation, not empty, and unique in its
  file. A subtest's name may interpolate; an empty one panics, and a repeated one gets ` #2`, ` #3`,
  and so on appended.
- A test fails when a panic leaves its body, when its body returns an `Err`, when one of its
  subtests fails, or when a goroutine it started panics or outlives it. A failed subtest does not
  stop its parent's body.
- A test body is like a script body: `return` ends the test, `try` returns an `Err` from it, and
  `defer` runs when the test ends. Its last expression is discarded.
- Top-level tests run one at a time, in source order, after the modules are initialized. Subtests
  run in the order their statements run. A test that calls
  [Test.parallel](../../lib/Test/parallel.md) runs alongside the others that call it.
- A test passes silently. What it prints is shown only when it fails, or under `ant test -v`.

## Basic examples

A library, `ports.ant`, with an exported function and a private helper:

```ant [ports.ant]
export func parsePort(text: string): Result[int, Error] {
  let port = text.toInt() ?? return Result.fail("not a number: ${text}")
  if !inRange(port) { return Result.fail("out of range: ${port}") }
  Result.ok(port)
}

func inRange(port: int): bool {
  port >= 1 && port <= 65535
}
```

Its tests, in `ports_test.ant`. A test file may import the private names of the module it tests:

<!-- test: cmd ant test ports_test.ant -->

```ant [ports_test.ant]
import "./ports.ant" { parsePort, inRange }

test "parsePort reads a number" {
  assert(parsePort("8080").unwrap() == 8080)
}

test "parsePort rejects a word" {
  let r = parsePort("http")
  assert(r is Err)
  assert(r.error.message == "not a number: http")
}

test "inRange" {
  assert(inRange(1))
  assert(!inRange(0))
  assert(!inRange(70000))
}
```

`ant test` runs every test and prints one line when they all pass:

```text
ok    3 passed (0.00s)
```

## Scope

### Subtests

A `test` statement inside a test's body runs a subtest. In a loop it runs one subtest per row of a
table, and its name may interpolate the row:

<!-- test: cmd ant test table_test.ant -->

```ant [table_test.ant]
import "./ports.ant" { parsePort }

type Case = { text: string, want: int | null }

test "parsePort" {
  let cases = [
    Case{ text: "80", want: 80 },
    Case{ text: "0", want: null },
    Case{ text: "65536", want: null },
    Case{ text: "port", want: null },
  ]
  for c in cases {
    test "${c.text}" {
      let got = parsePort(c.text)
      if c.want == null {
        assert(got is Err)
      } else {
        assert(got.unwrap() == c.want)
      }
    }
  }
}
```

```text
ok    4 passed (0.00s)
```

The run counts the four subtests, not the test that holds them. A subtest's full name joins the
names with `/`, such as `parsePort/65536`, and `ant test -run 'parsePort/655'` runs only that row.

A subtest sees the variables of the blocks around it, as a lambda does. Subtests nest to any depth.

### Repeated subtest names

When two subtests of one test get the same name, the second gets ` #2`, the third ` #3`.
[Test.name](../../lib/Test/name.md) gives the full name:

<!-- test: cmd ant test names_test.ant -->

```ant [names_test.ant]
test "names" {
  let names: Array[string] = []
  for word in ["a", "b", "a"] {
    test "${word}" {
      names.push(Test.name())
    }
  }
  assert(names == ["names/a", "names/b", "names/a #2"])
}
```

```text
ok    3 passed (0.00s)
```

### Failures

A failed `assert` ends its test and reports the condition and the values it used. A failed row does
not stop the table: the other rows still run, and `ant test` exits with status 1:

<!-- test: cmd ant test fail_test.ant -->
<!-- test: exit 1 -->

```ant [fail_test.ant]
test "even" {
  for n in [2, 3, 4] {
    test "${n}" {
      assert(n % 2 == 0)
    }
  }
}
```

```text
--- FAIL: fail_test.ant: even/3 (0.00s)
    assertion failed: n % 2 == 0 (1 == 0)
      n = 3
      at test "3" (fail_test.ant:4:13)
FAIL  1 failed, 2 passed (0.00s)
```

### `try` and `Err`

A test body may return an `Err`, which fails the test. `try` returns one early:

<!-- test: cmd ant test try_test.ant -->
<!-- test: exit 1 -->

```ant [try_test.ant]
import "./ports.ant" { parsePort }

test "default port" {
  let port = try parsePort("eighty")
  assert(port == 80)
}
```

```text
--- FAIL: try_test.ant: default port (0.00s)
    not a number: eighty
FAIL  1 failed, 0 passed (0.00s)
```

### Panics

To test that code panics, catch the panic with [Panic.catch](../../lib/Panic/catch.md) and check the
`Err`:

<!-- test: cmd ant test panic_test.ant -->

```ant [panic_test.ant]
func first(xs: Array[int]): int {
  if xs.isEmpty() { panic("first of an empty array") }
  xs[0]
}

test "first of an empty array panics" {
  let r = Panic.catch({ first([]) })
  assert(r is Err)
  assert(r.error.message.contains("first of an empty array"))
}
```

```text
ok    1 passed (0.00s)
```

### `defer` and output

`defer` in a test body runs when the test ends, passed or failed. What the test prints is collected
and shown with its failure:

<!-- test: cmd ant test defer_test.ant -->
<!-- test: exit 1 -->

```ant [defer_test.ant]
test "cleanup runs" {
  defer { println("cleaning up") }
  println("starting")
  assert(1 + 1 == 3)
}
```

```text
--- FAIL: defer_test.ant: cleanup runs (0.00s)
    starting
    cleaning up
    assertion failed: 1 + 1 == 3 (2 == 3)
      at test "cleanup runs" (defer_test.ant:4:9)
FAIL  1 failed, 0 passed (0.00s)
```

### Goroutines

A test owns the goroutines it starts. One still running shortly after the test ends fails it, and
the runner ends it. A panic in one fails the test when the test ends, without interrupting its body.
Here the goroutine waits on a channel that nothing sends to; the test's time includes the short wait
for it:

<!-- test: cmd ant test leak_test.ant -->
<!-- test: exit 1 -->

```ant [leak_test.ant]
test "leak" {
  let c: Channel[int] = Channel.new()
  go { c.receive() }
}
```

```text
--- FAIL: leak_test.ant: leak (0.10s)
    goroutine started at leak_test.ant:3:3 outlived the test
FAIL  1 failed, 0 passed (0.10s)
```

A test should wait for its goroutines, or close what they block on, before it ends.

### `test` outside test files

In any other file, `test` is an ordinary name:

```ant
let test = "a name like any other"
println(test)
```

```text
a name like any other
```

## Applications

Share helpers between test files in a test file of their own. Only a test file can import it:

```ant [helpers_test.ant]
export func assertErr(r: Result[int, Error], message: string): void {
  assert(r is Err)
  assert(r.error.message == message)
}
```

<!-- test: cmd ant test errors_test.ant -->

```ant [errors_test.ant]
import "./ports.ant" { parsePort }
import "./helpers_test.ant" { assertErr }

test "errors" {
  assertErr(parsePort("x"), "not a number: x")
  assertErr(parsePort("0"), "out of range: 0")
}
```

```text
ok    1 passed (0.00s)
```

Skip a test that cannot run here with [Test.skip](../../lib/Test/skip.md):

<!-- test: cmd ant test skip_test.ant -->

```ant [skip_test.ant]
test "needs a database" {
  let url = Os.env("TEST_DATABASE_URL") ?? Test.skip("TEST_DATABASE_URL is not set")
  assert(url.startsWith("postgres://"))
}
```

```text
ok    1 skipped (0.00s)
```

## Possible issues

A test file holds only imports, declarations, and tests. A top-level statement is an error:

```ant [setup_test.ant]
println("setting up")
// error: a test file cannot have top-level statements: it holds imports, declarations, and tests
```

A top-level test's name cannot interpolate, so tools can list the tests without running anything.
Use a subtest for a computed name:

```ant [name_test.ant]
let size = 3
test "size ${size}" {} // error: a top-level test's name is a string without interpolation
```

Each top-level name appears once in a file:

```ant [dup_test.ant]
test "parse" {}
test "parse" {} // error: test "parse" is already declared in this file
```

A subtest runs only in a test's body. A function, a lambda, a `go` block, or a `defer` block cannot
hold one:

```ant [helper_test.ant]
func check(n: int) {
  test "${n}" {} // error: a test is declared at the top level of a test file, or runs as a subtest
}
```

A test body returns `void` or an `Err`:

```ant [value_test.ant]
test "answer" {
  return 42 // error: a test returns void or an Err, not int
}
```

A test is not a value, and a `test` statement has no result to keep:

```ant [value2_test.ant]
test "outer" {
  let t = test "inner" {} // error: unexpected 'test'
}
```

In a test file, `test` cannot name a variable:

```ant [keyword_test.ant]
test "t" {
  let test = 1 // error: 'test' cannot be a variable name
}
```

`ant run` refuses a test file; run it with `ant test`. In a file whose name does not end in
`_test.ant`, a `test` declaration does not parse:

```ant
test "math" { // error: expected a newline or ';' before string literal
  assert(1 + 1 == 2)
}
```

## See also

- [Tests](../../tour/20-tests.md): the tour step on testing
- [Testing](../../guide/testing.md): how test files, subtests, and `Test` fit together
- [ant test](../../tools/test.md): the command, its flags, and its report
- [assert](../../lib/prelude/assert.md): the check a test makes
- [Test](../../lib/Test/index.md): skip, run in parallel, temporary directories
- [Panic.catch](../../lib/Panic/catch.md): test that code panics
- [import](import.md): importing private names into a test file
