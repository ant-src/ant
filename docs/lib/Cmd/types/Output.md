<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.Output

```ant signature
type Cmd.Output = { status: int, statuses: Array[int], stdout: string, stderr: string }
```

Cmd.Output is what [`Cmd.capture`](../capture.md) collects from a command that ran to its end.
status is its status, statuses each stage's own, and stdout and stderr the outputs as text, with
bytes that are not valid UTF-8 as U+FFFD. The stages' standard error interleaves as it arrived.

<!-- end generated -->

## Details

- A nonzero status is not an error: [Cmd.capture](../capture.md) returns `Ok` with the status in
  `status`. It is `Err` only when the command could not be started or run to its end.
- `stdout` and `stderr` hold the whole output, final newline included, unlike
  [Cmd.read](../read.md).
- Output that is not valid UTF-8 has U+FFFD in `stdout` or `stderr`.
  [Cmd.captureBytes](../captureBytes.md) gives a [Cmd.OutputBytes](OutputBytes.md), which holds
  standard output as it was written.
- `statuses` has one entry per stage, in pipeline order. For a pipeline, `status` is that of the
  rightmost stage that failed, or 0, and a stage before the last that SIGPIPE ended does not count.
- A child that a signal ended has status `128 + n`, as a shell reports it: SIGKILL is 137. On
  Windows the status is the exit code.

## Basic examples

Capture a command's status and both outputs with [Cmd.capture](../capture.md):

```ant
let out = try Cmd.capture(Cmd.new("sh", ["-c", "echo hello; echo oops >&2; exit 3"]))
println(out.status)
println(out.stdout.trim())
println(out.stderr.trim())
```

```text
3
hello
oops
```

A `Cmd.Output` prints as a record:

```ant
println(try Cmd.capture(Cmd.new("echo", ["hi"])))
```

```text
Cmd.Output{ status: 0, statuses: [0], stdout: "hi\n", stderr: "" }
```

## Scope

A command that writes nothing has empty outputs:

```ant
let out = try Cmd.capture(Cmd.new("true"))
println(out.stdout.isEmpty())
println(out.stderr.isEmpty())
```

```text
true
true
```

What `input` sends the command shows up in its output:

```ant
let out = try Cmd.capture(Cmd.new("tr", ["a-z", "A-Z"]), input: "shout")
println(out.stdout)
```

```text
SHOUT
```

A pipeline has one entry in `statuses` per stage. Its `status` is the rightmost failure:

```ant
let out = try Cmd.capture(Cmd.pipe([
  Cmd.new("sh", ["-c", "echo a; exit 2"]),
  Cmd.new("cat"),
  Cmd.new("cat"),
]))
println(out.status)
println(out.statuses)
```

```text
2
[2, 0, 0]
```

A command that a signal ended has status `128 + n`. SIGKILL is signal 9:

```ant
let out = try Cmd.capture(Cmd.new("sh", ["-c", "kill -KILL $$"]))
println(out.status)
println(out.statuses)
```

```text
137
[137]
```

A stage that stopped because a later stage stopped reading shows SIGPIPE's 141 in `statuses`, but
does not set `status`:

```ant
let out = try Cmd.capture(Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "1"])]))
println(out.status)
println(out.statuses)
println(out.stdout)
```

```text
0
[141, 0]
y

```

Output that is not UTF-8 has the bad byte as U+FFFD in `stdout`;
[Cmd.captureBytes](../captureBytes.md) keeps it exact:

```ant
let c = Cmd.new("printf", ["\\377ok"])
println((try Cmd.capture(c)).stdout == "\u{FFFD}ok")
println((try Cmd.captureBytes(c)).stdout)
```

```text
true
Bytes(ff 6f 6b)
```

`Cmd.Output` is a record type, so a literal builds one, for example to test a function that takes an
output without running a command:

```ant
func summary(out: Cmd.Output): string {
  if out.status == 0 { "ok" } else { "failed (${out.status}): ${out.stderr.trim()}" }
}

let failed = Cmd.Output{
  status: 1,
  statuses: [1],
  stdout: "",
  stderr: "disk full\n",
}
println(summary(failed))
```

