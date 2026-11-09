<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.unwrap

```ant signature
func Result.unwrap[T, E](r: Result[T, E]): T
```

Result.unwrap returns the value of r. It panics if r is an Err, with a message that shows r's error.

<!-- end generated -->

## Details

- Use it where an `Err` is a bug, such as parsing a constant. For a failure the caller expects,
  return it with [`try`](../../language/keywords/try.md) or give a fallback with
  [Result.unwrapOr](unwrapOr.md).
- In a script, both `try` and `Result.unwrap` end the script on an `Err` with exit status 1: `try`
  reports the error alone, and `Result.unwrap` reports a panic with its stack trace.

## Basic examples

Take the value out of a parsed document:

```ant
Result.unwrap(Json.parse("[1, 2]")) // [1.0, 2.0]
```

Call it through `.` or in a pipe:

```ant
Regex.compile("[0-9]+").unwrap()     // /[0-9]+/
Json.parse("true") | Result.unwrap() // true
```

An `Err` panics, and the message shows the error:

```ant
Result.unwrap(Json.parse("[1,"))
// panic: Result.unwrap on an Err: Error{ message: "unexpected end of JSON input" }
```

## Scope

The error type can be anything: a string, a record of your own, or a union:

```ant
type NotFound = { key: string }
let missing: Result[int, NotFound] = Result.err(NotFound{ key: "port" })
Result.unwrap(missing) // panic: Result.unwrap on an Err: NotFound{ key: "port" }
let bad: Result[int, string] = Result.err("bad input")
Result.unwrap(bad) // panic: Result.unwrap on an Err: "bad input"
```

A `Result[void, Error]`, such as the one [File.write](../File/write.md) returns, unwraps the same
way. On an `Ok` the call returns `void`, which carries no information, so there is nothing to use:

```ant
File.write("out.txt", "hi").unwrap()
println(File.read("out.txt").unwrap())
```

```text
hi
```

An `Ok` may hold `null`, which `Result.unwrap` returns like any value:

```ant
let found: Result[string | null, Error] = Result.ok(null)
Result.unwrap(found) // null
```

`Result.unwrap` reads only the outer tag. On an `Ok` that holds a `Result`, it returns the inner
`Result`, which a second call unwraps:

```ant
let inner: Result[int, Error] = Result.ok(7)
let outer: Result[Result[int, Error], Error] = Result.ok(inner)
Result.unwrap(outer)    // Ok{ value: 7 }
outer.unwrap().unwrap() // 7
```

The value comes back as it is, not copied: an array taken out of a `Result` is the same array:

```ant
let r: Result[Array[int], Error] = Result.ok([1, 2])
let xs = r.unwrap()
xs.push(3)
r // Ok{ value: [1, 2, 3] }
```

## Applications

Compile a regular expression written in the source once, at the top of a script. A bad pattern is a
bug in the script, so a panic fits:

```ant
let digits = Regex.compile("[0-9]+").unwrap()
for line in ["order 66", "no number", "room 101"] {
  println(Regex.find(digits, line)?.text ?? "-")
}
```

```text
66
-
101
```

Read a configuration written into the program, where a parse error is a mistake of the programmer:

```ant
let defaults = Json.parse("{\"port\": 8080, \"debug\": false}").unwrap()
println(defaults)
```

```text
Map{ "port": 8080.0, "debug": false }
```

Parse a fixed list of dates with [Result.all](all.md), and unwrap once for the whole list:

```ant
let holidays = ["2026-01-01", "2026-12-25"]
  | Array.map(Time.parsePlainDate)
  | Result.all()
  | Result.unwrap()
println(holidays.length())
```

```text
2
```

## Properties and relations

On an `Ok`, `Result.unwrap` returns what [Result.unwrapOr](unwrapOr.md) returns, whatever the
fallback:

```ant
let r: Result[int, Error] = Result.ok(5)
assert(Result.unwrap(r) == Result.unwrapOr(r, 0))
assert(Result.unwrap(r) == Result.unwrapOr(r, -1))
```

Unwrapping after [Result.map](map.md) is the same as applying the function to the unwrapped value:

```ant
let r: Result[string, Error] = Result.ok("ant")
let f = func(s: string): int { s.length() }
assert(Result.unwrap(Result.map(r, f)) == f(Result.unwrap(r)))
```

[Result.isOk](isOk.md) tells in advance whether `Result.unwrap` panics, which
[Panic.catch](../Panic/catch.md) confirms:

```ant
let r: Result[int, string] = Result.err("no")
assert(!r.isOk())
assert(!Panic.catch({ Result.unwrap(r) }).isOk())
```

## Possible issues

An `Err` from input the program does not control, such as a file or a user, is not a bug. With
`Result.unwrap` a missing file ends the script with a panic and a stack trace:

<!-- test: exit 1 -->

```ant
let text = Result.unwrap(File.read("settings.json"))
println(text)
```

Return the error with [`try`](../../language/keywords/try.md) instead, so the script ends with the
message alone:

<!-- test: exit 1 -->

```ant
let text = try File.read("settings.json")
println(text)
```

Or give a fallback with [Result.unwrapOr](unwrapOr.md):

```ant
let text = Result.unwrapOr(File.read("settings.json"), "{}")
println(text)
```

```text
{}
```

`Result.unwrap` takes a `Result`, not a `Result` that may be `null`. Handle the `null` first, with
`??` and a fallback `Result`:

```ant
let cached: Result[int, Error] | null = null
let n = Result.unwrap(cached)
// error: cannot use Result[int, Error] | null where Result[int, Error] is expected
```

```ant
let cached: Result[int, Error] | null = null
let fallback: Result[int, Error] = Result.ok(0)
Result.unwrap(cached ?? fallback) // 0
```

## See also

[Result.unwrapOr](unwrapOr.md) · [Result.isOk](isOk.md) · [Result.map](map.md) ·
[Result.all](all.md) · [try](../../language/keywords/try.md) · [Panic.catch](../Panic/catch.md) ·
[Errors guide](../../guide/errors.md)
