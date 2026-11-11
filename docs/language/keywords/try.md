<!-- status: reviewed -->

# `try`

`try r` unwraps a `Result`. On `Ok{ value }` it gives the value; on an `Err` it returns that error
from the enclosing function.

## Details

- On an `Err`, `try` returns a fresh `Err` carrying the same error. The enclosing function must
  return a `Result` whose error type accepts `E`, or such a `Result` or `null`.
- `try` returns from the innermost function body, lambda, or script body. A `go` or `defer` block
  has no result, so `try` cannot appear directly in one.
- In a script's top-level code, `try` on an `Err` ends the script with status 1. It writes the error
  to standard error, an `Error` as its message alone.
- `try` binds like a prefix operator: `try f(x) | g()` unwraps `f(x)` and pipes the value to `g`.
  Parenthesize a pipeline to unwrap its result.

## Basic examples

Unwrap each step and let an `Err` leave the function early:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.trim().toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func sum(a: string, b: string): Result[int, Error] {
  let x = try parseNumber(a)
  let y = try parseNumber(b)
  Result.ok(x + y)
}

println(sum("2", "40"))
println(sum("2", "forty"))
```

```text
Ok{ value: 42 }
Err{ error: Error{ message: "not a number: forty" } }
```

Unwrap the result of a library call:

<!-- test: file port.txt
8080
-->

```ant
func readPort(path: string): Result[string, Error] {
  let text = try File.read(path)
  Result.ok(text.trim())
}

println(readPort("port.txt"))
println(readPort("missing.txt") is Err)
```

```text
Ok{ value: "8080" }
true
```

## Scope

### In a script

At the top level of a script, `try` returns from the script body. On an `Err` the script ends, the
message goes to standard error, and the exit status is 1:

<!-- test: exit 1 -->

```ant
let config = try Json.parse("{\"port\": 8080")
println(config)
```

The script prints nothing to standard output. It writes `unexpected end of JSON input` to standard
error.

### Adding context

Wrap a call in `Result.wrapErr` to say what was being done, and put the parentheses around the whole
pipeline so that `try` unwraps the wrapped result:

```ant
func parsePort(text: string): Result[int, Error] {
  let port = text.trim().toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(port)
}

func loadPort(name: string, text: string): Result[string, Error] {
  let port = try (parsePort(text) | Result.wrapErr("load ${name}"))
  Result.ok("${name} listens on ${port}")
}

println(loadPort("web.conf", "8080"))
println(loadPort("db.conf", "five"))
```

```text
Ok{ value: "web.conf listens on 8080" }
Err{ error: Error{ message: "load db.conf: not a number: five" } }
```

### A whole pipeline

`try` around a parenthesized pipeline unwraps its last stage. Here `Result.all` turns an array of
results into one result, and the first `Err` stops the function:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func total(fields: Array[string]): Result[int, Error] {
  let numbers = try (fields.map(parseNumber) | Result.all())
  Result.ok(numbers.sum() ?? 0)
}

println(total(["1", "2", "3"]))
println(total(["1", "two", "3"]))
```

```text
Ok{ value: 6 }
Err{ error: Error{ message: "not a number: two" } }
```

### In a lambda

A lambda is a boundary of its own: `try` returns the `Err` from the lambda, not from the function
around it. The lambda needs an expected type whose result is a `Result`:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

let doubled: Array[Result[int, Error]] = ["4", "x", "10"].map({ Result.ok(try parseNumber(x) * 2) })
println(doubled)
```

```text
[Ok{ value: 8 }, Err{ error: Error{ message: "not a number: x" } }, Ok{ value: 20 }]
```

### Error unions

Unions are flat, so a function can `try` calls with different error types when its own error type is
their union. Each `try` widens the error on the way out:

```ant
type EmptyInput = {}
type TooLong = { length: int }

func checkName(name: string): Result[string, EmptyInput | TooLong] {
  if name.isEmpty() { return Result.err(EmptyInput{}) }
  if name.length() > 8 { return Result.err(TooLong{ length: name.length() }) }
  Result.ok(name)
}

