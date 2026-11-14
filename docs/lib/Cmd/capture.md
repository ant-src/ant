<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.capture

```ant signature
func Cmd.capture(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Cmd.Output, Error]
```

Cmd.capture runs c with input as its standard input, waits for it to exit, and captures its status
and outputs. A nonzero status is Ok; Err means a stage could not be started or run to its end, and
when ctx is done every stage is killed and the result is Err with ctx's reason.

<!-- end generated -->

## Details

- The result is `Err` for a program that is not found or a `dir` that does not exist.
- `stdout` and `stderr` hold the whole output as text, final newline included. Bytes that are not
  UTF-8 become U+FFFD in them; [Cmd.captureBytes](captureBytes.md) keeps standard output as the
  command wrote it.
- `input` is what the first stage reads: a string as its UTF-8, `Bytes` as they are. By default the
  command reads nothing.
- For a pipeline, `status` is that of the rightmost stage that failed, or 0, and `statuses` holds
  each stage's own. A stage before the last that SIGPIPE ended does not count, though `statuses`
  still shows its 141. The stages' standard error interleaves as it arrives, in no promised order.
- Where the host has signals, a command ended by signal `n` has the status `128 + n`, as a shell
  reports it. On Windows the status is the exit code.

## Basic examples

Run a command and read its status and both outputs. The exit status 2 is `Ok`:

```ant
let script = "echo built; echo 'warning: slow' >&2; exit 2"
let out = try Cmd.capture(Cmd.new("sh", ["-c", script]))
println(out.status)
print(out.stdout)
print(out.stderr)
```

```text
2
built
warning: slow
```

Give the command its standard input as a string:

```ant
let out = try Cmd.capture(Cmd.new("sort"), "pear\napple\nfig\n")
print(out.stdout)
```

```text
apple
fig
pear
```

## Scope

The whole `Cmd.Output` record, for a command that writes to both streams:

```ant
let out = try Cmd.capture(Cmd.new("sh", ["-c", "echo hi; echo oops >&2"]))
println(out)
```

```text
Cmd.Output{ status: 0, statuses: [0], stdout: "hi\n", stderr: "oops\n" }
```

`input` as `Bytes` reaches the command unchanged, bytes that are not text included:

```ant
let data = Bytes.fromArray([0, 1, 254, 255])
let out = try Cmd.capture(Cmd.new("wc", ["-c"]), data)
println(out.stdout.trim())
```

```text
4
```

Output that is not UTF-8 has U+FFFD in `stdout`; [Cmd.captureBytes](captureBytes.md) gives its
bytes:

```ant
let c = Cmd.new("printf", ["\\377\\n"])
println((try Cmd.capture(c)).stdout == "\u{FFFD}\n")
println((try Cmd.captureBytes(c)).stdout)
```

```text
true
Bytes(ff 0a)
```

A pipeline from [Cmd.pipe](pipe.md) is captured as one command. `input` goes to the first stage, and
the last stage's output is `stdout`:

```ant
let words = Cmd.pipe([Cmd.new("tr", ["a-z", "A-Z"]), Cmd.new("sort", ["-u"])])
let out = try Cmd.capture(words, "bee\nant\nbee\n")
print(out.stdout)
println(out.statuses)
```

```text
ANT
BEE
[0, 0]
```

The status of a pipeline is that of the rightmost stage that failed. A first stage that SIGPIPE
ended because a later stage stopped reading does not count, but `statuses` still shows it:

```ant
let failed = try Cmd.capture(Cmd.pipe([Cmd.new("false"), Cmd.new("cat")]))
println("${failed.status} ${failed.statuses}")
let first = try Cmd.capture(Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "1"])]))
println("${first.status} ${first.statuses}")
```

```text
1 [1, 0]
0 [141, 0]
```

A command's `dir` and `env` apply as they do for every function that runs it. `env` is the child's
whole environment:

```ant
let script = "echo \"$GREETING from $(pwd)\""
let c = Cmd.new("sh", ["-c", script], dir: "/", env: Map{ "GREETING": "hello" })
let out = try Cmd.capture(c)
print(out.stdout)
```

```text
hello from /
```

A command that a signal ends has the status `128 + n`, here 137 for SIGKILL:

```ant
let out = try Cmd.capture(Cmd.new("sh", ["-c", "kill -9 $$"]))
println(out.status)
```

```text
137
```

## Named arguments

`input:` names the standard input, which reads well when the command is built inline:

```ant
let out = try Cmd.capture(Cmd.new("wc", ["-l"]), input: "one\ntwo\nthree\n")
println(out.stdout.trim())
```

```text
3
```

`ctx:` bounds the run. When the context is done first, every stage is killed and the result is `Err`
with the context's reason:

