<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.map

```ant signature
func Result.map[T, U, E](r: Result[T, E], f: func(T): U): Result[U, E]
```

Result.map returns r with f applied to its value if r is an Ok, and r's error unchanged if it is an Err, without calling f.

<!-- end generated -->

## Details

- On an `Err`, the error and its type `E` stay as they are. Use [`Result.mapErr`](mapErr.md) to
  change the error.

## Basic examples

Transform the value of an `Ok`:

```ant
let r: Result[int, string] = Result.ok(2)
Result.map(r, { x * 10 }) // Ok{ value: 20 }
```

An `Err` passes through unchanged:

```ant
let r: Result[int, string] = Result.err("bad")
Result.map(r, { x * 10 }) // Err{ error: "bad" }
```

Map the result of a call that can fail, in a pipe:

```ant
Url.parse("https://ant-lang.org/docs") | Result.map({ x.host }) // Ok{ value: "ant-lang.org" }
```

## Scope

The value type may change while the error type stays:

```ant
let r: Result[int, string] = Result.ok(7)
r.map({ "#${x}" }) // Ok{ value: "#7" }
```

Pass a named function instead of a lambda:

```ant
func double(n: int): int {
  n * 2
}
let r: Result[int, string] = Result.ok(4)
r.map(double) // Ok{ value: 8 }
```

`f` may return `null`, which becomes the value of the `Ok`:

```ant
let s: Result[string, string] = Result.ok("12x")
s.map(String.toInt) // Ok{ value: null }
```

An error of a union type passes through with its member intact:

```ant
type NotFound = { id: int }
type Denied = { user: string }
let r: Result[int, NotFound | Denied] = Result.err(Denied{ user: "bob" })
r.map({ x + 1 }) // Err{ error: Denied{ user: "bob" } }
```

A `Result[void, E]` maps to a value of any type:

```ant
let saved: Result[void, string] = Result.ok(void)
saved.map({ "saved" }) // Ok{ value: "saved" }
```

It works on the `Result` of [`Panic.catch`](../Panic/catch.md) like on any other:

```ant
Panic.catch({ 10 / 2 }) | Result.map({ x * 3 }) // Ok{ value: 15 }
```

## Applications

Build a record from a parsed value, keeping the parse error:

```ant
type Event = { name: string, date: Time.PlainDate }

func parseEvent(name: string, date: string): Result[Event, Error] {
  Time.parsePlainDate(date) | Result.map({ Event{ name, date: x } })
}

println(parseEvent("launch", "2026-11-09"))
println(parseEvent("launch", "soon"))
```

```text
Ok{ value: Event{ name: "launch", date: 2026-11-09 } }
Err{ error: Error{ message: "not a date or time: \"soon\"" } }
```

Count the values of [`Result.all`](all.md), which stops at the first error:

```ant
func countDates(dates: Array[string]): Result[int, Error] {
  dates | Array.map(Time.parsePlainDate) | Result.all() | Result.map({ len(x) })
}

println(countDates(["2024-02-29", "2025-01-01"]))
println(countDates(["2024-02-29", "2023-02-29"]))
```

```text
Ok{ value: 2 }
Err{ error: Error{ message: "no such date: 2023-02-29" } }
```

Inside a function that returns a `Result`, map a pipeline and `try` it as a whole:

```ant
func hostOf(s: string): Result[string, Error] {
  let host = try (Url.parse(s) | Result.map({ x.host }))
  Result.ok(String.toUpper(host))
}

println(hostOf("https://ant-lang.org/docs"))
println(hostOf("http://[::1"))
```

```text
Ok{ value: "ANT-LANG.ORG" }
Err{ error: Error{ message: "parse \"http://[::1\": missing ']' in host" } }
```

## Properties and relations

Mapping with the identity function gives an equal `Result`:

```ant
let r: Result[int, string] = Result.ok(4)
let e: Result[int, string] = Result.err("bad")
assert(r.map({ x }) == r)
assert(e.map({ x }) == e)
```

Two maps in a row are one map with the two functions composed:

```ant
let f = func(x: int): int { x + 1 }
let g = func(x: int): int { x * 2 }
let r: Result[int, string] = Result.ok(4)
assert(r.map(f).map(g) == r.map({ g(f(x)) }))
```

It never turns an `Ok` into an `Err` or back:

```ant
let r: Result[int, string] = Result.ok(4)
let e: Result[int, string] = Result.err("bad")
assert(r.map({ x * 2 }).isOk() == r.isOk())
assert(e.map({ x * 2 }).isOk() == e.isOk())
```

In a function that returns a `Result`, `Result.map(r, f)` is the same as `Result.ok(f(try r))`:

```ant
func viaMap(r: Result[int, string]): Result[int, string] {
  Result.map(r, { x + 1 })
}

func viaTry(r: Result[int, string]): Result[int, string] {
  Result.ok(try r + 1)
}

let r: Result[int, string] = Result.ok(4)
let e: Result[int, string] = Result.err("bad")
assert(viaMap(r) == viaTry(r))
assert(viaMap(e) == viaTry(e))
```

## Possible issues

A function that itself returns a `Result` gives a nested `Result`:

```ant
func parseAge(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}
let input: Result[string, string] = Result.ok("41")
input.map(parseAge) // Ok{ value: Ok{ value: 41 } }
```

Annotating the result as a flat `Result` is a compile error:

```ant
func parseAge(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}
let input: Result[string, string] = Result.ok("41")
let age: Result[int, string] = input.map(parseAge)
// error: cannot use func(string): Result[int, string] where func(string): int is expected
```

Unwrap with `try` and call the function directly instead:

```ant
func parseAge(s: string): Result[int, string] {
  let n = String.toInt(s) ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}

func ageOf(input: Result[string, string]): Result[int, string] {
  parseAge(try input)
}

let input: Result[string, string] = Result.ok("41")
println(ageOf(input))
```

```text
Ok{ value: 41 }
```

`try` binds tighter than `|`, so without parentheses it unwraps the first stage and pipes the bare
value into `Result.map`:

```ant
func hostOf(s: string): Result[string, Error] {
  try Url.parse(s) | Result.map({ x.host })
  // error: cannot use Url.Parts where Result[T, Error] is expected
}
```

The result is still a `Result`, so reading its value needs narrowing first:

```ant
let r: Result[int, string] = Result.ok(2)
let n = Result.map(r, { x * 10 })
println(n.value)
// error: Result[int, string] has no field value on every member: narrow the value first
```

A bare `Ok` literal or `Result.ok(...)` without an expected type leaves `E` unknown:

```ant
println(Result.map(Ok{ value: 3 }, { x + 1 }))
// error: cannot infer the type parameter E of Result.map; annotate the result
```

```ant
println(Result.map(Result.ok(3), { x + 1 }))
// error: cannot infer the type parameter E of Result.ok; annotate the result
```

## See also

[Result.mapErr](mapErr.md) · [Result.wrapErr](wrapErr.md) · [Result.all](all.md) ·
[Result.unwrapOr](unwrapOr.md) · [Result.isOk](isOk.md) · [Array.map](../Array/map.md) ·
[Iter.map](../Iter/map.md) · [try](../../language/keywords/try.md) ·
[| operator](../../language/operators/pipe.md)
