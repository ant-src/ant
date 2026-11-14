<!-- status: reviewed -->

# ant test

`ant test` finds test files, runs their tests, and reports the ones that failed. A test file's name
ends in `_test.ant`; what goes in one is on the [test](../language/keywords/test.md) page.

```sh
ant test [flags] [path ...]
```

## Details

- A path is a test file or a directory. A directory is searched recursively for test files, leaving
  out directories whose names begin with `.`. With no path, `ant test` searches `.`.
- Flags come before the paths.
- Files run one after another, in the order of the paths, and a directory's files in lexical order.
  Each file runs as a program of its own, so module state never carries from one file to the next.
- `ant test` does not type-check. Run `ant check` on a test file to check it with its imports.
- A passing test prints nothing. What a test prints is shown when it fails, or under `-v`.
- Output printed while a file's modules initialize, outside any test, is not collected: it goes to
  standard output, or to standard error under `-json`.

## Basic examples

A library, `hex.ant`, reads decimal and hexadecimal numbers:

```ant [hex.ant]
export func parse(text: string): int | null {
  if text.startsWith("0x") { text.trimPrefix("0x").toInt(16) } else { text.toInt() }
}
```

Its tests, in `hex_test.ant`. `ant test hex_test.ant` runs both tests, and when every test passes,
it prints one line:

<!-- test: cmd ant test hex_test.ant -->

```ant [hex_test.ant]
import "./hex.ant" { parse }

test "parse" {
  assert(parse("42") == 42)
  assert(parse("0xff") == 255)
}

test "rejects words" {
  assert(parse("ff") == null)
}
```

```text
ok    2 passed (0.00s)
```

`ant test` with no path runs every test file under the current directory. In a project that holds
only `hex.ant` and `hex_test.ant`, it prints the same line:

```sh
ant test
```

```text
ok    2 passed (0.00s)
```

A failing test is reported with its file, its full name, and the failure. This table has a row the
library gets wrong, so `ant test broken_test.ant` prints the failure, then the counts, and exits
with status 1:

<!-- test: cmd ant test broken_test.ant -->
<!-- test: exit 1 -->

```ant [broken_test.ant]
import "./hex.ant" { parse }

test "parse" {
  for input in ["0x1f", "0X1F"] {
    test "${input}" {
      assert(parse(input) == 31)
    }
  }
}
```

```text
--- FAIL: broken_test.ant: parse/0X1F (0.00s)
    assertion failed: parse(input) == 31 (null == 31)
      input = "0X1F"
      at test "0X1F" (broken_test.ant:6:13)
FAIL  1 failed, 1 passed (0.00s)
```

The counts are of tests without subtests: the two rows, not the test `parse` that holds them. A
parent that fails only because a subtest did is not reported again.

## Flags

| Flag           | Meaning                                                                                    |
| -------------- | ------------------------------------------------------------------------------------------ |
| `-run regexp`  | Run only the tests whose full name matches, one part of the pattern per level              |
| `-skip regexp` | Leave out the tests that match, the same way                                               |
| `-v`           | Print every test as it starts and ends, with its output                                    |
| `-json`        | Print events as JSON lines instead of the report                                           |
| `-failfast`    | Start no test after the first failure                                                      |
| `-parallel n`  | The most tests that run at once after `Test.parallel()`; the default is the number of CPUs |
| `-timeout d`   | End the run after `d`, such as `30s` or `2m`; the default is `10m`, and `0` means no limit |

## Scope

### Directories

A directory runs every test file under it. With `-v`, each test prints a line when it starts and
when it ends. The file is named by the path it was found by. `ant test -v strings` runs
`trim_test.ant` before `upper_test.ant`, in lexical order, and shows what the test printed,
indented, as it comes:

```ant [strings/upper_test.ant]
test "upper" {
  assert("ant".toUpper() == "ANT")
}
```

<!-- test: cmd ant test -v strings -->

```ant [strings/trim_test.ant]
test "trim" {
  println("trimming")
  assert("  ant ".trim() == "ant")
}
```

```text
=== RUN   strings/trim_test.ant: trim
    trimming
--- PASS: strings/trim_test.ant: trim (0.00s)
=== RUN   strings/upper_test.ant: upper
--- PASS: strings/upper_test.ant: upper (0.00s)
ok    2 passed (0.00s)
```

### Choosing tests with `-run`

