<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.parallel

```ant signature
func Test.parallel(): void
```

Test.parallel runs the current test alongside the other parallel tests of its parent, once the tests
of that parent that did not call it have ended. It is called once, from the test's body.

<!-- end generated -->

## Details

- The test pauses at the call until its serial siblings have ended: the tests of the same parent
  that do not call `Test.parallel`, including those declared after it. For a top-level test, those
  are the serial top-level tests of the file.
- A subtest that calls `Test.parallel` returns control to its parent at the call, so the parent's
  body goes on. The parent ends only after its parallel subtests end, and its deferred blocks run
  before they end.
- At most `-parallel n` tests run at once after the call; the default is the number of CPUs. Tests
  in different files never run at once.
- Parallel tests share module variables and the variables of their parent. Guard what they change
  with a [Mutex](../Mutex/index.md).
- Calling it twice in one test, from a goroutine the test started, or outside any test panics.

## Basic examples

Call `Test.parallel()` first in each test that can run alongside the others:

<!-- test: cmd ant test strings_test.ant -->

```ant [strings_test.ant]
test "toUpper" {
  Test.parallel()
  assert("ant".toUpper() == "ANT")
}

test "trim" {
  Test.parallel()
  assert("  ant ".trim() == "ant")
}
```

```text
ok    2 passed (0.00s)
```

A serial test runs before the parallel ones, even when it comes later in the file. `ant test -v`
shows the order:

<!-- test: cmd ant test -v order_test.ant -->

```ant [order_test.ant]
test "first in the file" {
  Test.parallel()
  println("runs last")
}

test "second in the file" {
  println("runs first")
}
```

```text
=== RUN   order_test.ant: first in the file
=== RUN   order_test.ant: second in the file
    runs first
--- PASS: order_test.ant: second in the file (0.00s)
    runs last
--- PASS: order_test.ant: first in the file (0.00s)
ok    2 passed (0.00s)
```

## Scope

### Subtests

In a table of subtests, each row that calls `Test.parallel` runs alongside the other rows. Each
iteration of the loop has its own `c`, so every row checks its own case:

<!-- test: cmd ant test table_test.ant -->

```ant [table_test.ant]
type Case = { text: string, want: int }

test "toInt" {
  let cases = [
    Case{ text: "7", want: 7 },
    Case{ text: "-12", want: -12 },
    Case{ text: "+3", want: 3 },
  ]
  for c in cases {
    test "${c.text}" {
      Test.parallel()
      assert(c.text.toInt() == c.want)
    }
  }
}
```

```text
ok    3 passed (0.00s)
```

### Running at once

Parallel tests run at the same time. Here one test sends on an unbuffered channel and the other
receives, so neither could finish alone. `-parallel 2` lets both run:

<!-- test: cmd ant test -parallel 2 ping_test.ant -->

```ant [ping_test.ant]
let ball: Channel[string] = Channel.new()

test "ping" {
  Test.parallel()
  ball.send("ping")
}

test "pong" {
  Test.parallel()
  assert(ball.receive() == "ping")
}
```

```text
ok    2 passed (0.00s)
```

With `-parallel 1`, only one runs at a time, and these two wait for each other until `-timeout` ends
the run.

### From a helper

The call acts on the current test, so a helper function called from the test's body can make it
parallel:

<!-- test: cmd ant test helper_test.ant -->

```ant [helper_test.ant]
func independent(): void {
  Test.parallel()
}

test "sum" {
  independent()
  assert([1, 2, 3].sum() == 6)
}

test "max" {
  independent()
  assert([1, 2, 3].max() == 3)
}
```

```text
ok    2 passed (0.00s)
```

## Applications

Give each parallel test its own temporary directory with [Test.tempDir](tempDir.md), so they never
write to the same file:

<!-- test: cmd ant test files_test.ant -->

```ant [files_test.ant]
test "write then read" {
  for text in ["alpha", "beta", "gamma"] {
    test "${text}" {
      Test.parallel()
      let path = Path.join([Test.tempDir(), "data.txt"])
      try File.write(path, text)
      assert(try File.read(path) == text)
    }
  }
}
```

```text
ok    3 passed (0.00s)
```

Count what parallel tests did with a module variable behind a [Mutex](../Mutex/index.md). A serial
test after the parent runs once the parent and all of its parallel subtests have ended:

<!-- test: cmd ant test count_test.ant -->

