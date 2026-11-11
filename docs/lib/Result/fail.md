<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.fail

```ant signature
func Result.fail[T](message: string): Result[T, Error]
```

Result.fail returns an Err carrying an Error with the message; Result.fail(m) is the same value as
Result.err(Error\{ message: m \}). Its error type is Error itself, so where a wider one is expected,
such as ParseError \| Error, write Result.err.

<!-- end generated -->

## Details

- A message is lowercase and has no closing punctuation, since [Result.wrapErr](wrapErr.md) may put
  a context in front of it.

## Basic examples

Fail with a message. The annotation fixes the value type:

```ant
let r: Result[int, Error] = Result.fail("no input")
r // Err{ error: Error{ message: "no input" } }
```

Return a failure from a function. The declared return type fixes the value type:

```ant
func half(n: int): Result[int, Error] {
  if n % 2 != 0 {
    return Result.fail("${n} is odd")
  }
  Result.ok(n / 2)
}

half(8) // Ok{ value: 4 }
half(3) // Err{ error: Error{ message: "3 is odd" } }
```

Test the result with `is` and read the message of its error:

```ant
let r: Result[int, Error] = Result.fail("timeout")
if r is Err { r.error.message } // "timeout"
```

## Scope

The value type can be any type, `void` included, for an operation that has no value to return:

```ant
let saved: Result[void, Error] = Result.fail("disk full")
saved // Err{ error: Error{ message: "disk full" } }
```

The message can be any string, even an empty one:

```ant
let blank: Result[int, Error] = Result.fail("")
blank // Err{ error: Error{ message: "" } }
```

Elements of an annotated array take their type from it:

```ant
let rs: Array[Result[int, Error]] = [Result.ok(1), Result.fail("a"), Result.fail("b")]
Result.all(rs) // Err{ error: Error{ message: "a" } }
```

A function that returns a `Result` or `null` can fail the same way:

```ant
func setting(key: string): Result[int, Error] | null {
  if key.isEmpty() {
    return Result.fail("empty key")
  }
  null
}

setting("")     // Err{ error: Error{ message: "empty key" } }
setting("port") // null
```

Call it in a pipe:

```ant
let r: Result[int, Error] = "bad input" | Result.fail()
r // Err{ error: Error{ message: "bad input" } }
```

Name the argument to say what the string is:

```ant
let r: Result[int, Error] = Result.fail(message: "no input")
r // Err{ error: Error{ message: "no input" } }
```

A parameter type fixes the value type too:

```ant
func orZero(r: Result[int, Error]): int {
  r.unwrapOr(0)
}

orZero(Result.fail("no input")) // 0
```

Pass it as a function value to turn messages into failures:

```ant
let toErr: func(string): Result[int, Error] = Result.fail
println(["a", "b"] | Array.map(toErr))
```

```text
[Err{ error: Error{ message: "a" } }, Err{ error: Error{ message: "b" } }]
```

## Applications

Return early with `?? return` when a lookup comes back `null`:

```ant
let users = Map{ 1: "Ann", 2: "Bob" }

func greet(id: int): Result[string, Error] {
  let name = users[id] ?? return Result.fail("no such user ${id}")
  Result.ok("Hello, ${name}")
}

println(greet(1))
println(greet(7))
```

```text
Ok{ value: "Hello, Ann" }
Err{ error: Error{ message: "no such user 7" } }
```

Reject invalid input in guards, one message each, and keep the successful path last. The caller adds
what it was doing with [Result.wrapErr](wrapErr.md) and propagates the failure with
[try](../../language/keywords/try.md):

```ant
func port(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("port must be an integer")
  }
  if n < 1 || n > 65535 {
    return Result.fail("port ${n} out of range")
  }
  Result.ok(n)
}

func address(text: string): Result[string, Error] {
  let p = try (port(text) | Result.wrapErr("read app.conf"))
  Result.ok("localhost:${p}")
}

println(address("8080"))
println(address("http"))
println(address("70000"))
```

```text
Ok{ value: "localhost:8080" }
Err{ error: Error{ message: "read app.conf: port must be an integer" } }
Err{ error: Error{ message: "read app.conf: port 70000 out of range" } }
```

A script that returns an `Err` ends with exit status 1 and writes the message to standard error:

<!-- test: exit 1 -->

```ant
let args = ["--verbose"]
if !args.contains("--input") {
  return Result.fail("missing --input")
}
println("running")
```

## Properties and relations

`Result.fail(m)` is the same value as [Result.err](err.md) with an `Error`, and as the `Err`
literal:

```ant
let short: Result[int, Error] = Result.fail("x")
let long: Result[int, Error] = Result.err(Error{ message: "x" })
assert(short == long)
assert(short == Err{ error: Error{ message: "x" } })
```

The other `Result` functions pass the `Err` through or fall back:

```ant
let r: Result[int, Error] = Result.fail("x")
assert(!r.isOk())
assert(r.map({ x + 1 }) == r)
assert(r.unwrapOr(0) == 0)
```

Two calls build two records, equal but not the same:

```ant
let a: Result[int, Error] = Result.fail("x")
let b: Result[int, Error] = Result.fail("x")
assert(a == b)
assert(!Record.same(a, b))
```

## Possible issues

Without an expected type, nothing fixes `T`, and the call is a compile error:

```ant
let r = Result.fail("no input")
// error: cannot infer the type parameter T of Result.fail; annotate the result
```

The error type of the result is `Error` itself. Where a wider error type is expected, the call is a
compile error:

```ant
type BadDigit = { at: int }
type ParseError = BadDigit | Error

func digits(s: string): Result[int, ParseError] {
  if s.isEmpty() {
    return Result.fail("empty input")
    // error: cannot use Result.fail where Result[int, ParseError] is expected: it returns Result[T, Error]
  }
  let n = s.toInt() ?? return Result.err(BadDigit{ at: 0 })
  Result.ok(n)
}
```

Write [Result.err](err.md) there, which takes its error type from the expected one:

```ant
type BadDigit = { at: int }
type ParseError = BadDigit | Error

func digits(s: string): Result[int, ParseError] {
  if s.isEmpty() {
    return Result.err(Error{ message: "empty input" })
  }
  let n = s.toInt() ?? return Result.err(BadDigit{ at: 0 })
  Result.ok(n)
}

digits("")   // Err{ error: Error{ message: "empty input" } }
digits("x7") // Err{ error: BadDigit{ at: 0 } }
digits("42") // Ok{ value: 42 }
```

`Result.unwrap` on the result panics:

```ant
let r: Result[int, Error] = Result.fail("no input")
r.unwrap() // panic: Result.unwrap on an Err: Error{ message: "no input" }
```

## See also

[Result.err](err.md) · [Result.ok](ok.md) · [Result.wrapErr](wrapErr.md) ·
[Result.mapErr](mapErr.md) · [Result.unwrapOr](unwrapOr.md) · [Result](index.md) ·
[Error.wrap](../Error/wrap.md) · [try](../../language/keywords/try.md)
