<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.start

```ant signature
func Cmd.start(
  c: Cmd | Cmd.Pipeline, stdin: File.Handle | null = null, stdout: File.Handle | null = Os.stdout(),
  stderr: File.Handle | null = Os.stderr(), ctx: Context | null = null,
): Result[Cmd.Process, Error]
```

Cmd.start starts c in the background, every stage in one process group of its own, and returns once
it is running. Its standard streams are the handles given, null being the null device; standard
input is the null device and the outputs are the program's unless given. A terminal as stdin is Err.
Every handle stays the program's, which closes its copy of a pipe end it gave away. When ctx is done
every stage is killed, as when the run is cancelled; a command still running when the script ends is
left running.

<!-- end generated -->

## Details

- `Cmd.start` returns as soon as the command is running. Its `Err` means nothing was started: a
  program not found, a `dir` that does not exist, or a `ctx` already done. A nonzero exit shows up
  in [Cmd.wait](wait.md).
- Standard input defaults to the null device, as a shell's `cmd &` has it, so a command that reads
  its input sees the end at once. A terminal as `stdin` is `Err`, and nothing is started.
- Every handle given stays open and the program's. Close the program's copy of a pipe's write end
  after the call, or the read end never reaches its end.
- A child writing into a pipe nobody reads blocks once the pipe is full: read the pipe before the
  wait, or on another goroutine.
- The command runs in a process group of its own, so the terminal's Ctrl-C does not reach it, and a
  command still running when the script ends is left running. Stop it in a deferred block.

## Basic examples

Start a command and wait for its status. Its output goes to the program's standard output:

```ant
let child = try Cmd.start(Cmd.new("echo", ["hello"]))
println(try child.wait())
```

```text
hello
0
```

Read a command's output as it comes, through a pipe from [Os.openPipe](../Os/openPipe.md):

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("seq", ["3"]), stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println("got ${line}")
}
println(try child.wait())
```

```text
got 1
got 2
got 3
0
```

## Scope

A pipeline from [Cmd.pipe](pipe.md) starts every stage at once. Its last stage writes to `stdout`:

```ant
let words = Cmd.pipe([Cmd.new("printf", ["b\\na\\nc\\n"]), Cmd.new("sort")])
let child = try Cmd.start(words)
println(try child.wait())
```

```text
a
b
c
0
```

The status of a pipeline is that of the rightmost stage that failed, so `false | cat` fails:

```ant
let child = try Cmd.start(Cmd.pipe([Cmd.new("false"), Cmd.new("cat")]))
println(try child.wait())
```

```text
1
```

A command that runs and fails is still started: `Cmd.start` is `Ok`, and the status comes from the
wait:

```ant
let child = try Cmd.start(Cmd.new("sh", ["-c", "exit 3"]))
println(try child.wait())
```

```text
3
```

The `dir` and `env` of the `Cmd` apply as they do for every function of `Cmd`:

```ant
try File.mkdir("site")
let env = Os.environ()
env["GREETING"] = "hi"
let greet = Cmd.new("sh", ["-c", "echo $GREETING from $(basename $PWD)"], dir: "site", env: env)
let child = try Cmd.start(greet)
println(try child.wait())
```

```text
hi from site
0
```

## Named arguments

### stdin

Give a file as the command's standard input:

<!-- test: file names.txt
Cid
Ann
Bob
-->

```ant
let names = try File.open("names.txt")
defer { names.close() }
let child = try Cmd.start(Cmd.new("sort"), stdin: names)
println(try child.wait())
```

```text
Ann
Bob
Cid
0
```

Pass the program's own standard input along with `Os.stdin()` when it is not a terminal:

<!-- test: stdin
ant
bee
-->

```ant
let child = try Cmd.start(Cmd.new("tr", ["a-z", "A-Z"]), stdin: Os.stdin())
println(try child.wait())
```

```text
ANT
BEE
0
```

Write the command's input a piece at a time through a pipe. Closing the write end is what lets
`sort` see the end of its input:

```ant
let input = try Os.openPipe()
let child = try Cmd.start(Cmd.new("sort"), stdin: input.read)
try input.read.close()
for word in ["cid", "ann", "bob"] {
  try input.write.writeLine(word)
}
try input.write.close()
println(try child.wait())
```

```text
ann
bob
cid
0
```

### stdout

Send the output to a file:

```ant
let log = try File.create("echo.log")
defer { log.close() }
let child = try Cmd.start(Cmd.new("echo", ["to the file"]), stdout: log)
println(try child.wait())
print(try File.read("echo.log"))
```

```text
0
to the file
```

`null` discards the output, as `> /dev/null` does:

```ant
let child = try Cmd.start(Cmd.new("echo", ["nobody sees this"]), stdout: null)
println(try child.wait())
```

```text
0
```

### stderr

Standard error goes to the program's standard error unless given. Give `Os.stdout()` to send it to
standard output instead, as `2>&1` does:

```ant
let child = try Cmd.start(Cmd.new("sh", ["-c", "echo out; echo err >&2"]), stderr: Os.stdout())
println(try child.wait())
```

```text
out
err
0
```

Give the same handle twice to merge the outputs into one pipe, as `2>&1` does:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("sh", ["-c", "echo out; echo err >&2"]), stdout: out.write, stderr: out.write)
try out.write.close()
print(try out.read.readAll())
println(try child.wait())
```

