<!-- status: reviewed -->

# Errors

Ant has two ways to fail. A `Result` is a value: the function returns it, and the caller decides
what to do. A panic is a bug: it stops the goroutine and unwinds its stack. Everything else on this
page, `try`, wrapping, `assert`, `defer`, and `Panic.catch`, serves one of the two.

| Situation                                              | Use                              |
| ------------------------------------------------------ | -------------------------------- |
| Input from a user, a file, or the network may be wrong | Return a `Result`                |
| A caller passes your error on                          | `try`                            |
| A caller knows something the error does not say        | `Result.wrapErr`                 |
| A missing value has an obvious default                 | `??` on a `null`, before parsing |
| The program believes a condition is true               | `assert`                         |
| A state is impossible unless the code is wrong         | `panic`, `Result.unwrap`         |
| A resource must be released however the block exits    | `defer`                          |
| One broken job must not stop the others                | `Panic.catch` at the job's edge  |

The rule behind the table: if the program can recover by doing something else, it is a `Result`; if
the only fix is to change the code, it is a panic.

## Results for expected failure

A function that can fail says so in its return type. This one reads a two-line config file:

<!-- test: file good.conf
web
8080
-->
<!-- test: file bad.conf
web
80x
-->

```ant
type Config = { name: string, port: int }

func parsePort(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port ${text} is not a number")
  }
  if n < 1 || n > 65535 {
    return Result.fail("port ${n} is out of range")
  }
  Result.ok(n)
}

func loadConfig(path: string): Result[Config, Error] {
  let text = try File.read(path)
  let lines = text.lines()
  if lines.length() < 2 {
    return Result.fail("${path}: want a name and a port")
  }
  let port = try (parsePort(lines[1]) | Result.wrapErr("parse ${path}"))
  Result.ok(Config{ name: lines[0], port })
}

for path in ["good.conf", "bad.conf", "missing.conf"] {
  let config = loadConfig(path)
  println(if config is Ok { "${config.value.name} on ${config.value.port}" } else { "skip: ${config.error.message}" })
}
```

```text
web on 8080
skip: parse bad.conf: port 80x is not a number
skip: open missing.conf: no such file or directory
```

`Result[T, E]` is a union of two records, `Ok{ value }` and `Err{ error }`. `Result.ok` and
`Result.err` build them, and `Result.fail("...")` is `Result.err(Error{ message: "..." })`. Each
check is a guard that returns early, so the successful path reads straight down to the final
`Result.ok`.

The caller takes the result apart with `is`, like any union. There is nothing to catch: an error
that nobody looks at is a value nobody looks at, and the type checker makes sure you unwrap a
`Result` before you use its value.

Not every absence is a failure. A key missing from a map, or an empty search, is `null`; the
[null guide](null.md) covers when to pick which. When a failure has an obvious fallback,
`Result.unwrapOr` gives the value or the fallback:

<!-- test: before
func parsePort(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port ${text} is not a number")
  }
  Result.ok(n)
}
-->

```ant
parsePort("8443") | Result.unwrapOr(8080) // 8443
parsePort("80x") | Result.unwrapOr(8080)  // 8080
```

The second line shows the risk. A port the user wrote as `80x` is a mistake to report, not a reason
to use `8080` without a word. Apply a fallback where the value may be absent, before parsing, and
report a value that is present but wrong:

<!-- test: before
func parsePort(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port ${text} is not a number")
  }
  Result.ok(n)
}
let settings = Map{ host: "web" }
-->

```ant
parsePort(settings["port"] ?? "8080") // Ok{ value: 8080 }
parsePort("80x")                      // Err{ error: Error{ message: "port 80x is not a number", kind: null } }
```

The missing `port` gets its default, and a `port` written as `80x` still fails.

## Passing errors on with `try`

`try` unwraps a `Result`. On `Ok` it gives the value; on `Err` it returns the same error from the
enclosing function. So the function that uses `try` returns a `Result` too, with an error type that
accepts the one passed on:

```ant
func half(n: int): Result[int, Error] {
  if n % 2 != 0 {
    return Result.fail("${n} is odd")
  }
  Result.ok(n / 2)
}

func quarter(n: int): int {
  let h = try half(n) // error: try would return Err[Error], but the function returns int
  h / 2
}
```

`try` binds like a prefix operator, tighter than a pipe. `try f(x) | g()` unwraps `f(x)` and pipes
the value on. To pass on the error of a whole pipeline, put it in parentheses:

```ant
func parse(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("${text} is not a number")
  }
  Result.ok(n)
}

func total(texts: Array[string]): Result[int, Error] {
  let ns = try (texts | Array.map(parse) | Result.all())
  Result.ok(ns.sum() ?? 0)
}

println(total(["1", "2", "3"]))
println(total(["1", "two", "3"]))
```