```text
failed (1): disk full
```

## Applications

Tell `grep`'s three outcomes apart: 0 for a match, 1 for none, and 2 or more for trouble:

```ant
func search(pattern: string, text: string): Result[bool, Error] {
  let out = try Cmd.capture(Cmd.new("grep", ["-q", pattern]), input: text)
  if out.status > 1 {
    return Result.fail("grep: ${out.stderr.trim()}")
  }
  Result.ok(out.status == 0)
}

println(search("fix", "fix typo\nadd tests\n"))
println(search("docs", "fix typo\nadd tests\n"))
println(search("[", "fix typo\n") is Err)
```

```text
Ok{ value: true }
Ok{ value: false }
true
```

Report a failed step with what it wrote to standard error:

```ant
func step(name: string, script: string): Result[string, Error] {
  let out = try Cmd.capture(Cmd.new("sh", ["-c", script]))
  if out.status != 0 {
    return Result.fail("${name} failed with status ${out.status}: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout.trim())
}

println(step("version", "echo 1.4.2"))
println(step("lint", "echo 'main.ant:3: unused variable' >&2; exit 1"))
```

```text
Ok{ value: "1.4.2" }
Err{ error: Error{ message: "lint failed with status 1: main.ant:3: unused variable" } }
```

Find which stage of a pipeline failed:

```ant
let stages = [Cmd.new("echo", ["data"]), Cmd.new("grep", ["missing"]), Cmd.new("cat")]
let out = try Cmd.capture(Cmd.pipe(stages))
let failed = out.statuses.findIndex({ x != 0 })
if failed != null {
  println("stage ${failed + 1} (${stages[failed].program}) exited ${out.statuses[failed]}")
}
```

```text
stage 2 (grep) exited 1
```

## Properties and relations

`stdout` is the output [Cmd.captureBytes](../captureBytes.md) gives, decoded as text, and `stderr`
is the same in both:

```ant
for script in ["echo plain", "printf 'caf\\303\\251'", "printf '\\377\\376'", "echo err >&2"] {
  let c = Cmd.new("sh", ["-c", script])
  let out = try Cmd.capture(c)
  let raw = try Cmd.captureBytes(c)
  assert(out.stdout == String.fromBytesLossy(raw.stdout))
  assert(out.stderr == raw.stderr)
}
```

For a single command, `statuses` holds only `status`:

```ant
for code in [0, 1, 7] {
  let out = try Cmd.capture(Cmd.new("sh", ["-c", "exit ${code}"]))
  assert(out.status == code)
  assert(out.statuses == [code])
}
```

[Cmd.read](../read.md) returns `stdout` with one final newline removed:

```ant
let c = Cmd.new("printf", ["a\\nb\\n\\n"])
let out = try Cmd.capture(c)
assert(out.stdout == "a\nb\n\n")
assert(try Cmd.read(c) == "a\nb\n")
```

## Possible issues

`try` does not stop on a failing command, since a nonzero status is still `Ok`. Check `status`:

```ant
let out = try Cmd.capture(Cmd.new("false"))
println("still here, status ${out.status}")
```

```text
still here, status 1
```

`stdout` ends with the command's newline, so comparing it to a bare word fails. Trim it, or use
[Cmd.read](../read.md):

```ant
let out = try Cmd.capture(Cmd.new("echo", ["yes"]))
println(out.stdout == "yes")
println(out.stdout.trim() == "yes")
```

```text
false
true
```

A command that cannot be started gives no `Cmd.Output` at all, but an `Err`:

```ant
let result = Cmd.capture(Cmd.new("no-such-program"))
println(result is Err)
```

```text
true
```

## See also

[Cmd.capture](../capture.md) · [Cmd.read](../read.md) · [Cmd.readBytes](../readBytes.md) ·
[Cmd.run](../run.md) · [Cmd.pipe](../pipe.md) · [Cmd.Pipeline](Pipeline.md) ·
[String.fromBytesLossy](../../String/fromBytesLossy.md)
