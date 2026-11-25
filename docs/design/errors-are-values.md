<!-- status: reviewed -->

# Why no exceptions

Ant has no `throw` and no `catch` block. A function that can fail returns a `Result`, and a bug
panics. This page explains why Ant splits failure that way, and what the split costs.

## Two kinds of failure

A missing file, malformed JSON, and a refused connection are expected. The program can do something
about them: report them, retry, use another source. An index out of range, an integer division by
zero, and a state the code believes impossible are bugs. The only fix is to change the code.

Exceptions put both kinds through one mechanism. Ant gives each its own, and gives absence a third:

| Failure                        | Ant              | Who handles it                     |
| ------------------------------ | ---------------- | ---------------------------------- |
| Expected, such as bad input    | A `Result` value | The caller, at the call            |
| Absence, such as a missing key | `null`           | The caller, with `??` or a guard   |
| A bug                          | A panic          | The programmer, by fixing the code |

The rule that picks between the first and the last: if the program can recover by doing something
else, it is a `Result`; if the only fix is to change the code, it is a panic.

## A failure is part of the type

```ant
func parsePort(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port ${text} is not a number")
  }
  Result.ok(n)
}

let port = parsePort("8080")
let next = port + 1 // error: + needs both operands of one type, got Result[int, Error] and int
```

The signature of `parsePort` says it can fail and what it fails with. A reader sees that without
opening the body, and so does the type checker: the value inside an `Ok` is out of reach until the
code looks at which case it holds. With exceptions, any call may throw, and nothing at the call says
which ones do.

## Every early exit is marked

<!-- test: file app.conf
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

println(loadConfig("app.conf"))
```

```text
Err{ error: Error{ message: "parse app.conf: port 80x is not a number", kind: null } }
```

`try` and `return` are the only ways an error leaves `loadConfig` early, and each is written where
it happens. Reading the function top to bottom, you see every point where it may stop. An exception
unwinds through every frame between `throw` and `catch`, and none of those frames shows that it can
happen.

The call that passes an error on is also the natural place to add context. `Result.wrapErr` puts
what the caller was doing in front of the message, so the final message reads as one line from the
outermost operation down to the cause.

## Errors are ordinary values

A `Result` is a union of two records, `Ok{ value }` and `Err{ error }`. Everything that works on a
value works on it: store it, map over it, count it, send it over a channel.

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
let results = ["80", "http", "443"].map(parsePort)
results.countWhere({ x is Err }) // 1
Result.all(results)              // Err{ error: Error{ message: "port http is not a number", kind: null } }
```

No second set of rules applies to errors. A function that validates a whole form keeps every `Err`;
one that needs every value stops at the first with `Result.all` or a loop with `try`. The choice is
ordinary code, not a question of where a `catch` sits.

The error type is the author's choice too. `Error`, a record with a `message`, is the one the
standard library uses, for errors a person reads. Code that branches on the kind of failure declares
a union of its own, and since unions are flat, a function that both reads a file and parses it
returns `Result[Config, ParseError | Error]` with no wrapper classes and no hierarchy.

## Bugs panic

```ant
let xs = [1, 2, 3]
xs[5]     // panic: index 5 out of range for an array of length 3
xs.get(5) // null
```

If indexing returned a `Result`, every `xs[i]` would need a `try`, and the callers of every function
that indexes would need one too. The noise would hide the failures that matter, and no caller could
do anything useful with the error: the index is wrong because the code is wrong. So a bug stops the
goroutine at once, with a stack trace that points at the line. Integer overflow, integer division by
zero, a lossy numeric conversion, `Result.unwrap` on an `Err`, and a failed `assert` all work the
same way.

When an index may be out of range for a reason that is not a bug, such as one read from input,
`xs.get(i)` returns `null` instead, and the caller handles the absence with `??` or a guard.

A panic in the main goroutine ends the program. A panic in any other goroutine ends only that
goroutine, after its deferred blocks run, and the runtime reports it on standard error.

## `Panic.catch` is a boundary, not a handler

`Panic.catch` turns a panic into a value. It exists for the edge of a unit of work that must not
take the others down with it: a job in a queue, a request in a server, a test, a plugin. A worker
goroutine calls each job through it, so a bug in one job ends that job and the worker goes on to the
next.

It is not a second way to fail. A panic carries only a message and a trace, so there is nothing to
branch on, and a function that panics on bad input forces every caller to catch. A failure a caller
is meant to handle is a `Result`. A panic can also leave shared data half changed, and the code that
catches it decides whether that data can still be used.

## What it costs

The design has a price, paid in the code you write:

- **A word at every fallible call.** Each call that can fail needs `try` or code that looks at the
  `Result`. Ant keeps the common case to one word, where Go spends three lines on
  `if err != nil { return err }`.
- **Signatures change up the stack.** When a function starts to fail, it starts to return a
  `Result`, and so does every caller that passes the error on with `try`:

  ```ant
  func parsePort(text: string): Result[int, Error] {
    let n = text.toInt()
    if n == null {
      return Result.fail("port ${text} is not a number")
    }
    Result.ok(n)
  }

  func firstPort(texts: Array[string]): int {
    let port = try parsePort(texts[0])
    // error: try would return Err[Error], but the function returns int
    port
  }
  ```

  The change is visible in review, which is the point, but it touches more code than adding a
  `throw` would.

- **Context by hand.** An `Error` holds a message and a kind, with no stack trace. It says what
  happened only as far as each caller wrapped it, and a wrapped `Error` keeps its text and its kind,
  with no chain of causes to walk. The kind covers the few failures of the operating system every
  program tells apart, a missing path, a taken one, a denied operation; code that needs any other
  cause makes it a field of an error type of its own.
- **Callbacks return results.** `Array.map(xs, parse)` gives an array of `Result`s, not a failure
  halfway through. `Result.all` or a loop with `try` turns them into one.

## Scripts stay short

At the top level of a script, `try` returns from the script body. `ant run` writes the error's
message to standard error and exits with status 1, so a small script needs no error handling beyond
`try`:

<!-- test: exit 1 -->

```ant
let text = try File.read("settings.conf")
println(text.lines().length())
```

The script writes `open settings.conf: no such file or directory` to standard error and prints
nothing. The message names the operation and the file, as the standard library's messages do, so the
script's user sees what went wrong without a stack trace.

## See also

The [errors guide](../guide/errors.md) shows how to use each piece. The tour pages
[Errors](../tour/13-errors.md) and [Panics](../tour/14-panics.md) introduce them. Reference:
[`try`](../language/keywords/try.md), [`Result`](../lib/Result/index.md),
[`Result.wrapErr`](../lib/Result/wrapErr.md), [`Result.all`](../lib/Result/all.md),
[`panic`](../lib/prelude/panic.md), [`assert`](../lib/prelude/assert.md),
[`Panic.catch`](../lib/Panic/catch.md). The design note [Why goroutines and not async](no-async.md)
covers a related decision.