func parseAge(text: string): Result[int, Error] {
  let years = text.toInt() ?? return Result.fail("bad age: ${text}")
  Result.ok(years)
}

func register(name: string, age: string): Result[string, EmptyInput | TooLong | Error] {
  let checked = try checkName(name)
  let years = try parseAge(age)
  Result.ok("${checked} (${years})")
}

println(register("Ann", "31"))
println(register("", "31"))
println(register("Bartholomew", "31"))
println(register("Ann", "old"))
```

```text
Ok{ value: "Ann (31)" }
Err{ error: EmptyInput{} }
Err{ error: TooLong{ length: 11 } }
Err{ error: Error{ message: "bad age: old" } }
```

### A function that may return `null`

`Err[E]` is assignable to `Result[T, E] | null`, so `try` also works in a function that returns a
`Result` or `null`:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func half(text: string): Result[int, Error] | null {
  if text.isEmpty() { return null }
  let n = try parseNumber(text)
  Result.ok(n / 2)
}

println(half(""))
println(half("8"))
println(half("eight"))
```

```text
null
Ok{ value: 4 }
Err{ error: Error{ message: "not a number: eight" } }
```

### Nested results

`try` reads only the outer tag. On an `Ok` that holds a `Result`, it gives the inner `Result`:

```ant
func outer(): Result[Result[int, Error], Error] {
  Result.ok(Result.fail("inner failure"))
}

func unwrapOnce(): Result[string, Error] {
  let inner = try outer()
  Result.ok("got ${inner}")
}

println(unwrapOnce())
```

```text
Ok{ value: "got Err{ error: Error{ message: \"inner failure\" } }" }
```

## Applications

Copy the first line of one file to another, closing each file on every path out:

<!-- test: file notes.txt
first line
second line
-->

```ant
func copyFirstLine(src: string, dst: string): Result[void, Error] {
  let input = try File.open(src)
  defer { input.close() }
  let line = try input.readLine() ?? ""
  let output = try File.create(dst)
  defer { output.close() }
  try output.writeLine(line)
  output.close()
}

println(copyFirstLine("notes.txt", "first.txt"))
println(File.read("first.txt"))
println(copyFirstLine("missing.txt", "out.txt") is Err)
```

```text
Ok{ value: void }
Ok{ value: "first line\n" }
true
```

Parse `key=value` lines, naming the line that failed:

```ant
func parseLine(line: string): Result[Entry[string, int], Error] {
  let parts = line.split("=")
  if parts.length() != 2 { return Result.fail("expected key=value") }
  let value = parts[1].toInt() ?? return Result.fail("not a number: ${parts[1]}")
  Result.ok(Entry{ key: parts[0], value })
}

func parseAll(text: string): Result[Map[string, int], Error] {
  let settings: Map[string, int] = Map{}
  for line in text.lines() {
    let entry = try (parseLine(line) | Result.wrapErr("line \"${line}\""))
    settings[entry.key] = entry.value
  }
  Result.ok(settings)
}

println(parseAll("width=80\nheight=24"))
println(parseAll("width=80\nheight"))
```

```text
Ok{ value: Map{ "width": 80, "height": 24 } }
Err{ error: Error{ message: "line \"height\": expected key=value" } }
```

## Properties and relations

`try r` in a function is the same as checking the tag and returning the error by hand:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func withTry(text: string): Result[int, Error] {
  let n = try parseNumber(text)
  Result.ok(n + 1)
}

func byHand(text: string): Result[int, Error] {
  let r = parseNumber(text)
  if r is Err { return Result.err(r.error) }
  Result.ok(r.value + 1)
}

for input in ["41", "x"] {
  assert(withTry(input) == byHand(input))
}
```

## Possible issues

### `try` needs a `Result` to return

The enclosing function must return a `Result` that can hold the error:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func double(text: string): int {
  try parseNumber(text) * 2 // error: try would return Err[Error], but the function returns int
}
```