```ant [count_test.ant]
let lock = Mutex.new()
var finished = 0

test "rows" {
  for n in [1, 2, 3, 4] {
    test "${n}" {
      Test.parallel()
      assert(n > 0)
      lock.withLock({ finished += 1 })
    }
  }
}

test "every row finished" {
  assert(finished == 4)
}
```

```text
ok    5 passed (0.00s)
```

## Properties and relations

A parallel subtest waits for its serial siblings, so its work comes after theirs. A parent's
deferred blocks run before its parallel subtests end, and the next top-level test starts only after
they end:

<!-- test: cmd ant test steps_test.ant -->

```ant [steps_test.ant]
let log: Array[string] = []

test "steps" {
  defer { log.push("deferred") }
  test "parallel" {
    Test.parallel()
    log.push("parallel subtest")
  }
  log.push("body")
  test "serial" {
    log.push("serial subtest")
  }
}

test "order" {
  assert(log.slice(0, 2) == ["body", "serial subtest"])
  assert(log.slice(2, 4).sort() == ["deferred", "parallel subtest"])
}
```

```text
ok    3 passed (0.00s)
```

## Possible issues

A parent's `defer` runs before its parallel subtests end, so it can close what they still need:

<!-- test: cmd ant test conn_test.ant -->
<!-- test: exit 1 -->

```ant [conn_test.ant]
type Conn = { open: bool }

func Conn.query(c: Conn, sql: string): int {
  if !c.open { panic("query on a closed connection: ${sql}") }
  1
}

test "queries" {
  let conn = Conn{ open: true }
  defer { conn.open = false }

  test "select" {
    Test.parallel()
    assert(conn.query("select 1") == 1)
  }
}
```

```text
--- FAIL: conn_test.ant: queries/select (0.00s)
    query on a closed connection: select 1
      at Conn.query (conn_test.ant:4:21)
      at test "select" (conn_test.ant:14:22)
FAIL  1 failed, 0 passed (0.00s)
```

Put the parallel subtests in a serial subtest of their own. It ends only after they do, so the
parent's `defer` runs after them:

<!-- test: cmd ant test group_test.ant -->

```ant [group_test.ant]
type Conn = { open: bool }

func Conn.query(c: Conn, sql: string): int {
  if !c.open { panic("query on a closed connection: ${sql}") }
  1
}

test "queries" {
  let conn = Conn{ open: true }
  defer { conn.open = false }

  test "group" {
    for sql in ["select 1", "select 2"] {
      test "${sql}" {
        Test.parallel()
        assert(conn.query(sql) == 1)
      }
    }
  }
}
```

```text
ok    2 passed (0.00s)
```

A directory from [Test.tempDir](tempDir.md) has no such problem: it is removed after the parallel
subtests end.

The code before the call runs before the test waits, so call `Test.parallel()` first. Here the setup
of the first test runs before the serial test:

<!-- test: cmd ant test -v setup_test.ant -->

```ant [setup_test.ant]
test "parallel" {
  println("setup")
  Test.parallel()
  println("parallel work")
}

test "serial" {
  println("serial work")
}
```

```text
=== RUN   setup_test.ant: parallel
    setup
=== RUN   setup_test.ant: serial
    serial work
--- PASS: setup_test.ant: serial (0.00s)
    parallel work
--- PASS: setup_test.ant: parallel (0.00s)
ok    2 passed (0.00s)
```

A test calls `Test.parallel` once:

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

Only the test's body can call it, not a goroutine the test starts:

<!-- test: cmd ant test go_test.ant -->
<!-- test: exit 1 -->

```ant [go_test.ant]
test "from a goroutine" {
  let done: Channel[bool] = Channel.new()
  go {
    defer { done.send(true) }
    Test.parallel()
  }
  done.receive()
}
```

```text
--- FAIL: go_test.ant: from a goroutine (0.00s)
    goroutine started at go_test.ant:3:3 panicked: Test.parallel: called from a goroutine the test started, not from its body
      at go block (go_test.ant:5:18)
FAIL  1 failed, 0 passed (0.00s)
```

Outside a test it panics:

```ant
Test.parallel() // panic: Test.parallel: not in a test
```

## See also

- [test](../../language/keywords/test.md): tests and subtests
- [Test.tempDir](tempDir.md): a directory of the test's own
- [Test.skip](skip.md): end a test as skipped
- [Test.name](name.md): the current test's full name
- [Mutex](../Mutex/index.md): guard state that parallel tests share
- [ant test](../../tools/test.md): the `-parallel` flag
