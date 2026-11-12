<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.environ

```ant signature
func Os.environ(): Map[string, string]
```

Os.environ returns every environment variable, keyed by name, in the host's order, which is no order
to rely on. Each call returns a new map, which the caller may change without changing the
environment: a later Os.env, and a child without env, see it as it was.

<!-- end generated -->

## Details

- A variable that is not set has no key in the map, so `get` returns `null` for it.
- To read one variable, call [`Os.env`](env.md); `Os.environ` is for working with the whole set.
- A command's `env` is the child's whole environment. To give a child one more variable, start from
  `Os.environ()` and add to it.

## Basic examples

`Os.environ` returns a `Map[string, string]` from each variable's name to its value:

```ant
let env = Os.environ()
println(env.has("PATH"))
println(env.get("ANT_DOCS_NOT_SET"))
```

```text
true
null
```

Print every variable, sorted by name:

<!-- test: check-only the environment differs between machines -->

```ant
for e in Os.environ().entries().sortBy({ x.key }) {
  println("${e.key}=${e.value}")
}
```

## Scope

The values are strings. Indexing the map gives `string | null`, so give a fallback with `??`:

```ant
let home = Os.environ()["HOME"] ?? "/"
println(home.isEmpty())
```

```text
false
```

Map functions apply as to any map. Keep the variables whose name ends with `PATH`:

```ant
let paths = Os.environ().filter({ x.key.endsWith("PATH") })
println(paths.has("PATH"))
println(paths.keys().every({ x.endsWith("PATH") }))
```

```text
true
true
```

## Applications

Run a child process with one more variable. The map from `Os.environ` keeps everything else the
child needs, such as `PATH` and `HOME`:

```ant
let env = Os.environ()
env.set("GREETING", "hello from the parent")
println(try Cmd.read(Cmd.new("sh", ["-c", `echo "$GREETING"`], env: env)))
```

```text
hello from the parent
```

Keep secrets away from a child by filtering the map before passing it:

```ant
let env = Os.environ()
env.set("API_TOKEN", "s3cret")
let safe = env.filter({ !x.key.endsWith("_TOKEN") })
let script = `echo "token: ${API_TOKEN:-none}"`
println(try Cmd.read(Cmd.new("sh", ["-c", script], env: env)))
println(try Cmd.read(Cmd.new("sh", ["-c", script], env: safe)))
```

```text
token: s3cret
token: none
```

Read settings from variables that share a prefix. Writing the function against a `Map` lets a test
pass a literal, while the program passes `Os.environ()`:

```ant
func settings(env: Map[string, string], prefix: string): Map[string, string] {
  env.entries()
    | Array.filter({ x.key.startsWith(prefix) })
    | Array.map({ Entry{ key: x.key.trimPrefix(prefix).toLower(), value: x.value } })
    | Map.fromEntries()
}

let sample = Map{ APP_PORT: "8080", APP_DEBUG: "1", HOME: "/home/ann" }
println(settings(sample, "APP_"))
println(settings(Os.environ(), "ANT_DOCS_EXAMPLE_").isEmpty())
```

```text
Map{ "port": "8080", "debug": "1" }
true
```

## Properties and relations

`Os.environ().get(name)` is `Os.env(name)` for every name, set or not:

```ant
let env = Os.environ()
for e in env {
  assert(Os.env(e.key) == e.value)
}
assert(env.get("ANT_DOCS_NOT_SET") == Os.env("ANT_DOCS_NOT_SET"))
```

## Possible issues

The result is a map, not a record, so a variable is not a field:

```ant
let env = Os.environ()
println(env.PATH) // error: Map[string, string] is not a record, so it has no field PATH
```

Indexing gives `string | null`, so using it as a `string` is a compile error. Give a fallback with
`??`, or call [`Os.env`](env.md) and handle `null`:

```ant
let home: string = Os.environ()["HOME"] // error: cannot use string | null where string is expected
```

A command's `env` replaces the child's environment instead of adding to it. A map with only the new
variable leaves the child without `HOME`:

```ant
let only = Map{ GREETING: "hi" }
println(try Cmd.read(Cmd.new("sh", ["-c", `echo "[$GREETING] [$HOME]"`], env: only)))
```

```text
[hi] []
```

Start from `Os.environ()` instead, as in Applications.

## See also

[Os.env](env.md) · [Os](index.md) · [Cmd.new](../Cmd/new.md) · [Cmd](../Cmd/index.md) ·
[Map.get](../Map/get.md) · [Map.filter](../Map/filter.md)
