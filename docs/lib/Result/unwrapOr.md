<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.unwrapOr

```ant signature
func Result.unwrapOr[T, E](r: Result[T, E], fallback: T): T
```

Result.unwrapOr returns the value of r, or fallback if r is an Err.

<!-- end generated -->

## Details

- The fallback is an ordinary argument, so it is evaluated before the call, even when `r` is an
  `Ok`.
- The error of an `Err` is dropped. When it matters, use [try](../../language/keywords/try.md) or
  test the result with `is`.

## Basic examples

```ant
let good: Result[int, string] = Result.ok(42)
let bad: Result[int, string] = Result.err("not a number")
Result.unwrapOr(good, 0) // 42
Result.unwrapOr(bad, 0)  // 0
```

Call it through `.` on any `Result`:

```ant
Url.unescape("caf%C3%A9").unwrapOr("?") // "café"
Url.unescape("100%").unwrapOr("?")      // "?"
```

## Scope

The value type may be any type, such as `Time.Zone`:

```ant
Time.zone("Europe/Paris").unwrapOr(Time.utc) // Europe/Paris
Time.zone("Mars/Olympus").unwrapOr(Time.utc) // UTC
```

A `bool` result with `false` as the fallback treats an error as "no":

```ant
Path.match("*.md", "README.md").unwrapOr(false) // true
Path.match("[", "README.md").unwrapOr(false)    // false
```

An empty collection as the fallback takes the value type, so it needs no annotation:

```ant
Url.parseQuery("a=1&b=2").unwrapOr(Map{}) // Map{ "a": ["1"], "b": ["2"] }
Url.parseQuery("a=%zz").unwrapOr(Map{})   // Map{}
```

The error type may be any type, such as a union you declare:

```ant
type Empty = {}
type TooLong = { length: int }
type NameError = Empty | TooLong

func checkName(s: string): Result[string, NameError] {
  if s.isEmpty() {
    return Result.err(Empty{})
  }
  if s.length() > 8 {
    return Result.err(TooLong{ length: s.length() })
  }
  Result.ok(s)
}

checkName("ann").unwrapOr("guest")         // "ann"
checkName("").unwrapOr("guest")            // "guest"
checkName("bartholomew").unwrapOr("guest") // "guest"
```

When the value type includes `null`, `null` is a valid fallback:

```ant
Json.parse("[1, 2]").unwrapOr(null) // [1.0, 2.0]
Json.parse("[1,").unwrapOr(null)    // null
```

The fallback is a named parameter:

```ant
let r: Result[int, string] = Result.err("timeout")
r.unwrapOr(fallback: -1) // -1
```

In a pipe, the `Result` is the first argument:

```ant
"a%20b" | Url.unescape() | Result.unwrapOr("") // "a b"
```

## Applications

Read a setting, and use a default when it is malformed:

```ant
func parsePort(s: string): Result[int, string] {
  let n = s.toInt() ?? return Result.err("not a number: ${s}")
  if n < 1 || n > 65535 {
    return Result.err("out of range: ${n}")
  }
  Result.ok(n)
}

for input in ["8080", "http", "70000"] {
  println("${input} -> ${parsePort(input).unwrapOr(80)}")
}
```

```text
8080 -> 8080
http -> 80
70000 -> 80
```

Turn a panic into a default with [Panic.catch](../Panic/catch.md):

```ant
let xs = [10, 20, 30]
Panic.catch({ xs[1] }).unwrapOr(0) // 20
Panic.catch({ xs[5] }).unwrapOr(0) // 0
```

Map every input to a value, with the fallback standing in for the ones that fail:

```ant
["PT1H", "later", "PT15M"].map({ Time.parseDuration(x).unwrapOr(Time.duration()) })
// [PT1H, PT0S, PT15M]
```

## Properties and relations

`Result.unwrapOr` is the same as testing the tag with `is`:

```ant
let cases: Array[Result[string, int]] = [Result.ok("a"), Result.err(1)]
for r in cases {
  let viaIs = if r is Ok { r.value } else { "z" }
  assert(r.unwrapOr("z") == viaIs)
}
```

On an `Ok` it agrees with [Result.unwrap](unwrap.md):

```ant
let ok: Result[int, string] = Result.ok(3)
assert(ok.unwrapOr(0) == ok.unwrap())
```

## Possible issues

The fallback must have the value type. A fallback of another type is a compile error:

```ant
let text = Url.unescape("%zz").unwrapOr(null) // error: cannot use null where string is expected
```

The fallback is evaluated before the call, even when the result is an `Ok`. An expensive or
side-effecting fallback runs every time:

```ant
func defaultName(): string {
  println("computing the default")
  "guest"
}

let r: Result[string, string] = Result.ok("ann")
println(r.unwrapOr(defaultName()))
```

```text
computing the default
ann
```

Test the tag with `is` to compute the fallback only for an `Err`:

```ant
func defaultName(): string {
  println("computing the default")
  "guest"
}

let r: Result[string, string] = Result.ok("ann")
println(if r is Ok { r.value } else { defaultName() })
```

```text
ann
```

`Result.err` or `Result.ok` written inline as the first argument fixes only one of its type
parameters, and nothing fixes the other, so the call is a compile error:

```ant
Result.unwrapOr(Result.err("e"), 0)
// error: cannot infer the type parameter T of Result.err; annotate the result
```

Annotate the `Result` first:

```ant
let r: Result[int, string] = Result.err("e")
r.unwrapOr(0) // 0
```

When the value type includes `null` and the fallback is `null`, a `null` result does not tell an
`Ok` holding `null` from an `Err`:

```ant
let found: Result[int | null, string] = Result.ok(null)
let failed: Result[int | null, string] = Result.err("lookup failed")
found.unwrapOr(null)  // null
failed.unwrapOr(null) // null
```

On a nested `Result`, an inner `Err` is returned as a value, not replaced by the fallback:

```ant
let nested: Result[Result[int, string], string] = Result.ok(Result.err("inner"))
nested.unwrapOr(Result.ok(0))              // Err{ error: "inner" }
nested.unwrapOr(Result.ok(0)).unwrapOr(-1) // -1
```

## See also

[Result.unwrap](unwrap.md) · [Result.isOk](isOk.md) · [Result.map](map.md) · [Result.all](all.md) ·
[Panic.catch](../Panic/catch.md) · [try](../../language/keywords/try.md) ·
[?? operator](../../language/operators/null-coalescing.md) · [Errors guide](../../guide/errors.md)
