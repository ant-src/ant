<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Result.wrapErr

```ant signature
func Result.wrapErr[T](r: Result[T, Error], context: string): Result[T, Error]
```

Result.wrapErr returns r itself if it is an Ok, and an Err holding Error.wrap(r.error, context) if
it is an Err.

<!-- end generated -->

## Details

- The new message is `context`, `: `, and the old message. An empty `context` is joined like any
  other, so the message then starts with `: `.
- An `Err` comes back as a new `Err` holding a new `Error`; `r` and its error are untouched.

## Basic examples

Put the operation in front of the message of an error:

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "no such file" })
Result.wrapErr(failed, "load config") // Err{ error: Error{ message: "load config: no such file" } }
```

An `Ok` passes through:

```ant
let loaded: Result[int, Error] = Result.ok(8080)
Result.wrapErr(loaded, "load config") // Ok{ value: 8080 }
```

Wrap the result of a call in a pipe:

```ant
Json.parse("{") | Result.wrapErr("parse app.json")
// Err{ error: Error{ message: "parse app.json: unexpected end of JSON input" } }
```

## Scope

Each wrap adds one more context in front, so the outermost caller reads first:

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "timeout" })
failed | Result.wrapErr("fetch user") | Result.wrapErr("render page")
// Err{ error: Error{ message: "render page: fetch user: timeout" } }
```

An empty context still adds the separator:

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "timeout" })
Result.wrapErr(failed, "") // Err{ error: Error{ message: ": timeout" } }
```

Call it through `.`:

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "refused" })
failed.wrapErr("connect") // Err{ error: Error{ message: "connect: refused" } }
```

## Named arguments

Name the arguments to pass them in any order:

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "refused" })
Result.wrapErr(context: "connect", r: failed) // Err{ error: Error{ message: "connect: refused" } }
```

## Applications

Add context at each level and propagate with `try`. Parenthesize the call together with its
`wrapErr` stage, so that `try` unwraps the wrapped result:

<!-- test: file app.json
{"port":
-->

```ant
func loadConfig(path: string): Result[Json.Value, Error] {
  let text = try File.read(path)
  let json = try (Json.parse(text) | Result.wrapErr("parse ${path}"))
  Result.ok(json)
}

let config = loadConfig("app.json") | Result.wrapErr("load config")
if config is Err {
  println(config.error.message)
}
```

```text
load config: parse app.json: unexpected end of JSON input
```

Name the input that failed when you process several, then keep the first error with
[`Result.all`](all.md):

```ant
type Source = { name: string, text: string }
let sources = [Source{ name: "a.json", text: "[1]" }, Source{ name: "b.json", text: "[1," }]
sources | Array.map({ Json.parse(x.text) | Result.wrapErr("parse ${x.name}") }) | Result.all()
// Err{ error: Error{ message: "parse b.json: unexpected end of JSON input" } }
```

Turn an error of your own type into an `Error` for a person to read, then add the context:

```ant
type NotFound = { id: int }
let missing: Result[string, NotFound] = Result.err(NotFound{ id: 7 })
missing | Result.mapErr({ Error{ message: "user ${x.id} not found" } }) | Result.wrapErr("load profile")
// Err{ error: Error{ message: "load profile: user 7 not found" } }
```

## Properties and relations

`Result.wrapErr` is [`Result.mapErr`](mapErr.md) with [`Error.wrap`](../Error/wrap.md):

```ant
let failed: Result[int, Error] = Result.err(Error{ message: "timeout" })
assert(Result.wrapErr(failed, "fetch") == Result.mapErr(failed, { Error.wrap(x, "fetch") }))
```

On an `Ok`, the result is `r` itself:

```ant
let loaded: Result[int, Error] = Result.ok(1)
assert(Record.same(Result.wrapErr(loaded, "load"), loaded))
```

Wrapping leaves the original error as it was:

```ant
let cause = Error{ message: "timeout" }
let failed: Result[int, Error] = Result.err(cause)
let wrapped = Result.wrapErr(failed, "fetch")
cause // Error{ message: "timeout" }
```

## Possible issues

The error type must be `Error`. A `Result` with an error type of your own is a compile error; use
[`Result.mapErr`](mapErr.md) to make an `Error` first, as in Applications:

```ant
type NotFound = { id: int }
let missing: Result[string, NotFound] = Result.err(NotFound{ id: 7 })
let wrapped = Result.wrapErr(missing, "load profile")
// error: cannot use Result[string, NotFound] where Result[string, Error] is expected
```

`try` binds tighter than `|`, so without parentheses it unwraps the call and pipes the value into
`wrapErr`:

```ant
func parse(text: string): Result[Json.Value, Error] {
  let json = try Json.parse(text) | Result.wrapErr("parse")
  // error: cannot use Json.Value where Result[T, Error] is expected
  Result.ok(json)
}
```

The wrapped error is a new record that holds only text. `Record.same` with the original is false,
and nothing leads back from it to the original error. Branch on an error type of your own instead:

```ant
let cause = Error{ message: "timeout" }
let failed: Result[int, Error] = Result.err(cause)
let wrapped = Result.wrapErr(failed, "fetch")
if wrapped is Err { Record.same(wrapped.error, cause) } else { false } // false
```

## See also

[Error.wrap](../Error/wrap.md), [Result.mapErr](mapErr.md), [Result.map](map.md),
[Result.all](all.md), [Result.unwrap](unwrap.md), [try](../../language/keywords/try.md)
