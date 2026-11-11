<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result

```ant signature
type Result[T, E] = Ok[T] | Err[E]
```

Result is the outcome of an operation that can fail: an `Ok` with a value or an `Err` with an error.
Test it with is, or use the Result functions.

```ant signature
type Ok[T] = { value: T }
```

Ok is the success case of a `Result`, holding its value.

```ant signature
type Err[E] = { error: E }
```

Err is the failure case of a `Result`, holding its error.

<!-- end generated -->

A function that can fail in an expected way returns a `Result`: `Ok` with its value, or `Err` with
its error. The `Result` namespace holds the functions that build and take results. Each function
that takes a result takes it first, so `Result.unwrapOr(r, 0)`, `r.unwrapOr(0)`, and
`r | Result.unwrapOr(0)` are the same call.

## Details

- `Ok` and `Err` are ordinary records. `Result.ok(v)` is the same value as `Ok{ value: v }`, and
  `Result.err(e)` the same as `Err{ error: e }`.
- `Result.fail(m)` is `Result.err(Error{ message: m })`, for a result whose error type is `Error`.
- `try` unwraps an `Ok`. On an `Err` it returns an `Err` with the same error from the enclosing
  function, which must return a `Result` whose error type accepts it.
- The error type is any type: the library's `Error`, a string, or a union of records the program
  declares.
- [Result.unwrap](unwrap.md) panics on an `Err`, for code where an `Err` is a bug.

## Basic examples

`Result.ok` and `Result.err` build a result. `is` tells the two cases apart:

```ant
let good: Result[int, string] = Result.ok(8080)
let bad: Result[int, string] = Result.err("not a number")
good             // Ok{ value: 8080 }
bad              // Err{ error: "not a number" }
good is Ok       // true
Result.isOk(bad) // false
```

After `is Ok`, the value is in `.value`; in the `else` branch, the error is in `.error`:

```ant
let r = Json.parse("[1, 2]")
if r is Ok {
  println(r.value)
} else {
  println("bad input: ${r.error.message}")
}
```

```text
[1.0, 2.0]
```

`try` unwraps an `Ok` and returns any `Err` to the caller:

```ant
func parseInt(s: string): Result[int, string] {
  let n = s.toInt() ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}

func add(a: string, b: string): Result[int, string] {
  let x = try parseInt(a)
  let y = try parseInt(b)
  Result.ok(x + y)
}

add("2", "3")   // Ok{ value: 5 }
add("2", "two") // Err{ error: "not a number: two" }
```

## Scope

The error type can be a union of records. Narrow the result, then narrow its error:

```ant
type Empty = {}
type TooLong = { max: int }
type NameError = Empty | TooLong

func checkName(s: string): Result[string, NameError] {
  if s.isEmpty() {
    return Result.err(Empty{})
  }
  if s.length() > 8 {
    return Result.err(TooLong{ max: 8 })
  }
  Result.ok(s)
}

for s in ["ada", "", "bartholomew"] {
  let r = checkName(s)
  if r is Ok {
    println("ok: ${r.value}")
  } else {
    let e = r.error
    if e is Empty {
      println("empty name")
    } else {
      println("longer than ${e.max}")
    }
  }
}
```

```text
ok: ada
empty name
longer than 8
```

A `Result[void, E]` reports only success or failure. The standard library's `Error` is a record with
a `message`:

```ant
func save(ok: bool): Result[void, Error] {
  if !ok {
    return Result.fail("disk full")
  }
  Result.ok(void)
}
save(true)  // Ok{ value: void }
save(false) // Err{ error: Error{ message: "disk full" } }
```

A record literal builds the same value as the constructor:

```ant
let a: Result[int, string] = Ok{ value: 1 }
let b: Result[int, string] = Result.ok(1)
a == b // true
```

`try` reads only the outer case. On an `Ok` holding a result, it gives the inner result:

```ant
func outer(): Result[Result[int, string], string] {
  Result.ok(Result.err("inner"))
}
func run(): Result[int, string] {
  let inner = try outer()
  println(inner)
  Result.ok(0)
}
println(run())
```

```text
Err{ error: "inner" }
Ok{ value: 0 }
```

## Functions by job

### Building

- [Result.ok](ok.md): an `Ok` carrying a value
- [Result.err](err.md): an `Err` carrying an error
- [Result.fail](fail.md): an `Err` carrying an `Error` with a message

### Reading

- [Result.isOk](isOk.md): whether a result is an `Ok`
- [Result.unwrap](unwrap.md): the value, or a panic on an `Err`
- [Result.unwrapOr](unwrapOr.md): the value, or a fallback on an `Err`

### Transforming

- [Result.map](map.md): apply a function to the value of an `Ok`
- [Result.mapErr](mapErr.md): apply a function to the error of an `Err`
- [Result.wrapErr](wrapErr.md): put context in front of an `Error` message

