<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.mapErr

```ant signature
func Result.mapErr[T, E, F](r: Result[T, E], f: func(E): F): Result[T, F]
```

Result.mapErr returns r with f applied to its error if r is an Err, and r's value unchanged if it is
an Ok. With an identity function, as in Result.mapErr(r, \{ x \}), it widens the error type.

<!-- end generated -->

## Details

- `f` returns the new error itself, not a `Result`. To change the value of an `Ok`, use
  [Result.map](map.md).

## Basic examples

Change the error of an `Err`:

```ant
let failed: Result[int, string] = Result.err("not found")
Result.mapErr(failed, { String.toUpper(x) }) // Err{ error: "NOT FOUND" }
```

An `Ok` passes through, and its error type changes:

```ant
let found: Result[int, string] = Result.ok(42)
Result.mapErr(found, { String.length(x) }) // Ok{ value: 42 }
```

Turn an error of your own into an `Error` with a readable message:

```ant
type NotFound = { key: string }
let r: Result[int, NotFound] = Result.err(NotFound{ key: "port" })
Result.mapErr(r, { Error{ message: "missing key ${x.key}" } }) // Err{ error: Error{ message: "missing key port" } }
```

## Scope

Read a field of an `Error`:

```ant
let r: Result[string, Error] = Result.err(Error{ message: "unexpected end of input" })
Result.mapErr(r, { x.message }) // Err{ error: "unexpected end of input" }
```

Map each member of an error union with `is`:

```ant
type EmptyInput = {}
type BadDigit = { at: int }
type ParseError = EmptyInput | BadDigit

func describe(e: ParseError): string {
  if e is BadDigit { "bad digit at ${e.at}" } else { "nothing to parse" }
}

let empty: Result[int, ParseError] = Result.err(EmptyInput{})
let bad: Result[int, ParseError] = Result.err(BadDigit{ at: 2 })
Result.mapErr(empty, describe) // Err{ error: "nothing to parse" }
Result.mapErr(bad, describe)   // Err{ error: "bad digit at 2" }
```

Call it through `.` or in a pipe:

```ant
let r: Result[bool, int] = Result.err(404)
r.mapErr({ "status ${x}" })     // Err{ error: "status 404" }
r | Result.mapErr({ x >= 500 }) // Err{ error: false }
```

Name the arguments to pass them in any order:

```ant
let r: Result[int, string] = Result.err("eof")
Result.mapErr(f: { "${x}!" }, r: r) // Err{ error: "eof!" }
```

## Applications

Inside a function, convert the error before `try` returns it. The function's error type is `Error`,
so the error of `parsePort` has to become one:

```ant
type BadPort = { text: string }

func parsePort(s: string): Result[int, BadPort] {
  let n = String.toInt(s) ?? return Result.err(BadPort{ text: s })
  Result.ok(n)
}

func listen(s: string): Result[string, Error] {
  let port = try (parsePort(s) | Result.mapErr({ Error{ message: "bad port ${x.text}" } }))
  Result.ok("listening on ${port}")
}

println(listen("8080"))
println(listen("80x"))
```

```text
Ok{ value: "listening on 8080" }
Err{ error: Error{ message: "bad port 80x" } }
```

Return a narrower `Result` from a function with a wider error type. `try` widens on its own, but a
`Result` returned as it is needs `Result.mapErr` with `{ x }`:

```ant
type NotFound = { key: string }

func lookup(key: string): Result[int, NotFound] {
  if key == "port" { Result.ok(8080) } else { Result.err(NotFound{ key: key }) }
}

func portOf(key: string | null): Result[int, NotFound | Error] {
  if key == null {
    return Result.err(Error{ message: "no key" })
  }
  Result.mapErr(lookup(key), { x })
}

println(portOf(null))
println(portOf("port"))
```

```text
Err{ error: Error{ message: "no key" } }
Ok{ value: 8080 }
```

Put results of different error types in one array before `Result.all`:

```ant
type Timeout = { ms: int }
let a: Result[int, string] = Result.ok(1)
let b: Result[int, Timeout] = Result.err(Timeout{ ms: 500 })
let results: Array[Result[int, string | Timeout]] = [Result.mapErr(a, { x }), Result.mapErr(b, { x })]
Result.all(results) // Err{ error: Timeout{ ms: 500 } }
```

## Properties and relations

On an `Ok`, the result holds the value of `r`, whatever `f` does:

```ant
let r: Result[int, string] = Result.ok(7)
assert(Result.mapErr(r, { String.toUpper(x) }) == r)
```

`Result.mapErr` never changes whether a result is an `Ok`:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("no")]
for r in rs {
  assert(Result.isOk(Result.mapErr(r, { String.length(x) })) == Result.isOk(r))
}
```

With `Error.wrap` as `f`, it gives what [Result.wrapErr](wrapErr.md) gives:

```ant
let r: Result[int, Error] = Result.err(Error{ message: "eof" })
assert(Result.mapErr(r, { Error.wrap(x, "read") }) == Result.wrapErr(r, "read"))
```

Two maps in a row are one map with the composed function:

```ant
let r: Result[int, string] = Result.err("abc")
let f: func(string): int = { String.length(x) }
let g: func(int): int = { x * 10 }
assert(Result.mapErr(Result.mapErr(r, f), g) == Result.mapErr(r, { g(f(x)) }))
```

## Possible issues

A `Result` with a narrower error type is not assignable to one with a wider error type. Widen it
with `Result.mapErr(r, { x })`:

```ant
type NotFound = { key: string }
let narrow: Result[int, NotFound] = Result.err(NotFound{ key: "a" })
let wide: Result[int, NotFound | Error] = narrow
// error: cannot use Result[int, NotFound] where Result[int, NotFound | Error] is expected
```

`f` returns the error, not a `Result`. A function that returns `Result.err` nests one `Err` inside
another:

```ant
func reword(e: string): Result[int, string] {
  Result.err("read: ${e}")
}

let r: Result[int, string] = Result.err("eof")
Result.mapErr(r, reword)           // Err{ error: Err{ error: "read: eof" } }
Result.mapErr(r, { "read: ${x}" }) // Err{ error: "read: eof" }
```

An unannotated `Result.ok` has no error type to map from:

```ant
let r = Result.mapErr(Result.ok(1), { x })
// error: cannot infer the type parameter E of Result.ok; annotate the result
```

`try` binds like a prefix operator, so `try parsePort(s) | Result.mapErr(...)` applies `try` to
`parsePort(s)` alone:

```ant
type BadPort = { text: string }

func parsePort(s: string): Result[int, BadPort] {
  let n = String.toInt(s) ?? return Result.err(BadPort{ text: s })
  Result.ok(n)
}

func listen(s: string): Result[string, Error] {
  let port = try parsePort(s) | Result.mapErr({ Error{ message: "bad port ${x.text}" } })
  // error: try would return Err[BadPort], but the function returns Result[string, Error]
  // error: cannot use int where Result[T, E] is expected
  Result.ok("listening on ${port}")
}
```

Put the call and its `Result.mapErr` stage in parentheses after `try`.

## See also

[Result.map](map.md) · [Result.wrapErr](wrapErr.md) · [Result.all](all.md) ·
[Result.unwrapOr](unwrapOr.md) · [Result](index.md) · [Error.wrap](../Error/wrap.md) ·
[try](../../language/keywords/try.md)
