<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.all

```ant signature
func Result.all[T, E](rs: Array[Result[T, E]]): Result[Array[T], E]
```

Result.all returns Ok of every value in rs, in order, or the first Err in rs.

<!-- end generated -->

## Details

- When several elements are `Err`, the result is the first of them; the later errors are dropped.

## Basic examples

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.ok(2)]
Result.all(rs)                                   // Ok{ value: [1, 2] }
Result.all(Array.concat(rs, [Result.err("no")])) // Err{ error: "no" }
```

Map a function that can fail over an array, then combine the results:

```ant
func parsePort(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("not a number: ${s}")
  if n < 1 || n > 65535 {
    return Result.err("out of range: ${n}")
  }
  Result.ok(n)
}

["80", "443"].map(parsePort) | Result.all()           // Ok{ value: [80, 443] }
["80", "http", "99999"].map(parsePort) | Result.all() // Err{ error: "not a number: http" }
```

## Scope

An empty array gives `Ok` of an empty array. Annotate it so that `T` and `E` are known:

```ant
let none: Array[Result[int, string]] = []
Result.all(none) // Ok{ value: [] }
```

The first `Err` in array order wins, wherever it is:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("a"), Result.err("b")]
Result.all(rs) // Err{ error: "a" }
```

`rs` itself is left as it was:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("a")]
Result.all(rs) // Err{ error: "a" }
rs             // [Ok{ value: 1 }, Err{ error: "a" }]
```

Library functions that return a `Result` combine the same way. `Json.parse` fails with an `Error`:

```ant
["[1]", "{\"a\": 2}"].map(Json.parse) | Result.all() // Ok{ value: [[1.0], Map{ "a": 2.0 }] }
(["[1]", "{"].map(Json.parse) | Result.all()) is Err // true
```

The error type can be a union of records. The `Err` keeps the member it holds:

```ant
type Missing = { field: string }
type TooLong = { field: string, max: int }
type Invalid = Missing | TooLong
let checks: Array[Result[string, Invalid]] = [
  Result.ok("ann"),
  Result.err(Missing{ field: "email" }),
  Result.err(TooLong{ field: "name", max: 20 }),
]
Result.all(checks) // Err{ error: Missing{ field: "email" } }
```

A value type that includes `null` keeps its `null` values:

```ant
let maybe: Array[Result[int | null, string]] = [Result.ok(null), Result.ok(2)]
Result.all(maybe) // Ok{ value: [null, 2] }
```

When the values are results themselves, `T` is a `Result`, and an `Ok` holding an `Err` is a success
like any other:

```ant
let nested: Array[Result[Result[int, string], string]] = [Result.ok(Result.err("inner"))]
Result.all(nested) // Ok{ value: [Err{ error: "inner" }] }
```

`Panic.catch` returns a `Result` too, so `Result.all` reports the first panic of several calls:

```ant
let runs = [Panic.catch({ 1 + 1 }), Panic.catch({ [1][2] })]
Result.all(runs) is Err // true
```

## Applications

Unwrap every result with `try`, in a function that returns a `Result`. Parenthesize the whole
pipeline, so that `try` applies to the combined result:

```ant
func parsePrice(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("bad price: ${s}")
  Result.ok(n)
}

func total(prices: Array[string]): Result[int, string] {
  let cents = try (prices.map(parsePrice) | Result.all())
  Result.ok(cents.reduce(0, func(sum: int, c: int): int { sum + c }))
}

println(total(["120", "80", "5"]))
println(total(["120", "free", "5"]))
```

```text
Ok{ value: 205 }
Err{ error: "bad price: free" }
```

Validate every field of a form and return the cleaned values, or the first problem:

```ant
type Field = { name: string, value: string }

func check(f: Field): Result[string, string] {
  let value = String.trim(f.value)
  if value == "" {
    return Result.err("${f.name} is required")
  }
  Result.ok(value)
}