```text
Ok{ value: 6 }
Err{ error: Error{ message: "two is not a number", kind: null } }
```

`Result.all` gives the first `Err`, or `Ok` of every value. `Array.map` runs `parse` on every
element first. A `for` loop with `try` in it stops at the first error instead:

<!-- test: before
func parse(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("${text} is not a number")
  }
  Result.ok(n)
}
-->

```ant
func total(texts: Array[string]): Result[int, Error] {
  var sum = 0
  for text in texts {
    sum += try parse(text)
  }
  Result.ok(sum)
}

println(total(["1", "two", "x"]))
```

```text
Err{ error: Error{ message: "two is not a number", kind: null } }
```

To report every failure instead of the first, keep the errors:

<!-- test: before
func parse(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("${text} is not a number")
  }
  Result.ok(n)
}
-->

```ant
let results = ["1", "two", "x"].map(parse)
for r in results {
  if r is Err {
    println(r.error.message)
  }
}
```

```text
two is not a number
x is not a number
```

At the top level of a script, `try` returns from the script body. The script ends there: `ant run`
writes the error's message to standard error and exits with status 1, after the deferred blocks of
the script body have run.

<!-- test: exit 1 -->

```ant
let text = try File.read("settings.conf")
println(text)
```

The script writes `open settings.conf: no such file or directory` to standard error and prints
nothing. A small script needs no error handling beyond `try`. An error that is not an `Error` is
written as `String.from` prints it, such as `BadDigit{ at: 1 }`, so turn it into an `Error` with
`Result.mapErr` first when a person reads it.

## Error types