Use `Result.unwrapOr` for a fallback, or `Result.unwrap` when an `Err` is a bug.

### The error type must accept the error

A function whose error type is `Error` cannot `try` a call whose error type is something else. Widen
the function's error type to a union, or convert the error with `Result.mapErr`:

```ant
type Missing = { key: string }

func lookup(key: string): Result[int, Missing] {
  Result.err(Missing{ key })
}

func price(): Result[int, Error] {
  let n = try lookup("tea")
  // error: try would return Err[Missing], but the function returns Result[int, Error]
  Result.ok(n)
}
```

```ant
type Missing = { key: string }

func lookup(key: string): Result[int, Missing] {
  Result.err(Missing{ key })
}

func price(): Result[int, Error] {
  let n = try (lookup("tea") | Result.mapErr({ Error{ message: "no price for ${x.key}" } }))
  Result.ok(n * 2)
}

println(price())
```

```text
Err{ error: Error{ message: "no price for tea" } }
```

### `try` binds before a pipe

`try f(x) | g()` unwraps `f(x)` first. Without parentheses, a `Result.wrapErr` stage receives the
unwrapped value:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func port(text: string): Result[int, Error] {
  let n = try parseNumber(text) | Result.wrapErr("parse port")
  // error: cannot use int where Result[T, Error] is expected
  Result.ok(n)
}
```

Write `try (parseNumber(text) | Result.wrapErr("parse port"))`.

### `try` as the last expression

`try` gives the unwrapped value, so a function that ends in `try` returns a `T`, not a `Result`.
Return the `Result` itself instead:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func parsePort(text: string): Result[int, Error] {
  try parseNumber(text) // error: cannot use int where Result[int, Error] is expected
}
```

### A nullable `Result`

`try` has no nullable form. Handle `null` first, with `??`:

```ant
func first(results: Array[Result[int, Error]]): Result[int, Error] | null {
  let n = try results.first() // error: try needs a Result, got Result[int, Error] | null
  Result.ok(n)
}
```

```ant
func first(results: Array[Result[int, Error]]): Result[int, Error] | null {
  let r = results.first() ?? return null
  let n = try r
  Result.ok(n)
}

println(first([]))
println(first([Result.ok(1)]))
```

```text
null
Ok{ value: 1 }
```

### Inferred return types

Without a return type, the checker infers one from the body. `Result.ok` alone does not say the
error type, so annotate a function that uses `try`:

```ant
func parseNumber(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

func double(text: string) {
  let n = try parseNumber(text)
  // error: try needs an enclosing function that returns a Result, but double returns ? | Err[Error]
  Result.ok(n * 2) // error: cannot infer the type parameter E of Result.ok; annotate the result
}
```

### `go` and `defer` blocks

A `go` or `defer` block has no result, so a `try` in one has nowhere to return the error:

```ant
func save(path: string): Result[void, Error] {
  let file = try File.create(path)
  defer { try file.close() } // error: try needs an enclosing function that returns a Result
  file.writeText("saved")
}
```

### Type parameters

A value of a type parameter is not a `Result`, even when the caller passes one:

```ant
func unwrapAll[T](r: T): Result[int, Error] {
  let n = try r // error: try needs a Result, not a value of the type parameter T
  Result.ok(n)
}
```

## See also

[return](return.md) · [defer](defer.md) · [go](go.md) · [Result](../../lib/Result/index.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [Result.mapErr](../../lib/Result/mapErr.md) ·
[Result.all](../../lib/Result/all.md) · [Result.unwrap](../../lib/Result/unwrap.md) ·
[Result.unwrapOr](../../lib/Result/unwrapOr.md) · [Error.wrap](../../lib/Error/wrap.md) ·
[Pipe](../operators/pipe.md) · [Null coalescing](../operators/null-coalescing.md) ·
[Errors](../../guide/errors.md)
