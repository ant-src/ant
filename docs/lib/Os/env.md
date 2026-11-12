<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.env

```ant signature
func Os.env(name: string): string | null
```

Os.env returns the value of the environment variable name, or null when it is not set; a variable
set empty gives "". Names are matched as the host matches them: exactly on Unix, and ignoring case
on Windows. An empty name, or one holding =, names no variable and gives null.

<!-- end generated -->

## Details

- `null` means the variable is not set.
- `Os.env` reads the program's own environment. The `env` of a [Cmd](../Cmd/new.md) sets a child's
  environment and does not change what `Os.env` returns.

## Basic examples

A variable that is not set gives `null`:

```ant
println(Os.env("APP_NO_SUCH_VARIABLE"))
println(Os.env("PATH") != null)
```

```text
null
true
```

A variable that is set gives its value, which depends on the machine, such as `/Users/ann` for
`HOME`:

<!-- test: check-only the value of HOME differs between machines -->

```ant
println(Os.env("HOME"))
```

## Scope

The result is `string | null`, so give a default with `??` or narrow it before using it as a string:

```ant
let editor = Os.env("APP_EDITOR") ?? "vi"
println(editor)

let path = Os.env("PATH")
if path != null {
  println(path.isEmpty())
}
```

```text
vi
false
```

Optional chaining parses a value only when there is one:

```ant
let port = Os.env("APP_PORT")?.toInt() ?? 8080
println(port)
```

```text
8080
```

## Applications

Read a setting from the environment, default it when it is missing, and reject it when it is
malformed:

```ant
func parsePort(text: string | null): Result[int, Error] {
  if text == null { return Result.ok(8080) }
  let port = text.toInt()
  if port == null || port < 1 || port > 65535 {
    return Result.fail("APP_PORT: want a port from 1 to 65535, got ${text}")
  }
  Result.ok(port)
}

println(parsePort(Os.env("APP_PORT")))
println(parsePort("3000"))
println(parsePort("http"))
```

```text
Ok{ value: 8080 }
Ok{ value: 3000 }
Err{ error: Error{ message: "APP_PORT: want a port from 1 to 65535, got http" } }
```

Require a variable, and treat an empty value as missing:

```ant
func required(name: string): Result[string, Error] {
  let value = Os.env(name)
  if value == null || value.isEmpty() {
    return Result.fail("${name} is not set")
  }
  Result.ok(value)
}

println(required("PATH").isOk())
println(required("APP_DATABASE_URL"))
```

```text
true
Err{ error: Error{ message: "APP_DATABASE_URL is not set" } }
```

Give a child a small environment of its own, with the program's `PATH` and one variable more:

```ant
let child = Cmd.new(
  "sh",
  ["-c", "echo $GREETING"],
  env: Map{ "PATH": Os.env("PATH") ?? "", "GREETING": "hello" },
)
println(try Cmd.read(child))
```

```text
hello
```

## Properties and relations

`Os.env` agrees with [Os.environ](environ.md), which holds every variable that is set:

```ant
let env = Os.environ()
for e in env.entries() {
  assert(Os.env(e.key) == e.value)
}
assert(env["APP_NO_SUCH_VARIABLE"] == null)
assert(Os.env("APP_NO_SUCH_VARIABLE") == null)
```

## Possible issues

The result may be `null`, so it is not a `string`:

```ant
let home: string = Os.env("HOME") // error: cannot use string | null where string is expected
```

Give a default, or handle the missing case:

```ant
let home: string = Os.env("HOME") ?? "/"
println(home.isEmpty())
```

```text
false
```

## See also

[Os.environ](environ.md) · [Os.home](home.md) · [Os.which](which.md) · [Os.args](args.md) ·
[Cmd.new](../Cmd/new.md) · [??](../../language/operators/null-coalescing.md) ·
[?.](../../language/operators/optional-chaining.md)
