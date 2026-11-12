<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.wrap

```ant signature
func Error.wrap(e: Error, context: string): Error
```

Error.wrap returns a new Error whose message is context, ": ", and the message of e, leaving e
untouched. An empty context is joined like any other.

<!-- end generated -->

## Details

- The result keeps only the joined text; nothing leads from it back to `e`.

## Basic examples

Put the operation in front of the message:

```ant
let e = Error{ message: "no such file" }
Error.wrap(e, "load config") // Error{ message: "load config: no such file" }
```

Read the joined message:

```ant
let e = Error{ message: "connection refused" }
Error.wrap(e, "connect db.local:5432").message // "connect db.local:5432: connection refused"
```

## Scope

Each wrap adds one more context in front, so the outermost caller reads first:

```ant
let e = Error{ message: "timeout" }
e | Error.wrap("fetch user") | Error.wrap("render page") // Error{ message: "render page: fetch user: timeout" }
```

An empty context still adds the separator:

```ant
Error.wrap(Error{ message: "timeout" }, "") // Error{ message: ": timeout" }
```

An empty message keeps the separator too:

```ant
Error.wrap(Error{ message: "" }, "fetch") // Error{ message: "fetch: " }
```

Call it through `.`:

```ant
let e = Error{ message: "refused" }
e.wrap("connect") // Error{ message: "connect: refused" }
```

## Applications

Turn an error type of your own into an `Error` for a person to read, inside
[`Result.mapErr`](../Result/mapErr.md). For a `Result` that already holds an `Error`,
[`Result.wrapErr`](../Result/wrapErr.md) is the short form:

```ant
type BadLine = { line: int }

func parseSettings(text: string): Result[string, BadLine] {
  if text == "" {
    return Result.err(BadLine{ line: 1 })
  }
  Result.ok(text)
}

let parsed = parseSettings("")
  | Result.mapErr({ Error{ message: "bad line ${x.line}" }.wrap("parse settings") })
if parsed is Err {
  println(parsed.error.message)
}
```

```text
parse settings: bad line 1
```

Wrap an error that is not in a `Result`, such as one a check returns, with the field it is about:

```ant
func checkPort(port: int): Error | null {
  if port < 1 || port > 65535 {
    return Error{ message: "${port} is out of range" }
  }
  null
}

let problem = checkPort(70000)
if problem != null {
  println(problem.wrap("port").message)
}
```

```text
port: 70000 is out of range
```

## Properties and relations

[`Result.wrapErr`](../Result/wrapErr.md) applies `Error.wrap` to the error of an `Err`:

```ant
let e = Error{ message: "timeout" }
let failed: Result[int, Error] = Result.err(e)
let expected: Result[int, Error] = Result.err(Error.wrap(e, "fetch"))
assert(Result.wrapErr(failed, "fetch") == expected)
```

The message is the context, `: `, and the old message:

```ant
let e = Error{ message: "timeout" }
assert(Error.wrap(e, "fetch").message == "fetch" + ": " + e.message)
```

Wrapping leaves the original error as it was:

```ant
let e = Error{ message: "timeout" }
let wrapped = e.wrap("fetch")
e // Error{ message: "timeout" }
```

## Possible issues

`e` must be an `Error`. An error type of your own is a compile error; build an `Error` from it
first:

```ant
type NotFound = { id: int }
let wrapped = Error.wrap(NotFound{ id: 7 }, "load profile")
// error: cannot use NotFound where Error is expected
```

The result is a new record that holds only text. [`Record.same`](../Record/same.md) with the
original is false, and the original cannot be recovered from it. Branch on an error type of your own
instead:

```ant
let e = Error{ message: "timeout" }
Record.same(e.wrap("fetch"), e) // false
```

`Error.wrap` adds `: ` itself, so a context that ends in a colon gets two:

```ant
Error.wrap(Error{ message: "no such file" }, "failed to load config:") // Error{ message: "failed to load config:: no such file" }
```

Name the operation and its argument in lowercase, with no trailing punctuation: `load config`.

## See also

[Result.wrapErr](../Result/wrapErr.md), [Result.mapErr](../Result/mapErr.md),
[Result.err](../Result/err.md), [Record.same](../Record/same.md),
[try](../../language/keywords/try.md)