`Error` is the error type of the standard library: a `message` for a person to read, and a `kind`
for the few failures of the operating system a program tells apart ([Error kinds](#error-kinds),
below). Use it for errors a person reads. When code must branch on a failure of your own, declare a
type of your own, often a union of records, so the caller can tell the cases apart with `is`.

Unions are flat, so error types combine without wrapping. A function that reads a file and parses it
returns both kinds, and `try` passes each on:

<!-- test: file count.txt
4x
-->
<!-- test: file big.txt
99999999999999999999
-->

```ant
type Empty = {}
type BadDigit = { at: int }
type TooBig = {}

func parseCount(text: string): Result[int, Empty | BadDigit | TooBig] {
  if text == "" {
    return Result.err(Empty{})
  }
  let bad = text.chars().findIndex({ !x.isDigit() })
  if bad != null {
    return Result.err(BadDigit{ at: bad })
  }
  let n = text.toInt()
  if n == null {
    return Result.err(TooBig{})
  }
  Result.ok(n)
}

func readCount(path: string): Result[int, Empty | BadDigit | TooBig | Error] {
  let text = try File.read(path)
  Result.ok(try parseCount(text.trim()))
}

func describe(error: Empty | BadDigit | TooBig | Error): string {
  if error is Empty {
    return "empty"
  }
  if error is BadDigit {
    return "bad digit at ${error.at}"
  }
  if error is TooBig {
    return "too big"
  }
  error.message
}

for path in ["count.txt", "big.txt", "missing.txt"] {
  let r = readCount(path)
  println(if r is Ok { "${r.value}" } else { describe(r.error) })
}
```

```text
bad digit at 1
too big
open missing.txt: no such file or directory
```

`try` widens the error on its way out. A `Result` with the narrower error type is not itself a
`Result` with the wider one, so returning it without `try` is a compile error:

```ant
type BadDigit = { at: int }

func parseCount(text: string): Result[int, BadDigit] {
  Result.ok(1)
}

func readCount(text: string): Result[int, BadDigit | Error] {
  parseCount(text) // error: cannot use Result[int, BadDigit] where Result[int, BadDigit | Error] is expected
}
```

Write `parseCount(text) | Result.mapErr({ x })` to widen it where `try` does not fit. For the same
reason `Result.fail` works only where the error type is `Error` itself; under a wider one, such as
`BadDigit | Error`, write `Result.err(Error{ message: "..." })`.

## Error kinds

A library error says what failed where the operating system can tell. Its `kind` is `Error.NotExist`
when nothing is at the path, `Error.Exist` when something is there already or a directory is not
empty, `Error.Permission` when the OS denied the operation, and `null` for every other failure. Test
it with `is`:

```ant
func loadTheme(path: string): Result[string, Error] {
  let text = File.read(path)
  if text is Err && text.error.kind is Error.NotExist {
    return Result.ok("light")
  }
  Result.ok((try text).trim())
}

println(loadTheme("theme.txt"))
try File.mkdir("themes")
println(loadTheme("themes"))
```

```text
Ok{ value: "light" }
Err{ error: Error{ message: "read themes: is a directory", kind: null } }
```

A missing file gives the default, and any other failure is still an error. A call to `File.exists`
before the read cannot do this: it is `false` for a missing file and also for one the program may
not reach.

- Test the kind on the result of the call it is about. `Result.wrapErr` keeps the kind, so three
  functions up `Error.NotExist` no longer says which path is missing.
- `text.error.kind` does not narrow. To branch on several kinds, bind the kind to a `let` first, and
  end the chain with `else`: most errors have no kind.
- A program cannot add kinds. A failure in your program's own terms is a type of your own, as in
  [Error types](#error-types).

## Adding context

An error says what went wrong; the caller knows what it was doing. `Result.wrapErr` puts that in
front of the message, joined by `: `, and leaves an `Ok` as it is:

```ant
let e = Error{ message: "unexpected end" }
Error.wrap(e, "parse app.json") // Error{ message: "parse app.json: unexpected end", kind: null }
let failed: Result[int, Error] = Result.err(e)
failed | Result.wrapErr("load")          // Err{ error: Error{ message: "load: unexpected end", kind: null } }
Result.ok(8080) | Result.wrapErr("load") // Ok{ value: 8080 }
```

Each layer adds one piece, and the final message reads as one line:
`load config: parse app.json: unexpected end of JSON input`. A few habits keep it that way:

- Write messages in lowercase with no closing punctuation. They end up inside longer ones.
- Name the operation and the argument that matters: `parse app.json`, not
  `failed to parse the config!`. "Error" and "failed" add nothing to an error.
- Do not repeat what the callee already says. `File.read` names the path in its message, so a caller
  of `File.read` adds what it was doing, not the path.

A wrapped error keeps its text and its kind, and nothing else. It is a new record, and there is no
chain of causes to walk. When code needs the cause, make it a field of an error type of your own,
and turn that into an `Error` with `Result.mapErr` where a person reads it.

## Panics for bugs

A panic stops the current goroutine. Many operations panic on their own: an index out of range,
integer overflow, integer division by zero, a lossy numeric conversion. Call `panic` yourself for a
state the code believes impossible, and `Result.unwrap` for an `Err` that can only come from a bug:

```ant
func parsePort(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port ${text} is not a number")
  }
  Result.ok(n)
}

// The default is written in the code, so a parse error here is a bug.
let fallback = Result.unwrap(parsePort("8080"))
println(fallback)
let broken = Result.unwrap(parsePort("80x"))
// panic: Result.unwrap on an Err: Error{ message: "port 80x is not a number", kind: null }
```

```text
8080
```

`try` and `Result.unwrap` both end a script on an `Err`, in different voices. `try` reports an
expected failure: the message alone, exit status 1. `Result.unwrap` reports a bug: a panic with a
stack trace. Never `unwrap` a `Result` that depends on input.

A panic in the main goroutine ends the program. A panic in any other goroutine ends only that
goroutine; the [concurrency guide](concurrency.md) covers how to notice it.

## `assert`

`assert` states what the program believes is true, so a bug shows where it is made instead of where
it is noticed. When the condition is `false`, it panics with the condition, the values it compared,
and your reason:

```ant
type User = { name: string }

func greet(users: Map[string, User], id: string): string {
  let user = users[id]
  assert(user != null, "ids come from users.keys()")
  "hello, ${user.name}"
}

let users = Map{ a: User{ name: "Ann" } }
println(greet(users, "a"))
println(greet(users, "b"))
// panic: assertion failed: user != null (user is null): ids come from users.keys()
```

```text
hello, Ann
```

After `assert(user != null)` the type checker knows `user` is a `User`, as after
`if user == null { panic(...) }`. The reason names why the condition holds, not the condition again.

Assert preconditions, invariants, and values that cannot be `null` here. Never assert input from a
user, a file, or the network: that is a `Result`. `assert` is always on; no flag or build mode
removes it. In a test, a failed `assert` fails the test, as the [testing guide](testing.md) shows.

## Cleaning up with `defer`

`defer` runs a block when the enclosing block exits, however it exits: at its end, by `return`,
`break`, `continue`, or `try`, or by a panic. Register it right after you acquire a resource, so the
release sits next to the acquisition:

<!-- test: file names.txt
Ann
Bob
-->

```ant
func firstLine(path: string): Result[string | null, Error] {
  let file = try File.open(path)
  defer { file.close() }
  file.readLine()
}

println(firstLine("names.txt"))
println(firstLine("missing.txt"))
```

```text
Ok{ value: "Ann" }
Err{ error: Error{ message: "open missing.txt: no such file or directory", kind: Error.NotExist{} } }
```

Deferred blocks run in reverse order of registration, so resources are released in the reverse of
the order they were acquired. `Os.exit` ends the process without running them.

A deferred block has no result: its value is discarded, and `try` directly inside it is a compile
error. That is fine for a file you only read. A file you write can fail when it closes, so close it
yourself on the successful path and report what that returns. The deferred close covers the early
exits and then does nothing:

```ant
func save(path: string, lines: Array[string]): Result[void, Error] {
  let out = try File.create(path)
  defer { out.close() }
  for line in lines {
    try out.writeLine(line)
  }
  out.close()
}

println(save("out.txt", ["a", "b"]))
println(File.read("out.txt"))
```

```text
Ok{ value: void }
Ok{ value: "a\nb\n" }
```

## Catching a panic at a boundary

`Panic.catch(f)` calls `f` and returns `Ok` of its value, or `Err` of the panic that left it, after
the deferred blocks inside `f` have run. The panic is a record with a `message` and a `trace`, the
stack when the panic began, innermost first.

Catch at the edge of an independent unit of work, such as a job, a request, or a plugin, so one bug
does not stop the others:

```ant
type Job = { id: int, divisor: int }

func handle(job: Job): int {
  100 / job.divisor
}

let jobs = [Job{ id: 1, divisor: 4 }, Job{ id: 2, divisor: 0 }, Job{ id: 3, divisor: 5 }]
for job in jobs {
  let done = Panic.catch({ handle(job) })
  if done is Err {
    let frame = done.error.trace[0]
    println("job ${job.id}: ${done.error.message} at line ${frame.line}")
  } else {
    println("job ${job.id}: ${done.value}")
  }
}
```

```text
job 1: 25
job 2: division by zero at line 4
job 3: 20
```

A caught panic is not reported. When the boundary only wants to note the panic, `Panic.raise` sends
it on with its message and trace unchanged, and it ends the goroutine as though it had not been
caught:

```ant
type Job = { id: int, divisor: int }

func handle(job: Job): int {
  100 / job.divisor // panic: division by zero
}

func run(job: Job): int {
  let done = Panic.catch({ handle(job) })
  if done is Err {
    eprintln("job ${job.id} crashed")
    Panic.raise(done.error)
  }
  done.value
}

println(run(Job{ id: 1, divisor: 4 }))
println(run(Job{ id: 2, divisor: 0 }))
```

```text
25
```

`run` writes `job 2 crashed` to standard error, and then the panic ends the program. Its report
points at the division in `handle`, where the panic began, not at `Panic.raise`.

`Panic.catch` is not a way to handle expected failures. A function that panics on bad input forces
every caller to catch; one that returns a `Result` lets every caller decide, and the type checker
reminds them to.

## Coming from other languages

- **Exceptions (Java, Python, JavaScript, C#).** Ant has no `throw` and no `catch` for expected
  failures. A failure is a `Result` in the return type, so you see it at the call. Python's
  `except FileNotFoundError` and Node's `err.code === "ENOENT"` are
  `r.error.kind is Error.NotExist`. `Panic.catch` catches every panic, not one kind, and belongs at
  a boundary, not around each call.
- **`try` is not a block.** It is a prefix on one expression that passes the error to the caller.
  There is no `try { ... } catch`.
- **Go.** `try` replaces `if err != nil { return err }`. There is no `nil` error: a `Result` is
  either `Ok` or `Err`. `Result.wrapErr` is `fmt.Errorf("context: %w", err)`, and the wrapped error
  keeps its kind: `r.error.kind is Error.NotExist` is `errors.Is(err, fs.ErrNotExist)`, with
  `Error.Exist` and `Error.Permission` beside it. There is no chain beyond that and no `errors.As`;
  for an error of your own, branch on a type of your own. `defer` runs when the enclosing block
  exits, not the function.
- **Rust.** `try x` is `x?`, written in front and binding tighter than a pipe. `Error.kind` is
  `io::Error::kind()` with three members and `null` for the rest. `Result.unwrap` panics as Rust's
  does, and `Panic.catch` plays the part of `catch_unwind`.
- **C and Python asserts.** `assert` is always on: no flag or build mode removes it, as `NDEBUG` or
  `python -O` would.

## See also

[`try`](../language/keywords/try.md), [`defer`](../language/keywords/defer.md),
[`Result`](../lib/Result/index.md), [`Result.wrapErr`](../lib/Result/wrapErr.md),
[`Result.mapErr`](../lib/Result/mapErr.md), [`Result.all`](../lib/Result/all.md),
[`Result.unwrap`](../lib/Result/unwrap.md), [`Error.wrap`](../lib/Error/wrap.md),
[`panic`](../lib/prelude/panic.md), [`assert`](../lib/prelude/assert.md),
[`Panic.catch`](../lib/Panic/catch.md), [`Panic.raise`](../lib/Panic/raise.md), and the tour pages
[Errors](../tour/13-errors.md) and [Panics](../tour/14-panics.md).