The pattern is split on `/` into parts, one per level of subtests. Each part is a regular expression
matched anywhere in the name at its level. `ant test -v -run parse/0x run_test.ant` runs the rows of
`parse` whose names contain `0x`. The parent `parse` runs too, since its body is what starts the
rows:

<!-- test: cmd ant test -v -run parse/0x run_test.ant -->

```ant [run_test.ant]
import "./hex.ant" { parse }

test "parse" {
  for input in ["7", "0x10", "0xff"] {
    test "${input}" {
      assert(parse(input) != null)
    }
  }
}

test "rejects words" {
  assert(parse("seven") == null)
}
```

```text
=== RUN   run_test.ant: parse
=== RUN   run_test.ant: parse/0x10
--- PASS: run_test.ant: parse/0x10 (0.00s)
=== RUN   run_test.ant: parse/0xff
--- PASS: run_test.ant: parse/0xff (0.00s)
--- PASS: run_test.ant: parse (0.00s)
ok    2 passed (0.00s)
```

`-run parse` alone runs `parse` with all its rows. A test that is left out is neither run nor
counted.

### Leaving tests out with `-skip`

`-skip` leaves out the tests that match, with their subtests. `ant test -skip network skip_test.ant`
runs only `parse`:

<!-- test: cmd ant test -skip network skip_test.ant -->

```ant [skip_test.ant]
test "parse" {
  assert("42".toInt() == 42)
}

test "network" {
  test "fetch" {
    let addr = Os.env("SERVICE_ADDR") ?? "localhost:8080"
    assert(addr.contains(":"))
  }
}
```

```text
ok    1 passed (0.00s)
```

### Stopping at the first failure

`-failfast` starts no test after one fails. `ant test -failfast failfast_test.ant` reports `first`
and never starts `second`:

<!-- test: cmd ant test -failfast failfast_test.ant -->
<!-- test: exit 1 -->

```ant [failfast_test.ant]
test "first" {
  assert(1 + 1 == 3)
}

test "second" {
  assert(2 + 2 == 5)
}
```

```text
--- FAIL: failfast_test.ant: first (0.00s)
    assertion failed: 1 + 1 == 3 (2 == 3)
      at test "first" (failfast_test.ant:2:9)
FAIL  1 failed, 0 passed (0.00s)
```

### Skipped tests

A test that calls `Test.skip` stops there and counts as skipped. Under `-v`, the report gives the
reason on the line after `--- SKIP:`, and the counts end with the skipped tests:

<!-- test: cmd ant test -v db_test.ant -->

```ant [db_test.ant]
test "db" {
  Test.skip("no database")
}

test "parse" {
  assert("42".toInt() == 42)
}
```

```text
=== RUN   db_test.ant: db
--- SKIP: db_test.ant: db (0.00s)
    no database
=== RUN   db_test.ant: parse
--- PASS: db_test.ant: parse (0.00s)
ok    1 passed, 1 skipped (0.00s)
```

### Time limit

`-timeout` ends a run that takes too long:

<!-- test: cmd ant test -timeout 100ms slow_test.ant -->
<!-- test: exit 1 -->

```ant [slow_test.ant]
test "waits" {
  Time.sleep(1m)
}
```

The command exits with status 1 and writes the tests still running to standard error:

```sh
ant test -timeout 100ms slow_test.ant
```

```text
test timed out after 100ms
running tests:
    slow_test.ant: waits
```

### JSON events

Under `-json`, `ant test` writes one JSON object per line to standard output and nothing else. Each
has an `action`: `"run"`, `"output"`, `"pass"`, `"skip"`, `"fail"`, and last `"done"`:

```sh
ant test -json broken_test.ant
```

```text
{"action":"run","file":"broken_test.ant","test":"parse"}
{"action":"run","file":"broken_test.ant","test":"parse/0x1f"}
{"action":"pass","file":"broken_test.ant","test":"parse/0x1f","elapsed":0.000021125}
{"action":"run","file":"broken_test.ant","test":"parse/0X1F"}
{"action":"fail","file":"broken_test.ant","test":"parse/0X1F","elapsed":0.00006125,"message":"assertion failed: parse(input) == 31 (null == 31)\n  input = \"0X1F\"\n  at test \"0X1F\" (broken_test.ant:6:13)"}
{"action":"fail","file":"broken_test.ant","test":"parse","elapsed":0.000208667}
{"action":"done","elapsed":0.000576584,"passed":1,"failed":1,"skipped":0}
```

