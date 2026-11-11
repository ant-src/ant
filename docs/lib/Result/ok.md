<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.ok

```ant signature
func Result.ok[T, E](value: T): Result[T, E]
```

Result.ok returns an Ok carrying value; Result.ok(v) is the same value as Ok\{ value: v \}.

<!-- end generated -->

## Details

- The success value of a `Result[void, E]` is `Result.ok(void)`.

## Basic examples

End a function that returns a `Result` with the success value:

```ant
func half(n: int): Result[int, string] {
  if n % 2 != 0 {
    return Result.err("${n} is odd")
  }
  Result.ok(n / 2)
}

half(10) // Ok{ value: 5 }
half(7)  // Err{ error: "7 is odd" }
```

Build a success value directly, with the `Result` type on the `let`:

```ant
let r: Result[string, Error] = Result.ok("done")
r // Ok{ value: "done" }
```

## Scope

The value can be `void`, for an operation that succeeds with nothing to return:

```ant
let saved: Result[void, Error] = Result.ok(void)
saved // Ok{ value: void }
```

`null`, records, arrays, and maps are carried as they are:

```ant
type Point = { x: int, y: int }
let none: Result[int | null, string] = Result.ok(null)
none // Ok{ value: null }
let p: Result[Point, string] = Result.ok(Point{ x: 1, y: 2 })
p // Ok{ value: Point{ x: 1, y: 2 } }
let words: Result[Array[string], string] = Result.ok(["ant", "bee"])
words // Ok{ value: ["ant", "bee"] }
let ages: Result[Map[string, int], string] = Result.ok(Map{ "Ann": 31 })
ages // Ok{ value: Map{ "Ann": 31 } }
```

The error type can be a union of your own; `Result.ok` does not care which:

```ant
type EmptyInput = {}
type BadDigit = { at: int }
type ParseError = EmptyInput | BadDigit
let parsed: Result[int, ParseError] = Result.ok(42)
parsed // Ok{ value: 42 }
```

A `Result` can carry another `Result`:

```ant
let outer: Result[Result[int, string], string] = Result.ok(Result.ok(3))
outer // Ok{ value: Ok{ value: 3 } }
```

Pipe a value into it:

```ant
let piped: Result[int, string] = 5 | Result.ok()
piped // Ok{ value: 5 }
```

Pass it as a function value where the expected function type fixes `E`:

```ant
let all: Array[Result[int, string]] = [1, 2, 3].map(Result.ok)
all // [Ok{ value: 1 }, Ok{ value: 2 }, Ok{ value: 3 }]
```

## Applications

Validate input and return it, or say what is wrong:

```ant
type Signup = { name: string, age: int }

func validate(s: Signup): Result[Signup, Error] {
  if s.name.isEmpty() {
    return Result.err(Error{ message: "name is empty" })
  }
  if s.age < 13 {
    return Result.err(Error{ message: "too young: ${s.age}" })
  }
  Result.ok(s)
}

validate(Signup{ name: "Ann", age: 31 }) // Ok{ value: Signup{ name: "Ann", age: 31 } }
validate(Signup{ name: "", age: 31 })    // Err{ error: Error{ message: "name is empty" } }
```

Wrap the total at the end of a loop that returns early on bad input:

```ant
func sum(texts: Array[string]): Result[int, string] {
  var total = 0
  for t in texts {
    total += t.toInt() ?? return Result.err("not a number: ${t}")
  }
  Result.ok(total)
}

sum(["1", "2", "3"]) // Ok{ value: 6 }
sum(["1", "two"])    // Err{ error: "not a number: two" }
sum([])              // Ok{ value: 0 }
```

Look a key up in a fixed table, and return the hit early:

```ant
func port(name: string): Result[int, string] {
  let ports = Map{ "http": 80, "https": 443 }
  let hit = ports[name]
  if hit != null {
    return Result.ok(hit)
  }
  Result.err("unknown service ${name}")
}

port("https") // Ok{ value: 443 }
port("ftp")   // Err{ error: "unknown service ftp" }
```

## Properties and relations

`Result.ok(v)` is the same value as the literal `Ok{ value: v }`:

```ant
let r: Result[int, string] = Result.ok(7)
assert(r == Ok{ value: 7 })
```

[`Result.isOk`](isOk.md) is true for it, and [`Result.unwrap`](unwrap.md) gives the value back:

```ant
let r: Result[string, Error] = Result.ok("hi")
assert(r.isOk())
assert(r.unwrap() == "hi")
assert(r.unwrapOr("fallback") == "hi")
```

[`Result.map`](map.md) of it is `Result.ok` of the mapped value:

```ant
let r: Result[int, string] = Result.ok(4)
let doubled: Result[int, string] = Result.ok(8)
assert(r.map({ x * 2 }) == doubled)
```

`try` on it gives the value:

```ant
func check(): Result[void, string] {
  let r: Result[int, string] = Result.ok(9)
  assert(try r == 9)
  Result.ok(void)
}

check() // Ok{ value: void }
```

## Possible issues

Without an expected type, nothing fixes the error type, and the call is a compile error. Annotate
the binding:

```ant
let r = Result.ok(1) // error: cannot infer the type parameter E of Result.ok; annotate the result
```

A parameter that fixes the error type, or the left side of `??`, is enough:

```ant
Result.wrapErr(Result.ok(1), "load") // Ok{ value: 1 }
let cached: Result[int, Error] | null = null
cached ?? Result.ok(0) // Ok{ value: 0 }
```

Type arguments are invariant, so a `Result[int, string]` does not widen to
`Result[int | string, string]`. Build it at the wider type instead:

```ant
let r: Result[int, string] = Result.ok(1)
let wide: Result[int | string, string] = r
// error: cannot use Result[int, string] where Result[int | string, string] is expected
let built: Result[int | string, string] = Result.ok(1)
```

An `if` whose branches are `Result.ok` and `Result.err` fixes neither type on its own: each branch
leaves the other type parameter open:

```ant
func pick(c: bool) {
  let r = if c {
    Result.ok(1) // error: cannot infer the type parameter E of Result.ok; annotate the result
  } else {
    Result.err("x") // error: cannot infer the type parameter T of Result.err; annotate the result
  }
}
```

Annotate the `if`; it passes the expected type to both branches:

```ant
func pick(c: bool): Result[int, string] {
  let r: Result[int, string] = if c { Result.ok(1) } else { Result.err("x") }
  r
}

pick(true)  // Ok{ value: 1 }
pick(false) // Err{ error: "x" }
```

## See also

[Result.err](err.md) · [Result.isOk](isOk.md) · [Result.unwrap](unwrap.md) · [Result.map](map.md) ·
[Result.all](all.md) · [Result](index.md)
