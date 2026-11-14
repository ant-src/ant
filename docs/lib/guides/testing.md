<!-- status: reviewed -->

# Testing

Tests live in test files, whose names end in `_test.ant`, and [`ant test`](../../tools/test.md) runs
them. A [`test`](../../language/keywords/test.md) declaration names a test and holds its body. The
body checks what it expects with [`assert`](../prelude/assert.md), and a failed `assert` fails the
test. The [`Test`](../Test/index.md) namespace holds the rest a test body needs: skipping, running
in parallel, a temporary directory, a context, and the test's name. Each `Test` function acts on the
current test, so a helper function can call it without being given a handle.

A library, `notes.ant`, and its tests. Each test gets its own temporary directory from
`Test.tempDir`, which `ant test` removes when the test ends:

```ant [notes.ant]
export func save(dir: string, name: string, text: string): Result[string, Error] {
  if name.isEmpty() { return Result.fail("empty note name") }
  let path = Path.join([dir, "${name}.txt"])
  try File.write(path, text)
  Result.ok(path)
}
```

<!-- test: cmd ant test notes_test.ant -->

```ant [notes_test.ant]
import "./notes.ant" { save }

test "save rejects an empty name" {
  let r = save(Test.tempDir(), "", "text")
  assert(r is Err)
  assert(r.error.message == "empty note name")
}

test "save writes the file" {
  let path = try save(Test.tempDir(), "todo", "buy milk")
  let text = try File.read(path)
  assert(text == "buy milk")
}
```

```text
ok    2 passed (0.00s)
```

A failed `assert` reports the condition, the values it compared, the names it read, and the message.
Here `total` has an off-by-one bug and skips the last price:

<!-- test: cmd ant test total_test.ant -->
<!-- test: exit 1 -->

```ant [total_test.ant]
func total(prices: Array[int]): int {
  var sum = 0
  for i in Array.range(0, prices.length() - 1) {
    sum += prices[i]
  }
  sum
}

test "total" {
  let prices = [1, 2, 3]
  assert(total(prices) == 6, "every price is counted")
}
```

```text
--- FAIL: total_test.ant: total (0.00s)
    assertion failed: total(prices) == 6 (3 == 6): every price is counted
      prices = [1, 2, 3]
      at test "total" (total_test.ant:11:9)
FAIL  1 failed, 0 passed (0.00s)
```

A failed `assert` ends the test's body at once, after its deferred blocks run. A failed subtest does
not end its parent: the parent goes on to its next subtest and fails when it ends.

## Functions by job

### Checking

- [assert](../prelude/assert.md): panic, and so fail the test, when a condition is false
- [Panic.catch](../Panic/catch.md): call a function and return the panic that left it, to test that
  code panics

### Controlling a test

- [Test.skip](../Test/skip.md): end the current test as skipped, with a reason
- [Test.parallel](../Test/parallel.md): run the current test alongside its parent's other parallel
  tests

### Resources for a test

- [Test.tempDir](../Test/tempDir.md): a new directory, removed with its contents when the test ends
- [Test.context](../Test/context.md): a context done when the test's body ends, to stop blocking
  calls and goroutines

### Naming

- [Test.name](../Test/name.md): the current test's full name, such as `parse/0x10`

## Applications

A goroutine that a test starts must end with the test. Give its blocking calls `Test.context()`, so
they return once the body ends:

<!-- test: cmd ant test worker_test.ant -->

```ant [worker_test.ant]
test "worker stops with the test" {
  let jobs: Channel[int] = Channel.new()
  let ctx = Test.context()
  go {
    while true {
      let job = jobs.receive(ctx: ctx) ?? break
      assert(job > 0)
    }
  }
  jobs.send(1)
}
```

```text
ok    1 passed (0.00s)
```

Subtests that call `Test.parallel` run alongside each other, after the subtests of the same parent
that do not call it. [`Test.name`](../Test/name.md) gives each its full name:

<!-- test: cmd ant test names_test.ant -->

```ant [names_test.ant]
test "names" {
  for name in ["a", "b"] {
    test "${name}" {
      Test.parallel()
      assert(Test.name() == "names/${name}")
    }
  }
}
```

```text
ok    2 passed (0.00s)
```

To test that code panics, call it through `Panic.catch` and check the `Err`:

<!-- test: cmd ant test stack_test.ant -->

```ant [stack_test.ant]
func pop(stack: Array[int]): int {
  stack.pop() ?? panic("pop from an empty stack")
}

test "pop on an empty stack panics" {
  let r = Panic.catch({ pop([]) })
  assert(r is Err)
  assert(r.error.message.contains("empty"))
}
```

```text
ok    1 passed (0.00s)
```

A test that needs something the environment may lack skips with [`Test.skip`](../Test/skip.md) in a
guard, before its main path. A helper can do the guard, since `Test.skip` acts on the current test.
A skip is not a panic, so `Panic.catch` does not stop it:

<!-- test: cmd ant test db_test.ant -->

```ant [db_test.ant]
func databaseUrl(): string {
  Os.env("DATABASE_URL") ?? Test.skip("DATABASE_URL is not set")
}

test "connects to the database" {
  let url = databaseUrl()
  assert(url.startsWith("postgres://"))
}
```

```text
ok    1 skipped (0.00s)
```

## See also

- [Testing](../../guide/testing.md): how test files, subtests, and `Test` fit together
- [test](../../language/keywords/test.md): the `test` declaration and subtests
- [ant test](../../tools/test.md): the command, its flags, and its report
- [Write table tests](../../howto/testing/table.md) ·
  [Test code that reads and writes files](../../howto/testing/files.md) ·
  [Run tests in parallel](../../howto/testing/parallel.md)
- [Errors](errors.md): `Result`, `Error`, and `Panic`
