<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.isOk

```ant signature
func Result.isOk[T, E](r: Result[T, E]): bool
```

Result.isOk reports whether r is an Ok.

<!-- end generated -->

## Details

- `Result.isOk` does not narrow `r`. To read `r.value` or `r.error`, test `r is Ok` or `r is Err`
  instead.
- There is no `Result.isErr`; write `!r.isOk()`.

## Basic examples

```ant
let good: Result[int, string] = Result.ok(1)
let bad: Result[int, string] = Result.err("no")
Result.isOk(good) // true
Result.isOk(bad)  // false
```

Call it through `.` on the result of a call that can fail:

```ant
Json.parse("[1, 2]").isOk() // true
Json.parse("[1, 2").isOk()  // false
```

## Scope

A record literal of `Ok` or `Err` is a `Result` once its type is known:

```ant
let hit: Result[int, string] = Ok{ value: 3 }
let miss: Result[int, string] = Err{ error: "not found" }
hit.isOk()  // true
miss.isOk() // false
```

The value inside does not matter, so an `Ok` of `void` or of `null` is still an `Ok`:

```ant
let saved: Result[void, Error] = Result.ok(void)
let found: Result[string | null, Error] = Ok{ value: null }
saved.isOk() // true
found.isOk() // true
```

An `Ok` whose value is itself an `Err` is an `Ok`. Ask the inner result separately:

```ant
let outer: Result[Result[int, string], string] = Result.ok(Result.err("inner"))
outer.isOk()                          // true
outer.isOk() && outer.unwrap().isOk() // false
```

The error type is whatever the code chooses, such as a union of records:

```ant
type Empty = {}
type TooLong = { limit: int }
func checkName(s: string): Result[string, Empty | TooLong] {
  if s.isEmpty() {
    return Result.err(Empty{})
  }
  if s.length() > 8 {
    return Result.err(TooLong{ limit: 8 })
  }
  Result.ok(s)
}
checkName("ann").isOk()         // true
checkName("").isOk()            // false
checkName("bartholomew").isOk() // false
```

A panic caught by `Panic.catch` is an `Err`:

```ant
Panic.catch({ [1, 2][0] }).isOk() // true
Panic.catch({ [1, 2][5] }).isOk() // false
```

In a pipe, the result is the first argument:

```ant
"90s" | Time.parseDuration() | Result.isOk()  // true
"soon" | Time.parseDuration() | Result.isOk() // false
```

## Applications

Keep the inputs that parse, and count the rest:

```ant
let inputs = ["PT1H", "90s", "P2D", "soon", "5"]
let valid = inputs.filter({ Time.parseDuration(x).isOk() })
println(valid)
println("${inputs.length() - valid.length()} rejected")
```

```text
["PT1H", "90s", "P2D"]
2 rejected
```

Count the successes in a batch by passing `Result.isOk` by name, and the failures with a lambda:

```ant
let docs = ["{}", "[1]", "{", "nul"]
let results = docs.map({ Json.parse(x) })
println("${results.countWhere(Result.isOk)} parsed")
println("${results.countWhere({ !x.isOk() })} failed")
```

```text
2 parsed
2 failed
```

Check that a pattern compiles before using it, without needing the compiled pattern:

```ant
func validPattern(source: string): bool {
  Regex.compile(source).isOk()
}

println(validPattern("[a-z]+"))
println(validPattern("[a-z"))
```

```text
true
false
```

## Properties and relations

`r.isOk()` is the same as `r is Ok`, and the opposite of `r is Err`:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("no")]
for r in rs {
  assert(r.isOk() == (r is Ok))
  assert(r.isOk() == !(r is Err))
}
```

`Result.map` and `Result.mapErr` keep the tag, and `Result.all` is an `Ok` exactly when every
element is:

```ant
let rs: Array[Result[int, string]] = [Result.ok(1), Result.err("no"), Result.ok(3)]
for r in rs {
  assert(r.map({ x * 2 }).isOk() == r.isOk())
  assert(r.mapErr({ x.length() }).isOk() == r.isOk())
}
assert(Result.all(rs).isOk() == rs.every({ x.isOk() }))
assert(Result.all([rs[0], rs[2]]).isOk())
```

## Possible issues

`Result.isOk` returns a `bool` and does not narrow, so reading `value` after it is a compile error:

```ant
let r = Json.parse("[1, 2]")
if r.isOk() {
  println(r.value)
  // error: Result[Json.Value, Error] has no field value on every member: narrow the value first
}
```

Test the tag with `is`, which narrows `r` inside the branch:

```ant
let r = Json.parse("[1, 2]")
if r is Ok {
  println(r.value)
}
```

```text
[1.0, 2.0]
```

`Result.ok(1)` alone does not fix the error type, so passing it straight to `Result.isOk` is a
compile error. Annotate a local first:

```ant
println(Result.isOk(Result.ok(1))) // error: cannot infer the type parameter E of Result.ok; annotate the result
```

A bare `Ok` or `Err` literal fixes only one type parameter, so calling `Result.isOk` on it is a
compile error. Give it a `Result` type first, as in Scope:

```ant
Ok{ value: 3 }.isOk()    // error: cannot infer the type parameter E of Result.isOk
Err{ error: "x" }.isOk() // error: cannot infer the type parameter T of Result.isOk
```

## See also

[Result](index.md) · [Result.unwrap](unwrap.md) · [Result.unwrapOr](unwrapOr.md) ·
[Result.all](all.md) · [Result.map](map.md) · [Result.ok](ok.md) · [Result.err](err.md)
