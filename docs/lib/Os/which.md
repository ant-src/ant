<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.which

```ant signature
func Os.which(program: string): string | null
```

Os.which returns the path of the executable in PATH that a [`Cmd`](../Cmd/index.md) naming program
starts, or null when PATH holds none. A program with a path separator is returned unchanged when it
names an executable file, taken against the working directory, and null otherwise.

<!-- end generated -->

## Details

- `Os.which` looks `program` up in the program's own `PATH`, `Os.env("PATH")`, as a
  [Cmd](../Cmd/new.md) does. The `env` of a `Cmd` never changes that lookup. An entry of `PATH` that
  is not absolute, `.` included, finds nothing.

## Basic examples

Find a program in `PATH`:

<!-- test: check-only the path differs between machines -->

```ant
println(Os.which("sh"))
```

A program that no directory of `PATH` holds gives `null`:

```ant
println(Os.which("no-such-program"))
```

```text
null
```

## Scope

The result is `string | null`. Narrow it before using it as a path; the path is absolute and ends
with the program's name:

```ant
let sh = Os.which("sh")
if sh != null {
  println(Path.isAbs(sh))
  println(Path.base(sh))
}
```

```text
true
sh
```

Give a default with `??` to fall back to the bare name:

```ant
let shell = Os.which("no-such-shell") ?? "sh"
println(shell)
```

```text
sh
```

An empty name, or one with a trailing space, finds nothing:

```ant
println(Os.which(""))
println(Os.which("sh "))
```

```text
null
null
```

## Applications

Fail early, with a message that says what to install, when a script needs a tool:

```ant
func requireTool(name: string): Result[string, Error] {
  let path = Os.which(name) ?? return Result.fail("${name} not found in PATH; install it first")
  Result.ok(path)
}

println(requireTool("no-such-tool"))
println(requireTool("sh").isOk())
```

```text
Err{ error: Error{ message: "no-such-tool not found in PATH; install it first" } }
true
```

Pick the first program of several that is installed, such as a pager:

```ant
let pager = ["no-such-pager", "cat"].find({ Os.which(x) != null }) ?? "more"
println(pager)
```

```text
cat
```

Fall back to Ant code when a tool is missing:

```ant
func countLines(text: string): Result[string, Error] {
  if Os.which("wc") == null { return Result.ok("${text.split("\n").length()}") }
  let count = try Cmd.read(Cmd.new("wc", ["-l"]), input: "${text}\n")
  Result.ok(count.trim())
}

println(try countLines("a\nb\nc"))
```

```text
3
```

## Properties and relations

`null` from `Os.which` means a `Cmd` naming the program fails to start:

```ant
let missing = "no-such-program"
assert(Os.which(missing) == null)
assert(!Cmd.run(Cmd.new(missing)).isOk())
```

A `Cmd` naming a program that `Os.which` finds starts it:

```ant
assert(Os.which("sh") != null)
let out = try Cmd.read(Cmd.new("sh", ["-c", "echo ok"]))
assert(out == "ok")
```

## Possible issues

Setting `PATH` in a command's `env` does not change where the command's program is found. The
program's own `PATH` still finds `sh` here, and `Os.which` agrees with it:

```ant
let c = Cmd.new("sh", ["-c", "echo ok"], env: Map{ "PATH": "/no-such-dir" })
println(Cmd.read(c))
println(Os.which("sh") != null)
```

```text
Ok{ value: "ok" }
true
```

## See also

[Cmd.new](../Cmd/new.md) · [Cmd.run](../Cmd/run.md) · [Cmd.read](../Cmd/read.md) ·
[Cmd.start](../Cmd/start.md) · [Os.env](env.md) · [Path.base](../Path/base.md)
