<!-- status: reviewed -->

# Stream a command's output

This program runs a build and prints each line of its output as the build writes it, instead of
waiting for the build to end:

```ant
let build = Cmd.new("sh", ["-c", "echo compiling; sleep 0.1; echo linking; echo done"])
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(build, stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println("[build] ${line}")
}
let status = try child.wait()
if status != 0 {
  eprintln("${build.string()} failed with status ${status}")
  Os.exit(status)
}
println("build ok")
```

```text
[build] compiling
[build] linking
[build] done
build ok
```

1. Describe the command with [Cmd.new](../../lib/Cmd/new.md).
2. Open a pipe with [Os.openPipe](../../lib/Os/openPipe.md), and close its read end with `defer`.
3. Start the command with [Cmd.start](../../lib/Cmd/start.md), giving the pipe's write end as
   `stdout:`. The call returns once the command is running.
4. Close your own copy of the write end with [File.close](../../lib/File/close.md). Until you do,
   the read end never reaches the end of the output.
5. Read the output with [File.readLine](../../lib/File/readLine.md) in a loop. Each call waits for
   the next line and returns `null` once the command has closed its output; end the loop there with
   `?? break`. Print each line with [println](../../lib/prelude/println.md).
6. Get the exit status with [Cmd.wait](../../lib/Cmd/wait.md). A nonzero status is `Ok`, not an
   error, so check it yourself. Here the script names the command with
   [Cmd.string](../../lib/Cmd/string.md), reports it with [eprintln](../../lib/prelude/eprintln.md),
   and exits with the same status through [Os.exit](../../lib/Os/exit.md).

Read the pipe before you wait. A command that writes into a pipe nobody reads blocks once the pipe
is full, and the wait then never ends.

## Merge standard error into the stream

Give the same write end as `stdout:` and `stderr:` to read both outputs as one stream, as `2>&1`
does in a shell:

```ant
let test = Cmd.new("sh", ["-c", "echo ok 1; echo 'not ok 2' >&2; exit 1"])
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(test, stdout: out.write, stderr: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println("| ${line}")
}
println("status ${try child.wait()}")
```

```text
| ok 1
| not ok 2
status 1
```

## Read standard error separately

Give each output its own pipe to tell them apart. Read one of them on another goroutine, so that
neither pipe fills while you read the other. Here a goroutine reads all of standard error with
[File.readAll](../../lib/File/readAll.md) into an array slot, and a
[WaitGroup](../../lib/WaitGroup/new.md) tells the script when it is done. The goroutine writes into
an array slot, not a local, since a local assigned inside `go` cannot hand a result back; the slot
starts as [Result.ok](../../lib/Result/ok.md) of an empty string:

```ant
let job = Cmd.new("sh", ["-c", "echo step 1; echo 'warning: slow disk' >&2; echo step 2"])
let out = try Os.openPipe()
defer { out.read.close() }
let errs = try Os.openPipe()
defer { errs.read.close() }
let child = try Cmd.start(job, stdout: out.write, stderr: errs.write)
try out.write.close()
try errs.write.close()
let errText: Array[Result[string, Error]] = [Result.ok("")]
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  errText[0] = errs.read.readAll()
}
while true {
  let line = try out.read.readLine() ?? break
  println("out: ${line}")
}
wg.wait()
let stderr = try errText[0]
println("stderr: ${stderr.trim()}")
println("status ${try child.wait()}")
```

```text
out: step 1
out: step 2
stderr: warning: slow disk
status 0
```

Close both write ends after [Cmd.start](../../lib/Cmd/start.md), or one of the reads never ends.
[WaitGroup.add](../../lib/WaitGroup/add.md), [WaitGroup.done](../../lib/WaitGroup/done.md) and
[WaitGroup.wait](../../lib/WaitGroup/wait.md) join the goroutine, and
[String.trim](../../lib/String/trim.md) drops the final line break.

## Stop the command at a line

Stop reading at the line you wait for, such as a server's ready message, and end the command with
[Cmd.stop](../../lib/Cmd/stop.md). It sends SIGTERM to the command's process group, then SIGKILL if
it still runs after `grace:`, five seconds by default, and returns the status: here 143, which is
128 plus SIGTERM's number, 15. On Windows the status is 1.

```ant
let server = Cmd.new("sh", ["-c", "echo starting; echo ready on port 8080; sleep 10"])
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(server, stdout: out.write)
try out.write.close()
while true {
  let line = try out.read.readLine() ?? break
  println(line)
  if line.startsWith("ready") { break }
}
println("stopped with status ${try child.stop(grace: 1s)}")
```

```text
starting
ready on port 8080
stopped with status 143
```

[String.startsWith](../../lib/String/startsWith.md) tests each line.

A server that never prints the line keeps the loop waiting. To put a time limit on the wait, create
a context with [Context.new](../../lib/Context/new.md) before the loop and give it to each read as
`ctx:`. When the time runs out, the read returns `Err` and `try` ends the script, and a deferred
`child.stop()` still ends the command:

<!-- test: exit 1 -->

```ant
let server = Cmd.new("sh", ["-c", "echo starting; sleep 10"])
let out = try Os.openPipe()
defer { out.read.close() }
let child = try Cmd.start(server, stdout: out.write)
defer { child.stop(grace: 1s) }
try out.write.close()
let ready = Context.new(1s)
while true {
  let line = try out.read.readLine(ctx: ready) ?? break
  println(line)
  if line.startsWith("ready") { break }
}
println("server is ready")
```

```text
starting
```

To end a command after a time limit instead, see [Stop or time out a process](stop.md).

## Stream the output into a file

To send the output straight to a file as the command writes it, give an open file as `stdout:`. Open
it with [File.create](../../lib/File/create.md); the command writes into it until it exits:

```ant
let log = try File.create("build.log")
defer { log.close() }
let child = try Cmd.start(Cmd.new("sh", ["-c", "echo compiling; echo linking"]), stdout: log)
println("status ${try child.wait()}")
print(try File.read("build.log"))
```

```text
status 0
compiling
linking
```

[File.read](../../lib/File/read.md) and [print](../../lib/prelude/print.md) show the log once the
command has exited. To feed one command's output into another, see [Build a pipeline](pipeline.md).