```ant
let ctx = Context.new(Time.duration(milliseconds: 100))
println(Cmd.capture(Cmd.new("sleep", ["5"]), ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Applications

`grep` exits with 1 when nothing matches and with 2 on an error. `Cmd.capture` lets a program tell
the three apart, where [Cmd.read](read.md) would fail on both 1 and 2:

```ant
func countMatches(pattern: string, text: string): Result[int, Error] {
  let out = try Cmd.capture(Cmd.new("grep", ["-c", pattern]), text)
  if out.status > 1 {
    return Result.fail("grep: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout.trim().toInt() ?? 0)
}

let log = "ok\nerror: disk\nok\nerror: net\n"
println(try countMatches("error", log))
println(try countMatches("panic", log))
```

```text
2
0
```

Run a list of checks, report each one, and show the standard error of those that failed:

```ant
let checks = [
  Cmd.new("true"),
  Cmd.new("sh", ["-c", "echo 'missing LICENSE' >&2; exit 1"]),
  Cmd.new("test", ["-d", "/"]),
]
var failures = 0
for c in checks {
  let out = try Cmd.capture(c)
  if out.status == 0 {
    println("ok   ${Cmd.string(c)}")
    continue
  }
  failures += 1
  println("FAIL ${Cmd.string(c)}: ${out.stderr.trim()}")
}
println("${failures} failed")
```

```text
ok   true
FAIL sh -c 'echo '\''missing LICENSE'\'' >&2; exit 1': missing LICENSE
ok   test -d /
1 failed
```

## Properties and relations

For output that is UTF-8, `stdout` is what [Cmd.captureBytes](captureBytes.md) gives, read as text:

```ant
let c = Cmd.new("echo", ["héllo"])
let raw = try Cmd.captureBytes(c)
assert(String.fromBytes(raw.stdout) == (try Cmd.capture(c)).stdout)
assert(raw.stdout.length() == 7)
```

On status 0, [Cmd.read](read.md) returns `stdout` without its final newline:

```ant
let c = Cmd.new("printf", ["a\\nb\\n"])
let out = try Cmd.capture(c)
assert(out.status == 0)
assert(try Cmd.read(c) == out.stdout.trimSuffix("\n"))
```

For a single command, `statuses` holds one status, the same as `status`:

```ant
let out = try Cmd.capture(Cmd.new("sh", ["-c", "exit 5"]))
assert(out.statuses == [out.status])
assert(out.status == 5)
```

## Possible issues

`try` passes a failed command through: a nonzero status is `Ok`. Check `status`, or use
[Cmd.read](read.md) when a nonzero status should be `Err`:

```ant
let out = try Cmd.capture(Cmd.new("ls", ["/no/such/file"]))
println(out.status == 0)
println(Cmd.read(Cmd.new("false")))
```

```text
false
Err{ error: Error{ message: "false: exit status 1" } }
```

A program that is not found is `Err`, not the status 127 a shell gives. Only a shell started by the
command reports 127:

```ant
println(Cmd.capture(Cmd.new("no-such-program")))
let out = try Cmd.capture(Cmd.new("sh", ["-c", "no-such-program"]))
println(out.status)
```

```text
Err{ error: Error{ message: "exec: \"no-such-program\": executable file not found in $PATH" } }
127
```

`env` replaces the child's whole environment, so a command given one variable loses `HOME`, `PATH`,
and every other. To add a variable, start from [Os.environ](../Os/environ.md):

```ant
let script = "echo \"$GREETING, home is '$HOME'\""
let bare = try Cmd.capture(Cmd.new("sh", ["-c", script], env: Map{ "GREETING": "hi" }))
print(bare.stdout)

let env = Os.environ()
env.set("GREETING", "hi")
let full = try Cmd.capture(Cmd.new("sh", ["-c", script], env: env))
println(full.stdout.trim() == "hi, home is '${Os.env("HOME") ?? ""}'")
```

```text
hi, home is ''
true
```

`stdout` keeps the final newline the command wrote, so a comparison with the bare value fails. Trim
it, or use [Cmd.read](read.md):

```ant
let out = try Cmd.capture(Cmd.new("echo", ["yes"]))
println(out.stdout == "yes")
println(out.stdout.trimEnd() == "yes")
```

```text
false
true
```

`Cmd.capture` holds the whole output in memory and returns only once the command exits. To read
output as it comes, or more than fits in memory, use [Cmd.start](start.md) with a pipe from
[Os.openPipe](../Os/openPipe.md).

## See also

[Cmd.read](read.md) · [Cmd.readBytes](readBytes.md) · [Cmd.run](run.md) · [Cmd.start](start.md) ·
[Cmd.pipe](pipe.md) · [Cmd.new](new.md) · [Cmd.string](string.md) · [Cmd.Output](types/Output.md) ·
[Os.openPipe](../Os/openPipe.md)