`elapsed` is in seconds. A `"fail"` event's `message` holds the test's failures as the report shows
them.

## Applications

### Summarize failures for a CI job

Run `ant test -json` from a script and print one line per failure:

```ant
let run = try Cmd.capture(Cmd.new("ant", ["test", "-json", "hex_test.ant", "broken_test.ant"]))
for line in run.stdout.lines() {
  let event = try Json.parse(line)
  if !(event is Map) { continue }
  let message = event["message"]
  if event["action"] == "fail" && message is string {
    println("${event["file"]}: ${event["test"]}: ${message.lines()[0]}")
  }
}
println("exit status ${run.status}")
```

```text
broken_test.ant: parse/0X1F: assertion failed: parse(input) == 31 (null == 31)
exit status 1
```

## Exit status

| Status | When                                                                                              |
| ------ | ------------------------------------------------------------------------------------------------- |
| `0`    | No test failed, including when tests were skipped and when there were none                        |
| `1`    | A test failed, or the time ran out                                                                |
| `2`    | A file could not be loaded, such as for a syntax error or a missing module, or the usage is wrong |

A file that cannot be loaded is reported on standard error, and the other files still run.
`ant test syntax_test.ant sum_test.ant` reports the syntax error, runs `sum_test.ant`, and exits
with status 2:

```ant [syntax_test.ant]
test "sum" {
  assert(1 + == 2) // error: unexpected '=='
}
```

<!-- test: cmd ant test syntax_test.ant sum_test.ant -->
<!-- test: exit 2 -->

```ant [sum_test.ant]
test "sum" {
  assert(1 + 1 == 2)
}
```

```text
FAIL  1 passed (0.00s)
```

`Os.exit` in a test ends the whole command with its status, as it ends a program. The tests after it
do not run:

<!-- test: cmd ant test exit_test.ant -->
<!-- test: exit 3 -->

```ant [exit_test.ant]
test "exits" {
  Os.exit(3)
}

test "never runs" {
  assert(false)
}
```

`ant test exit_test.ant` prints nothing and exits with status 3.

## Possible issues

### A panic while a file loads

A panic while a test file's modules are initialized, such as in a top-level `let`, stops that file:
none of its tests run, and the file counts as one failed test. The panic goes to standard error, as
under `ant run`, and standard output has only the counts:

<!-- test: cmd ant test init_test.ant -->
<!-- test: exit 1 -->

```ant [init_test.ant]
let ports = [80, 443]
let third = ports[2]

test "third port" {
  assert(third == 8080)
}
```

```text
FAIL  1 failed, 0 passed (0.00s)
```

### A name that holds `/`

`-run` splits its pattern on every `/`, so a test named `v1/users` is matched by `v1` alone, and its
subtests by `users`, which none of them match. `ant test -v -run v1/users routes_test.ant` runs the
parent's body and none of its subtests:

<!-- test: cmd ant test -v -run v1/users routes_test.ant -->

```ant [routes_test.ant]
test "v1/users" {
  test "empty" {
    assert("".isEmpty())
  }
  test "one" {
    assert("a".length() == 1)
  }
}
```

```text
=== RUN   routes_test.ant: v1/users
--- PASS: routes_test.ant: v1/users (0.00s)
ok    1 passed (0.00s)
```

Escape the slash, `-run 'v1\/users'`, or put it in brackets, `-run 'v1[/]users'`, to match it as
part of a name.

### A pattern that matches nothing

A `-run` pattern that matches no test is not an error: `ant test` prints `ok    no tests (0.00s)`
and exits with status 0. Patterns are case-sensitive, so `-run Parse` matches no test named `parse`.

### A file that is not a test file

A path named directly must be a test file, whose name ends in `_test.ant`. Naming the library
instead of its tests, as in `ant test words.ant`, is a usage error: nothing runs, and the command
exits with status 2.

<!-- test: cmd ant test words.ant -->
<!-- test: exit 2 -->

```ant [words.ant]
export func count(text: string): int {
  text.fields().length()
}
```

`ant test` writes the reason to standard error:

```text
ant test: words.ant is not a test file, whose name ends in _test.ant
```

## See also

- [test](../language/keywords/test.md): writing tests and subtests
- [Testing](../guide/testing.md): how test files, subtests, and `Test` fit together
- [assert](../lib/prelude/assert.md): the check a test makes
- [Test](../lib/Test/index.md): skip, run in parallel, temporary directories
- [ant check](check.md): type-check a test file with its imports
