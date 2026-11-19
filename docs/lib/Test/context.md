<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.context

```ant signature
func Test.context(): Context
```

Test.context returns the current test's context, done when its body ends, before its deferred
blocks. Give it to blocking calls that must stop when the test does.

<!-- end generated -->

## Details

- Every test has its own context, and a subtest has one apart from its parent's. Within one test,
  every call returns the same context, from the body and from the goroutines the test owns.
- Its reason is `"cancelled"`. A wait in a deferred block that takes it returns at once.
- Called outside a test, `Test.context` panics.

## Basic examples

The context runs while the body runs, and is done with the reason `"cancelled"` once the body ends:

<!-- test: cmd ant test basic_test.ant -->

```ant [basic_test.ant]
test "the context ends with the body" {
  let ctx = Test.context()
  defer {
    assert(ctx.isDone())
    assert(ctx.reason() == "cancelled")
  }
  assert(!ctx.isDone())
}
```

```text
ok    1 passed (0.00s)
```

Hand it to a blocking call, so the wait gives up when the test ends:

<!-- test: cmd ant test receive_test.ant -->

```ant [receive_test.ant]
test "a worker sends its result" {
  let c: Channel[int] = Channel.new()
  go { c.send(42) }
  let got = c.receive(Test.context())
  assert(got == 42)
}
```

```text
ok    1 passed (0.00s)
```

## Scope

### Subtests

A subtest has a context of its own, done when the subtest's body ends, while its parent's context
still runs:

<!-- test: cmd ant test sub_test.ant -->

```ant [sub_test.ant]
test "outer" {
  let outer = Test.context()
  var inner: Context | null = null
  test "inner" {
    inner = Test.context()
    assert(inner != outer)
  }
  assert(inner?.isDone() == true)
  assert(!outer.isDone())
}
```

```text
ok    1 passed (0.00s)
```

### Goroutines the test owns

A goroutine started by the test gets the test's context too, so code running on it needs no handle:

<!-- test: cmd ant test owned_test.ant -->

```ant [owned_test.ant]
test "a goroutine sees the test's context" {
  let fromGoroutine: Channel[Context] = Channel.new(1)
  go { fromGoroutine.send(Test.context()) }
  assert(fromGoroutine.receive() == Test.context())
}
```

```text
ok    1 passed (0.00s)
```

## Applications

### A server that stops with the test

Give it to a goroutine the test starts, so the goroutine stops when the test does. Here a server
accepts connections until the context ends, then closes its listener. It ignores per-connection
errors, so it discards what `answer` returns:

<!-- test: cmd ant test echo_test.ant -->

```ant [echo_test.ant]
func answer(conn: File.Handle, ctx: Context): Result[void, Error] {
  defer { conn.close() }
  let line = try conn.readLine(ctx) ?? return Result.ok(void)
  conn.writeLine(line.toUpper())
}

func serve(addr: string, ctx: Context): Result[string, Error] {
  let listener = try Net.listen(addr)
  go {
    defer { listener.close() }
    while true {
      let conn = listener.accept(ctx)
      if conn is Err { return }
      _ = answer(conn.value, ctx)
    }
  }
  Result.ok(listener.address())
}

test "the server answers in upper case" {
  let addr = try serve("127.0.0.1:0", Test.context())
  let conn = try Net.dial(addr)
  defer { conn.close() }
  try conn.writeLine("hello")
  assert(try conn.readLine() == "HELLO")
}
```

```text
ok    1 passed (0.00s)
```

### One wait bounded by a timeout

Derive a context from the test's with a timeout. The wait ends at the timeout or when the test does,
whichever comes first:

<!-- test: cmd ant test timeout_test.ant -->

```ant [timeout_test.ant]
test "the worker replies in time" {
  let replies: Channel[string] = Channel.new(1)
  go { replies.send("pong") }
  let ctx = Context.new(timeout: 1s, parent: Test.context())
  let reply = replies.receive(ctx)
  assert(reply != null)
}
```

```text
ok    1 passed (0.00s)
```

### A helper that stops with the test

Since `Test.context` acts on the current test, a test helper can take the context itself instead of
asking every caller for one. This heartbeat ticks until the test that started it ends:

<!-- test: cmd ant test beat_test.ant -->

```ant [beat_test.ant]
func startHeartbeat(): Channel[int] {
  let ctx = Test.context()
  let beats: Channel[int] = Channel.new()
  go {
    var n = 0
    while !ctx.isDone() {
      n += 1
      beats.send(n, ctx)
      Time.sleep(10ms, ctx)
    }
  }
  beats
}

test "beats count up" {
  let beats = startHeartbeat()
  assert(beats.receive() == 1)
  assert(beats.receive() == 2)
}
```

```text
ok    1 passed (0.00s)
```

## Properties and relations

Every call within one test returns the same context, and a test's context is not done while its body
runs:

<!-- test: cmd ant test same_test.ant -->

```ant [same_test.ant]
test "one context per test" {
  assert(Test.context() == Test.context())
  assert(!Test.context().isDone())
}
```

```text
ok    1 passed (0.00s)
```

## Possible issues

### A goroutine that ignores the context outlives the test

A goroutine that blocks without the test's context keeps running after the body ends, and the test
fails. Here a goroutine waits on `listener.accept()` with no context:

<!-- test: cmd ant test leak_test.ant -->
<!-- test: exit 1 -->

```ant [leak_test.ant]
test "the listener outlives the test" {
  let listener = try Net.listen("127.0.0.1:0")
  go {
    defer { listener.close() }
    _ = listener.accept()
  }
}
```

```text
--- FAIL: leak_test.ant: the listener outlives the test (0.00s)
    goroutine started at leak_test.ant:3:3 outlived the test
FAIL  1 failed, 0 passed (0.00s)
```

### Deferred blocks run after the context is done

A wait in a deferred block that takes `Test.context()` gives up at once. This one does not wait for
the worker's result:

<!-- test: cmd ant test flush_test.ant -->
<!-- test: exit 1 -->

```ant [flush_test.ant]
test "the worker logs its result" {
  let log: Channel[string] = Channel.new(1)
  go {
    Time.sleep(20ms)
    log.send("written")
  }
  defer {
    let last = log.receive(Test.context())
    assert(last == "written")
  }
}
```

```text
--- FAIL: flush_test.ant: the worker logs its result (0.00s)
    assertion failed: last == "written" (null == "written")
      at test "the worker logs its result" (flush_test.ant:9:11)
FAIL  1 failed, 0 passed (0.00s)
```

Give a deferred wait a context of its own, such as `Context.new(timeout: 1s)`.

### Outside a test

`Test.context` panics when no test is current, as in a script:

```ant
let ctx = Test.context() // panic: Test.context: not in a test
```

## See also

[Context.new](../Context/new.md) · [Context.isDone](../Context/isDone.md) ·
[Context.reason](../Context/reason.md) · [Test.tempDir](tempDir.md) · [Test.parallel](parallel.md) ·
[Test.name](name.md) · [test](../../language/keywords/test.md) · [Test](index.md)