```text
out
err
0
```

### ctx

When `ctx` is done, every stage is killed, waited for or not, and every wait is `Err` with the
context's reason:

```ant
let ctx = Context.new(Time.duration(milliseconds: 100))
let child = try Cmd.start(Cmd.new("sleep", ["5"]), ctx: ctx)
println(child.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context already done starts nothing:

```ant
let ctx = Context.new()
ctx.cancel()
println(Cmd.start(Cmd.new("echo", ["never"]), ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Run commands at once and wait for all of them. Each writes its own log, so the outputs stay apart:

```ant
type Job = { name: string, child: Cmd.Process }
let running: Array[Job] = []
for name in ["lint", "test", "build"] {
  let log = try File.create("${name}.log")
  defer { log.close() }
  let child = try Cmd.start(Cmd.new("sh", ["-c", "sleep 0.1; echo ${name} ok"]), stdout: log)
  running.push(Job{ name, child })
}
for job in running {
  let status = try job.child.wait()
  let output = try File.read("${job.name}.log")
  println("${job.name}: ${status}, ${output.trim()}")
}
```

```text
lint: 0, lint ok
test: 0, test ok
build: 0, build ok
```

Write a header into a log, then the command's output after it. A buffered handle is flushed before
the child gets it, so the header comes first:

```ant
let log = File.buffered(try File.create("build.log"))
defer { log.close() }
try log.writeLine("== build ==")
let child = try Cmd.start(Cmd.new("echo", ["compiled 3 files"]), stdout: log)
try log.writeLine("== exit ${try child.wait()} ==")
try log.flush()
print(try File.read("build.log"))
```

```text
== build ==
compiled 3 files
== exit 0 ==
```

Keep a helper running in the background while the script works, and stop it in a deferred block so
it does not outlive the script:

```ant
let helper = try Cmd.start(Cmd.new("sh", ["-c", "while true; do sleep 1; done"]))
defer { helper.stop() }
println("working")
```

```text
working
```

## Properties and relations

Every wait of a process gets the same outcome:

```ant
let child = try Cmd.start(Cmd.new("sh", ["-c", "exit 2"]))
assert(child.wait() == child.wait())
assert(try child.wait() == 2)
```

For a command that reads no input, starting it and waiting gives the status [Cmd.run](run.md) gives:

```ant
let c = Cmd.new("sh", ["-c", "exit 4"])
let child = try Cmd.start(c)
assert(try child.wait() == try Cmd.run(c))
```

## Possible issues

The default standard input is the null device, not the program's. A command that reads its input
sees the end at once, and `cat` here copies nothing:

<!-- test: stdin
never read
-->

```ant
let child = try Cmd.start(Cmd.new("cat"))
println(try child.wait())
```

```text
0
```

The read end of a pipe stays open while the program holds its copy of the write end, even after the
child exits. Without the close, `readAll` waits forever; here a context stops it:

```ant
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(Cmd.new("echo", ["done"]), stdout: out.write)
println(try child.wait())
println(out.read.readAll(ctx: Context.new(Time.duration(milliseconds: 100))))
try out.write.close()
```

```text
0
Err{ error: Error{ message: "timeout" } }
```

A program that is not found is `Err`, and nothing is started:

```ant
println(Cmd.start(Cmd.new("no-such-program")))
```

```text
Err{ error: Error{ message: "exec: \"no-such-program\": executable file not found in $PATH" } }
```

## See also

[Cmd.wait](wait.md) · [Cmd.stop](stop.md) · [Cmd.kill](kill.md) · [Cmd.pid](pid.md) ·
[Cmd.run](run.md) · [Cmd.capture](capture.md) · [Cmd.pipe](pipe.md) ·
[Os.openPipe](../Os/openPipe.md) · [Os.stdout](../Os/stdout.md) · [Cmd.Process](types/Process.md)