let form = [Field{ name: "name", value: " Ann " }, Field{ name: "city", value: "Oslo" }]
println(form.map(check) | Result.all())
let partial = [Field{ name: "name", value: "Ann" }, Field{ name: "city", value: "  " }]
println(partial.map(check) | Result.all())
```

```text
Ok{ value: ["Ann", "Oslo"] }
Err{ error: "city is required" }
```

Run a function that can fail on every element in parallel with `Parallel.map`, then combine the
results. `Parallel.map` returns `null` when it was cancelled before every element started:

```ant
func square(n: int, work: Context): Result[int, string] {
  if n < 0 {
    return Result.err("negative: ${n}")
  }
  Result.ok(n * n)
}

func squares(xs: Array[int]): Result[Array[int], string] {
  let results = Parallel.map(xs, square, limit: 2) ?? return Result.err("cancelled")
  let values = try (results | Result.all())
  Result.ok(values)
}

println(squares([1, 2, 3]))
println(squares([1, -2, -3]))
```

```text
Ok{ value: [1, 4, 9] }
Err{ error: "negative: -2" }
```

`Parallel.any` is the counterpart: the first `Ok`, or every error.

Add context to the combined error with `Result.wrapErr`:

```ant
let configs = ["{}", "[1, 2"]
println(configs.map(Json.parse) | Result.all() | Result.wrapErr("load configs"))
```

```text
Err{ error: Error{ message: "load configs: unexpected end of JSON input" } }
```

## Properties and relations

`Result.all` gives an `Ok` exactly when every element is an `Ok`, and its values are what
`Result.unwrap` gives for each element, in order:

```ant
let rs: Array[Result[int, string]] = [Result.ok(3), Result.ok(1), Result.ok(2)]
let all = Result.all(rs)
assert(all is Ok == rs.every(Result.isOk))
assert(Result.unwrap(all) == rs.map(Result.unwrap))
assert(Result.unwrap(all).length() == rs.length())
```

On a failure, the error is the error of the first `Err` that `Array.find` finds:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("x"), Result.err("y")]
let all = Result.all(rs)
let first = rs.find({ x is Err })
assert(all is Err && first is Err && all.error == first.error)
```

Over a concatenation, `Result.all` is an `Ok` only when it is an `Ok` on both parts:

```ant
let xs: Array[Result[int, string]] = [Result.ok(1)]
let ys: Array[Result[int, string]] = [Result.ok(2), Result.err("e")]
assert(Result.all(xs.concat(ys)) is Ok == (Result.all(xs) is Ok && Result.all(ys) is Ok))
```

## Possible issues

`Array.map` calls its function on every element before `Result.all` sees any result, so the work
after a failure still runs:

<!-- test: before
func parsePort(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}
-->

```ant
let seen: Array[string] = []
let ports = ["1", "x", "3"].map({ seen.push(x); parsePort(x) }) | Result.all()
ports // Err{ error: "not a number: x" }
seen  // ["1", "x", "3"]
```

To stop at the first failure, write a loop that returns at the first error:

```ant
func parseAll(xs: Array[string]): Result[Array[int], string] {
  let out: Array[int] = []
  for s in xs {
    println("parsing ${s}")
    let n = String.toInt(s) ?? return Result.err(s)
    out.push(n)
  }
  Result.ok(out)
}

println(parseAll(["1", "x", "3"]))
```

```text
parsing 1
parsing x
Err{ error: "x" }
```

Only the first error survives. To report every error, keep the `Err` elements with `Array.filter`:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("a"), Result.err("b")]
Result.all(rs)          // Err{ error: "a" }
rs.filter({ x is Err }) // [Err{ error: "a" }, Err{ error: "b" }]
```

`try` binds tighter than `|`. Without parentheses it applies to the mapped array, which is not a
`Result`:

```ant
func parsePrice(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("bad price: ${s}")
  Result.ok(n)
}

func total(prices: Array[string]): Result[int, string] {
  let cents = try prices.map(parsePrice) | Result.all() // error: try needs a Result, got Array[Result[int, string]]
  Result.ok(cents.length())
}
```

Write `try (prices.map(parsePrice) | Result.all())`, as in Applications.

## See also

[Result.map](map.md) · [Result.unwrap](unwrap.md) · [Result.isOk](isOk.md) ·
[Result.wrapErr](wrapErr.md) · [Array.map](../Array/map.md) · [Parallel.map](../Parallel/map.md) ·
[Parallel.any](../Parallel/any.md)
