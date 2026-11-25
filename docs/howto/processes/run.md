<!-- status: reviewed -->

# Run a command and capture its output

Run another program and use what it prints.

1. Build the command with [Cmd.new](../../lib/Cmd/new.md): the program, then its arguments as an
   array. No shell runs it, so each argument reaches the program as it is.
2. Call [Cmd.read](../../lib/Cmd/read.md) to get what the command prints, without its final line
   break. Any nonzero status is an `Err`.
3. Call [Cmd.capture](../../lib/Cmd/capture.md) instead when a nonzero status is an answer, or when
   you need standard error too.
4. Use [`try`](../../language/keywords/try.md) to return the `Err`, and
   [Result.wrapErr](../../lib/Result/wrapErr.md) to say what the program was doing.

## Read what a command prints

This script prints the TODO lines of `notes.txt` that `grep` finds:

<!-- test: file notes.txt
buy milk
TODO: call Ann
TODO: pay rent
-->

```ant
let todos = try Cmd.read(Cmd.new("grep", ["TODO", "notes.txt"]))
for line in todos.lines() {
  println(line.trimPrefix("TODO: "))
}
```

```text
call Ann
pay rent
```

Call [Cmd.readBytes](../../lib/Cmd/readBytes.md) for output that is not text.

## Handle a command that fails

When the command exits with a nonzero status, `Cmd.read` returns an `Err` whose message names the
program and its status, followed by what it wrote to standard error. Add context with
`Result.wrapErr`:

```ant
let c = Cmd.new("sh", ["-c", "echo 'disk full' >&2; exit 3"])
let saved = Cmd.read(c) | Result.wrapErr("save report")
println(saved)
```

```text
Err{ error: Error{ message: "save report: sh: exit status 3: disk full", kind: null } }
```

With `try`, a script stops there, writes the message to standard error, and exits with status 1:

<!-- test: exit 1 -->

```ant
let c = Cmd.new("sh", ["-c", "echo 'disk full' >&2; exit 3"])
let saved = try (Cmd.read(c) | Result.wrapErr("save report"))
println(saved)
```

A program that cannot be found is an `Err` too. Call [Os.which](../../lib/Os/which.md) first to
report it your own way; it returns `null` when the `PATH` has no such program:

<!-- test: exit 1 -->
<!-- test: env PATH= -->

```ant
if Os.which("convert") == null {
  return Result.fail("install ImageMagick to resize images")
}
```

## Keep the status and both outputs

`grep` exits with 1 when nothing matches and with 2 on an error, so `Cmd.read` fails on both.
`Cmd.capture` returns `Ok` for every status, with the status, standard output, and standard error in
one [Cmd.Output](../../lib/Cmd/types/Output.md) record:

<!-- test: file clean.txt
all done
-->

```ant
func countTodos(path: string): Result[int, Error] {
  let out = try Cmd.capture(Cmd.new("grep", ["-c", "TODO", path]))
  if out.status > 1 {
    return Result.fail("grep ${path}: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout.trim().toInt() ?? 0)
}

for path in ["notes.txt", "clean.txt"] {
  println("${path}: ${try countTodos(path)}")
}
```

```text
notes.txt: 2
clean.txt: 0
```

The outputs keep their final line breaks, so trim them with [String.trim](../../lib/String/trim.md)
before you compare or convert them.

## Give the command input

Pass the text the command reads from standard input as `input:`. Without it, the command reads
nothing:

```ant
let sorted = try Cmd.read(Cmd.new("sort", ["-u"]), input: "pear\napple\npear\n")
println(sorted.lines())
```

```text
["apple", "pear"]
```

## Choose the directory and environment

Give `dir:` to run the command in another directory. `env:` is the command's whole environment, so
start from [Os.environ](../../lib/Os/environ.md) and add to the copy with
[Map.set](../../lib/Map/set.md):

<!-- test: file reports/today.txt
done
-->

```ant
let env = Os.environ()
env.set("GREETING", "hello")
let c = Cmd.new("sh", ["-c", "echo \"$GREETING\" $(ls)"], dir: "reports", env: env)
println(try Cmd.read(c))
```

```text
hello today.txt
```

## Limit how long it runs

Pass a context with a deadline as `ctx:`. When it ends first, the command is killed and the result
is an `Err` with the context's reason:

```ant
let ctx = Context.new(100ms)
println(Cmd.read(Cmd.new("sleep", ["5"]), ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout", kind: null } }
```

[Stop or time out a process](stop.md) shows more.

## Log the command

[Cmd.string](../../lib/Cmd/string.md) writes the command as a shell would need it quoted, and leaves
out `dir` and `env`:

```ant
let c = Cmd.new("grep", ["-c", "to do", "notes.txt"])
println("running ${Cmd.string(c)}")
```

```text
running grep -c 'to do' notes.txt
```

## See also

[Cmd.new](../../lib/Cmd/new.md) · [Cmd.read](../../lib/Cmd/read.md) ·
[Cmd.readBytes](../../lib/Cmd/readBytes.md) · [Cmd.capture](../../lib/Cmd/capture.md) ·
[Cmd.Output](../../lib/Cmd/types/Output.md) · [Cmd.string](../../lib/Cmd/string.md) ·
[Os.which](../../lib/Os/which.md) · [Os.environ](../../lib/Os/environ.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [Result.fail](../../lib/Result/fail.md) ·
[Context.new](../../lib/Context/new.md) · [Map.set](../../lib/Map/set.md) ·
[String.lines](../../lib/String/lines.md) · [String.trim](../../lib/String/trim.md) ·
[String.trimPrefix](../../lib/String/trimPrefix.md) · [String.toInt](../../lib/String/toInt.md) ·
[Cmd.run](../../lib/Cmd/run.md) · [Stream a command's output](stream.md) ·
[Build a pipeline](pipeline.md) · [Stop or time out a process](stop.md)
