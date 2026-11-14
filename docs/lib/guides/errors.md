<!-- status: reviewed -->

# Errors

Ant has two ways to fail. A failure the caller is meant to handle, such as bad input or a missing
file, is a value: a function returns a [`Result`](../Result/index.md), `Ok` with its value or `Err`
with its error. A bug or an impossible state is a panic: [`panic`](../prelude/panic.md) or a failed
[`assert`](../prelude/assert.md) stops the goroutine. If the program can recover by doing something
else, return a `Result`; if the only fix is to change the code, panic.

The standard library's error type is [`Error`](../Error/index.md), a record with a `message`. An
error that code branches on is a type of its own, often a tagged union such as
`EmptyInput | BadDigit`: build it with [`Result.err`](../Result/err.md), and turn it into an `Error`
with [`Result.mapErr`](../Result/mapErr.md) when a person is to read it.

[`try`](../../language/keywords/try.md) reads a result inside a function: it gives the value of an
`Ok` and passes an `Err` up to the caller. Each caller that knows more adds context in front of the
message with [`Result.wrapErr`](../Result/wrapErr.md), so the final text says what the program was
doing. A `wrapErr` stage goes in parentheses together with the call it wraps:

```ant
type Address = { host: string, port: int }

func parsePort(s: string): Result[int, Error] {
  let n = s.toInt() ?? return Result.fail("${s} is not a number")
  if n < 1 || n > 65535 {
    return Result.fail("${n} is out of range")
  }
  Result.ok(n)
}

func parseAddress(s: string): Result[Address, Error] {
  let cut = s.cut(":") ?? return Result.fail("${s} has no port")
  let port = try (parsePort(cut.after) | Result.wrapErr("parse address ${s}"))
  Result.ok(Address{ host: cut.before, port })
}

for s in ["localhost:8080", "localhost", "localhost:http", "localhost:70000"] {
  let r = parseAddress(s)
  println(if r is Ok { "${r.value.host} port ${r.value.port}" } else { r.error.message })
}
```

```text
localhost port 8080
localhost has no port
parse address localhost:http: http is not a number
parse address localhost:70000: 70000 is out of range
```

A panic carries only a message and a trace. [`Panic.catch`](../Panic/catch.md) stops one at the edge
of a unit of work, such as a job or a request, so that a bug in one does not end the others:

```ant
func percent(part: int, whole: int): int {
  assert(whole > 0, "callers pass a nonempty total")
  part * 100 / whole
}

for whole in [8, 0] {
  let r = Panic.catch({ percent(2, whole) })
  println(if r is Ok { "${r.value}%" } else { "bug: ${r.error.message}" })
}
```

```text
25%
bug: assertion failed: whole > 0 (0 > 0): callers pass a nonempty total
```

## Functions by job

### Building a result

- [Result.ok](../Result/ok.md): an `Ok` carrying a value
- [Result.err](../Result/err.md): an `Err` carrying an error of any type
- [Result.fail](../Result/fail.md): an `Err` carrying an `Error` with a message; only where the
  error type is `Error` itself, so under a wider error type use `Result.err`

### Reading a result

- [try](../../language/keywords/try.md): the value of an `Ok`, or return the `Err` to the caller
- [is](../../language/operators/is.md): tell `Ok` from `Err`, then read `.value` or `.error`
- [Result.isOk](../Result/isOk.md): whether a result is an `Ok`
- [Result.unwrap](../Result/unwrap.md): the value, or a panic on an `Err`
- [Result.unwrapOr](../Result/unwrapOr.md): the value, or a fallback on an `Err`

### Transforming a result

- [Result.map](../Result/map.md): apply a function to the value of an `Ok`
- [Result.mapErr](../Result/mapErr.md): apply a function to the error of an `Err`, such as to widen
  its type or turn it into an `Error`

### Combining results

- [Result.all](../Result/all.md): `Ok` of every value, or the first `Err`

### Adding context

- [Error](../Error/index.md): the library's error, a record with a `message`
- [Error.wrap](../Error/wrap.md): a new `Error` with context in front of the message
- [Result.wrapErr](../Result/wrapErr.md): wrap the error of an `Err`, and keep an `Ok` as it is

### Panicking

- [panic](../prelude/panic.md): stop the goroutine with a message
- [assert](../prelude/assert.md): panic when a condition is false, showing the values it read

### Catching a panic

- [Panic.catch](../Panic/catch.md): call a function and return `Ok` of its value, or `Err` of the
  panic that left it
- [Panic.raise](../Panic/raise.md): panic again with a caught panic's message and trace
- [Panic](../Panic/index.md): a caught panic, its message and trace
- [Panic.Frame](../Panic/types/Frame.md): one frame of a trace

### Cleaning up

- [defer](../../language/keywords/defer.md): run a block when the enclosing block exits, however it
  exits, a panic or an error passed on with `try` included

## See also

- [Errors](../../guide/errors.md): how `Result`, `panic`, `assert`, `defer`, and `Panic.catch` fit
  together
- [Tour: errors](../../tour/13-errors.md) and [Tour: panics](../../tour/14-panics.md)
- [Add context to an error](../../howto/errors/context.md)
- [Turn a panic into an error at a boundary](../../howto/errors/boundary.md)