### Combining

- [Result.all](all.md): `Ok` of every value, or the first `Err`

## Applications

Add context as an error goes up. Under `try`, parenthesize the call together with its
`Result.wrapErr` stage:

```ant
func loadPort(text: string): Result[int, Error] {
  let json = try (Json.parse(text) | Result.wrapErr("parse settings"))
  if !(json is float64) {
    return Result.fail("port is not a number")
  }
  let port = Math.toInt(json) ?? return Result.fail("port ${json} is not whole")
  Result.ok(port)
}

for text in ["8080", "80.5", "{"] {
  let r = loadPort(text) | Result.wrapErr("load config")
  if r is Ok {
    println("port ${r.value}")
  } else {
    println(r.error.message)
  }
}
```

```text
port 8080
load config: port 80.5 is not whole
load config: parse settings: unexpected end of JSON input
```

In a script's top-level code, `try` on an `Err` ends the script with status 1. This script writes
`parse settings: unexpected end of JSON input` to standard error and prints nothing:

<!-- test: exit 1 -->

```ant
let settings = try (Json.parse("{") | Result.wrapErr("parse settings"))
println(settings)
```

Parse every input, and get the first failure or every value:

```ant
func parsePort(s: string): Result[int, string] {
  let n = s.toInt() ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}

["80", "443"].map(parsePort) | Result.all()  // Ok{ value: [80, 443] }
["80", "http"].map(parsePort) | Result.all() // Err{ error: "not a number: http" }
```

Or keep going, with a fallback for each failure:

```ant
func parsePort(s: string): Result[int, string] {
  let n = s.toInt() ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}

["80", "http"].map({ parsePort(x).unwrapOr(0) }) // [80, 0]
```

## Properties and relations

`Result.isOk(r)` is `r is Ok`, `Result.unwrapOr(r, 0)` is the value of an `Ok` or the fallback, and
`Result.map` keeps an `Err` as it is:

```ant
let rs: Array[Result[int, string]] = [Result.ok(2), Result.err("no")]
for r in rs {
  assert(Result.isOk(r) == (r is Ok))
  assert(Result.unwrapOr(r, 0) == if r is Ok { r.value } else { 0 })
}
let e: Result[int, string] = Result.err("no")
assert(Result.map(e, { x * 2 }) == e)
```

## Possible issues

`Result.ok` alone does not say what the error type is. Annotate the variable:

```ant
let r = Result.ok(1) // error: cannot infer the type parameter E of Result.ok; annotate the result
```

```ant
let r: Result[int, string] = Result.ok(1)
r // Ok{ value: 1 }
```

`try` needs a function that returns a `Result`:

```ant
func first(text: string): float64 {
  let json = try Json.parse(text)
  // error: try would return Err[Error], but the function returns float64
  0.0
}
```

`try` binds tighter than `|`: it unwraps the first stage and pipes the value on. Parenthesize the
whole pipeline:

```ant
func load(text: string): Result[Json.Value, Error] {
  let json = try Json.parse(text) | Result.wrapErr("parse settings")
  // error: cannot use Json.Value where Result[T, Error] is expected
  Result.ok(json)
}
```

```ant
func load(text: string): Result[Json.Value, Error] {
  let json = try (Json.parse(text) | Result.wrapErr("parse settings"))
  Result.ok(json)
}
load("{").isOk() // false
```

A result with a narrower error type does not fit where a wider one is expected, because type
arguments are invariant. `try` widens on the way out; elsewhere, widen with
`Result.mapErr(r, { x })`:

```ant
type Empty = {}
type TooLong = { max: int }
func check(s: string): Result[string, Empty] {
  if s.isEmpty() { Result.err(Empty{}) } else { Result.ok(s) }
}
func validate(s: string): Result[string, Empty | TooLong] {
  check(s)
  // error: cannot use Result[string, Empty] where Result[string, Empty | TooLong] is expected
}
```

```ant
type Empty = {}
type TooLong = { max: int }
func check(s: string): Result[string, Empty] {
  if s.isEmpty() { Result.err(Empty{}) } else { Result.ok(s) }
}
func validate(s: string): Result[string, Empty | TooLong] {
  Result.mapErr(check(s), { x })
}
validate("") // Err{ error: Empty{} }
```

`Result.unwrap` on an `Err` panics. Prefer `try`, `Result.unwrapOr`, or `is`, unless an `Err` is a
bug:

```ant
let r: Result[int, string] = Result.err("no")
Result.unwrap(r) // panic: Result.unwrap on an Err: "no"
```

## See also

- [Error](../Error/index.md), the library's error type, and [Error.wrap](../Error/wrap.md)
- [Panic.catch](../Panic/catch.md): turn a panic into an `Err`
- [try](../../language/keywords/try.md), [is](../../language/operators/is.md)
- [Errors](../guides/errors.md)
