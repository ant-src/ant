<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.err

```ant signature
func Result.err[T, E](error: E): Result[T, E]
```

Result.err returns an Err carrying error; Result.err(e) is the same value as Err\{ error: e \}.

<!-- end generated -->

## Details

- Nothing in the argument fixes the value type `T`, so the call needs an expected type, such as a
  declared return type, an annotated `let`, or a parameter type.

## Basic examples

Build a failed result for an annotated name:

```ant
let r: Result[int, string] = Result.err("no input")
r // Err{ error: "no input" }
```

Return a failure from a function. The declared return type fixes both type parameters:

```ant
func half(n: int): Result[int, string] {
  if n % 2 != 0 {
    return Result.err("${n} is odd")
  }
  Result.ok(n / 2)
}

half(8) // Ok{ value: 4 }
half(3) // Err{ error: "3 is odd" }
```

Test the result with `is` and read its `error` field:

```ant
let r: Result[int, string] = Result.err("timeout")
if r is Err { r.error } // "timeout"
```

## Scope

The error can be an `Error`, the record the standard library uses. [Result.fail](fail.md) builds the
same result from the message alone:

```ant
let r: Result[string, Error] = Result.err(Error{ message: "no such file" })
r // Err{ error: Error{ message: "no such file" } }
let short: Result[string, Error] = Result.fail("no such file")
assert(short == r)
```

The error can be a member of a union of your own error records:

```ant
type NotFound = { id: int }
type Denied = { user: string }
type LookupError = NotFound | Denied

func find(id: int): Result[string, LookupError] {
  if id < 0 {
    return Result.err(Denied{ user: "guest" })
  }
  if id > 3 {
    return Result.err(NotFound{ id: id })
  }
  Result.ok("item ${id}")
}

find(9)  // Err{ error: NotFound{ id: 9 } }
find(-1) // Err{ error: Denied{ user: "guest" } }
find(1)  // Ok{ value: "item 1" }
```

An operation that has no value to return succeeds with `void` and can still fail:

```ant
let saved: Result[void, string] = Result.err("disk full")
saved // Err{ error: "disk full" }
```

Any value can be the error, even an empty string or `null`:

```ant
let blank: Result[int, string] = Result.err("")
blank // Err{ error: "" }
let missing: Result[int, string | null] = Result.err(null)
missing // Err{ error: null }
```

Elements of an annotated array take their type from it:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("a"), Result.err("b")]
Result.all(rs) // Err{ error: "a" }
```

Call it in a pipe:

```ant
let r: Result[int, string] = "bad" | Result.err()
r // Err{ error: "bad" }
```

## Applications

Return early with `?? return` when a lookup comes back `null`:

```ant
let stock = Map{ "nut": 4, "bolt": 0 }

func take(name: string): Result[int, string] {
  let count = stock[name] ?? return Result.err("unknown item ${name}")
  if count == 0 {
    return Result.err("${name} is out of stock")
  }
  Result.ok(count - 1)
}

println(take("nut"))
println(take("bolt"))
println(take("screw"))
```

```text
Ok{ value: 3 }
Err{ error: "bolt is out of stock" }
Err{ error: "unknown item screw" }
```

An `Err` from `Result.err` stops a chain of `try` at the first step that fails:

```ant
func half(n: int): Result[int, string] {
  if n % 2 != 0 {
    return Result.err("${n} is odd")
  }
  Result.ok(n / 2)
}

func quarter(n: int): Result[int, string] {
  let h = try half(n)
  half(h)
}

println(quarter(8))
println(quarter(6))
```

```text
Ok{ value: 2 }
Err{ error: "3 is odd" }
```

A script that returns an `Err` ends with exit status 1 and writes the error to standard error:

<!-- test: exit 1 -->

```ant
let args = ["--verbose"]
if !args.contains("--input") {
  return Result.err(Error{ message: "missing --input" })
}
println("running")
```

## Properties and relations

`Result.err(e)` is the same value as the literal `Err{ error: e }`:

```ant
let r: Result[int, string] = Result.err("x")
assert(r == Err{ error: "x" })
```

The other `Result` functions pass an `Err` through or fall back:

```ant
let r: Result[int, string] = Result.err("x")
assert(!r.isOk())
assert(r.map({ x + 1 }) == r)
assert(r.unwrapOr(0) == 0)
let upper: Result[int, string] = Result.err("X")
assert(r.mapErr({ x.toUpper() }) == upper)
```

Two calls build two records, equal but not the same:

```ant
let a: Result[int, string] = Result.err("x")
let b: Result[int, string] = Result.err("x")
assert(a == b)
assert(!Record.same(a, b))
```

## Possible issues

Without an expected type, nothing fixes `T`, and the call is a compile error:

```ant
let r = Result.err("no input")
// error: cannot infer the type parameter T of Result.err; annotate the result
```

`Result.unwrap` on an `Err` panics:

```ant
let r: Result[int, string] = Result.err("x")
r.unwrap() // panic: Result.unwrap on an Err: "x"
```

## See also

[Result.ok](ok.md) · [Result.fail](fail.md) · [Result.isOk](isOk.md) · [Result.mapErr](mapErr.md) ·
[Result.wrapErr](wrapErr.md) · [Result.unwrapOr](unwrapOr.md) · [Result](index.md) ·
[Error.wrap](../Error/wrap.md) · [try](../../language/keywords/try.md)
