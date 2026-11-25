<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error

```ant signature
type Error = { message: string, kind: Error.Kind | null = null }
```

Error is the error the standard library returns in an [`Err`](../Result/index.md): a message for a
person to read, and a kind for the program where the library can say what failed, null otherwise.
[`Error.wrap`](wrap.md) and [`Result.wrapErr`](../Result/wrapErr.md) add context in front of the
message and keep the kind.

<!-- end generated -->

The I/O namespaces, such as [Json.parse](../Json/parse.md) and [File.read](../File/read.md), return
it in the `Err` of an operation that fails.

## Details

- `message` is for a person and `kind` is for the program. The library sets a [kind](types/Kind.md)
  for a missing path, a path that is already taken, and an operation the OS denied; its other errors
  have `kind: null`, as does an `Error` you build without one.
- There is no cause to walk, and a program cannot add kinds. An error of your own that code branches
  on is a type of your own, usually a union of records.
- A message is lowercase and has no closing punctuation, so that it reads well inside a longer one.

## Basic examples

Build an `Error` with a record literal and read its message:

```ant
let e = Error{ message: "disk full" }
e         // Error{ message: "disk full", kind: null }
e.message // "disk full"
```

A library call that fails returns an `Err` holding an `Error`:

```ant
Json.parse("{") // Err{ error: Error{ message: "unexpected end of JSON input", kind: null } }
```

Narrow the result with `is`, then read the message:

```ant
let r = Json.parse("[1, 2")
if r is Err {
  println("bad input: ${r.error.message}")
}
```

```text
bad input: unexpected end of JSON input
```

Test the kind to treat one failure differently from the rest:

```ant
let saved = File.read("theme.txt")
if saved is Err && saved.error.kind is Error.NotExist {
  println("no theme saved, using light")
}
```

```text
no theme saved, using light
```

## Scope

Return an `Error` from a function of your own. [Result.fail](../Result/fail.md) builds the `Err`
from the message:

```ant
func half(n: int): Result[int, Error] {
  if n % 2 != 0 {
    return Result.fail("${n} is odd")
  }
  Result.ok(n / 2)
}
half(10) // Ok{ value: 5 }
half(7)  // Err{ error: Error{ message: "7 is odd", kind: null } }
```

An `Error` is an ordinary value outside a `Result` too, such as a check that returns one or `null`:

```ant
func checkPort(port: int): Error | null {
  if port < 1 || port > 65535 {
    return Error{ message: "port ${port} is out of range" }
  }
  null
}
checkPort(8080)  // null
checkPort(70000) // Error{ message: "port 70000 is out of range", kind: null }
```

An `Error` you build has no kind unless you give it one of the library's:

```ant
Error{ message: "disk full" }.kind // null
let missing = Error{ message: "open a.txt: not found", kind: Error.NotExist{} }
missing.kind is Error.NotExist // true
```

A script whose body returns an `Err`, here through `try`, writes the error's message alone and a
newline to standard error, and exits with status 1:

<!-- test: exit 1 -->

```ant
let n = try (Json.parse("{") | Result.wrapErr("parse input"))
println(n)
```

This script writes `parse input: unexpected end of JSON input` to standard error and prints nothing.

## Functions by job

- [Error.wrap](wrap.md): a new `Error` with context in front of the message
- [Result.wrapErr](../Result/wrapErr.md): the same, applied to the error of an `Err`
- [Error.Kind](types/Kind.md): [Error.NotExist](types/NotExist.md), [Error.Exist](types/Exist.md) or
  [Error.Permission](types/Permission.md)

## Applications

Add context at each level, so the final message says what the program was doing:

```ant
func readPort(text: string): Result[float64, Error] {
  let json = try (Json.parse(text) | Result.wrapErr("parse settings"))
  if !(json is float64) {
    return Result.fail("port is not a number")
  }
  Result.ok(json)
}

for text in ["8080", "true", "{"] {
  let r = readPort(text) | Result.wrapErr("load config")
  if r is Ok {
    println("port ${r.value}")
  } else {
    println(r.error.message)
  }
}
```

```text
port 8080.0
load config: port is not a number
load config: parse settings: unexpected end of JSON input
```

Keep an error type of your own for the code that branches on it, and turn it into an `Error` with
[Result.mapErr](../Result/mapErr.md) where a person reads it:

```ant
type Empty = {}
type TooLong = { max: int }

func checkName(s: string): Result[string, Empty | TooLong] {
  if s.isEmpty() {
    return Result.err(Empty{})
  }
  if s.length() > 8 {
    return Result.err(TooLong{ max: 8 })
  }
  Result.ok(s)
}

func describe(e: Empty | TooLong): Error {
  if e is Empty {
    Error{ message: "name is empty" }
  } else {
    Error{ message: "name is longer than ${e.max}" }
  }
}

for s in ["ada", "", "bartholomew"] {
  let r = checkName(s) | Result.mapErr(describe)
  if r is Err {
    println(r.error.message)
  }
}
```

```text
name is empty
name is longer than 8
```

Turn a caught panic into an `Error`, for a job that must not stop the others:

```ant
func work(n: int): int {
  if n == 0 {
    panic("no work")
  }
  100 / n
}

for n in [4, 0] {
  let r = Panic.catch({ work(n) }) | Result.mapErr({ Error{ message: x.message } })
  println(r)
}
```

```text
Ok{ value: 25 }
Err{ error: Error{ message: "no work", kind: null } }
```

## Properties and relations

Errors compare by message and kind, and `Error.wrap` builds the joined message and keeps the kind:

```ant
let e = Error{ message: "timeout" }
assert(e == Error{ message: "timeout" })
assert(e != Error{ message: "timeout", kind: Error.NotExist{} })
assert(Error.wrap(e, "fetch") == Error{ message: "fetch: ${e.message}" })
```

`Result.wrapErr` wraps the error of an `Err` and keeps an `Ok` as it is:

```ant
let e = Error{ message: "timeout" }
let bad: Result[int, Error] = Result.err(e)
let wrapped: Result[int, Error] = Result.err(Error.wrap(e, "fetch"))
let good: Result[int, Error] = Result.ok(1)
assert((bad | Result.wrapErr("fetch")) == wrapped)
assert((good | Result.wrapErr("fetch")) == good)
```

## Possible issues

`message` has no default, and `Error` has no field besides `message` and `kind`:

```ant
let empty = Error{}
// error: Error literal is missing the field message
let coded = Error{ message: "not found", code: 404 }
// error: Error has no field code
```

A [Panic](../Panic/index.md) has a `message` too, but it is not an `Error`. Build one from the
message:

```ant
let caught = Panic.catch({ panic("boom") })
if caught is Err {
  let e: Error = caught.error // error: cannot use Panic where Error is expected
}
```

## See also

[Error.wrap](wrap.md), [Error.Kind](types/Kind.md), [Result](../Result/index.md),
[Result.wrapErr](../Result/wrapErr.md), [Result.mapErr](../Result/mapErr.md),
[Panic](../Panic/index.md), [try](../../language/keywords/try.md), [Errors](../guides/errors.md)
