<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.new

```ant signature
func Cmd.new(program: string, args: Array[string] = [], dir: string | null = null, env: Map[string, string] | null = null): Cmd
```

Cmd.new returns the Cmd of program with args, run in dir and with env when they are given.

<!-- end generated -->

## Details

- `Cmd.new` runs nothing. It returns a `Cmd` record, the same as the literal `Cmd{ program, args }`
  with the defaults filled in. [Cmd.run](run.md), [Cmd.read](read.md), [Cmd.capture](capture.md) and
  [Cmd.start](start.md) run it.
- No shell runs the command. Each element of `args` reaches the program as it is: no globbing, no
  `$VAR` expansion, no splitting on spaces.
- A `program` without a path separator is looked up in the program's own `PATH`, never in the `PATH`
  of `env`. A `program` with a separator is taken relative to `dir`.
- `env` is the child's whole environment. To add one variable, start from
  [`Os.environ()`](../Os/environ.md).
- The `Cmd` holds the `args` array it was given, not a copy, as any record does.

## Basic examples

Make a command from a program and its arguments:

```ant
Cmd.new("git", ["rev-parse", "--short", "HEAD"]) // Cmd{ program: "git", args: ["rev-parse", "--short", "HEAD"], dir: null, env: null }
Cmd.new("ls")                                    // Cmd{ program: "ls", args: [], dir: null, env: null }
```

Run it with a function of `Cmd`:

```ant
let out = try Cmd.read(Cmd.new("echo", ["hello", "world"]))
println(out)
```

```text
hello world
```

## Scope

Each argument arrives as one argument, exactly as written. A shell would expand `*` and `$HOME`, and
split `a b` in two:

```ant
let c = Cmd.new("sh", ["-c", `printf '[%s]\n' "$@"`, "sh", "*", "$HOME", "a b"])
println(try Cmd.read(c))
```

```text
[*]
[$HOME]
[a b]
```

The result is a record, so its fields read and change like any record's:

```ant
let build = Cmd.new("go", ["build"])
build.args.push("./...")
println(build.program)
println(build.string())
```

```text
go
go build ./...
```

## Named arguments

### args

`args` can be named, which keeps the call readable when `dir` or `env` follows it. The two calls
make the same command:

```ant
Cmd.new("ls", args: ["-a"], dir: "/") == Cmd.new("ls", ["-a"], "/") // true
```

### dir

`dir` sets the working directory of the child. Without it the child runs in the program's working
directory:

```ant
println(try Cmd.read(Cmd.new("pwd", dir: "/")))
```

```text
/
```

### env

`env` replaces the child's environment. Copy `Os.environ()` and add to it to keep `PATH`, `HOME` and
the rest:

```ant
let env = Os.environ()
env.set("GREETING", "hi")
println(try Cmd.read(Cmd.new("sh", ["-c", `echo "$GREETING"`], env: env)))
```

```text
hi
```

## Applications

A helper that returns a `Cmd` keeps how a tool is called in one place and leaves how to run it to
the caller:

```ant
func git(repo: string, args: Array[string]): Cmd {
  Cmd.new("git", ["-C", repo, ...args])
}

println(git("/srv/app", ["log", "-n", "1"]).string())
```

```text
git -C /srv/app log -n 1
```

Build the arguments from data before running. Options come before the file names:

```ant
let files = ["notes.txt", "todo list.txt"]
let verbose = true
let flags: Array[string] = if verbose { ["-v"] } else { [] }
let c = Cmd.new("rm", [...flags, "--", ...files])
println(c.string())
```

```text
rm -v -- notes.txt 'todo list.txt'
```

Run one command per input:

```ant
let words = ["one", "two", "three"]
for c in words.map({ Cmd.new("echo", [x.toUpper()]) }) {
  println(try Cmd.read(c))
}
```

```text
ONE
TWO
THREE
```

## Properties and relations

`Cmd.new` is the record literal with its defaults:

```ant
let c = Cmd.new("ls", ["-l"])
assert(c == Cmd{ program: "ls", args: ["-l"] })
assert(c.dir == null && c.env == null)
```

Making a command checks neither the program nor `dir`; running it does:

```ant
let c = Cmd.new("no-such-program")
assert(c.program == "no-such-program")
assert(!Cmd.read(c).isOk())
let d = Cmd.new("pwd", dir: "no-such-dir")
assert(!Cmd.read(d).isOk())
```

## Possible issues

A command line in one string is a program name, not a command. No shell splits it, so the lookup
fails:

```ant
println(Cmd.read(Cmd.new("echo hi")).isOk())
println(try Cmd.read(Cmd.new("echo", ["hi"])))
```

```text
false
hi
```

Pass the line to `sh -c` only when you want a shell, with its quoting and expansion.

`dir` does not change where a program without a path separator is found. `hello` is looked up in
`PATH`, not in `tools`; write `./hello`:

<!-- test: before
try File.mkdir("tools")
try File.write("tools/hello", "#!/bin/sh\necho hello\n")
try File.chmod("tools/hello", 0o755)
-->

```ant
println(Cmd.read(Cmd.new("hello", dir: "tools")).isOk())
println(try Cmd.read(Cmd.new("./hello", dir: "tools")))
```

```text
false
hello
```

The command shares the array passed as `args`. Changing the array afterwards changes the command:

```ant
let args = ["a"]
let c = Cmd.new("echo", args)
args.push("b")
println(try Cmd.read(c))
```

```text
a b
```

Pass a copy, such as `[...args]`, to keep the command as it was.

A map with only the new variable leaves the child without the rest of the environment, such as
`HOME`:

```ant
let c = Cmd.new("sh", ["-c", `echo "[$GREETING] [$HOME]"`], env: Map{ GREETING: "hi" })
println(try Cmd.read(c))
```

```text
[hi] []
```

## See also

[Cmd](index.md) · [Cmd.pipe](pipe.md) · [Cmd.string](string.md) · [Cmd.run](run.md) ·
[Cmd.read](read.md) · [Cmd.capture](capture.md) · [Cmd.start](start.md) · [Os.which](../Os/which.md)
· [Os.environ](../Os/environ.md)
