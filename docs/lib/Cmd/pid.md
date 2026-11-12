<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.pid

```ant signature
func Cmd.pid(p: Cmd.Process): int
```

Cmd.pid returns the process id p's last stage started with, the same after it exits; once it has
been reaped the system may give that id to another process, so a program does not signal it by id.

<!-- end generated -->

## Details

- For a pipeline it is the last stage's id, as a shell's `$!` is. The other stages have ids of their
  own, which `Cmd.pid` does not give.

## Basic examples

Start a command in the background and print its process id:

<!-- test: check-only the id changes from run to run -->

```ant
let child = try Cmd.new("sleep", ["5"]).start()
defer { child.stop() }
println(Cmd.pid(child))
```

The id is a positive number, and not this program's own:

```ant
let child = try Cmd.new("sleep", ["5"]).start()
defer { child.stop() }
assert(child.pid() > 0)
assert(child.pid() != Os.pid())
```

## Scope

### A pipeline

The id of a started pipeline is its last stage's. Here each stage prints its own id, `$$`, and the
last stage's output reaches the pipe:

```ant
let p = try Os.openPipe()
defer { p.read.close() }
let pipeline = try Cmd.pipe([
  Cmd.new("sh", ["-c", "echo $$"]),
  Cmd.new("sh", ["-c", "cat >/dev/null; echo $$"]),
]).start(stdout: p.write)
try p.write.close()
let reported = try p.read.readAll()
try pipeline.wait()
println(reported.trim() == "${pipeline.pid()}")
```

```text
true
```

## Applications

### Log what runs where

Name the command and its id when it starts, so a log line can be matched with `ps` or `top`:

<!-- test: check-only the id changes from run to run -->

```ant
let server = Cmd.new("sleep", ["30"])
let process = try server.start()
defer { process.stop() }
println("started ${server.string()} as ${process.pid()}")
```

### Write a pid file

Another tool that only has a pid file can then find the child:

```ant
let server = try Cmd.new("sleep", ["30"]).start()
defer { server.stop() }
let dir = try File.tempDir("app")
defer { File.removeAll(dir) }
let pidFile = Path.join([dir, "server.pid"])
try File.write(pidFile, "${server.pid()}\n")
let saved = try File.read(pidFile)
assert(saved.trim().toInt() == server.pid())
```

### Ask another program about the child

Pass the id to a program that looks processes up, such as `ps`:

```ant
let child = try Cmd.new("sleep", ["30"]).start()
defer { child.stop() }
let name = try Cmd.new("ps", ["-p", "${child.pid()}", "-o", "comm="]).read()
println(name)
```

```text
sleep
```

## Properties and relations

Two children running at once have different ids:

```ant
let a = try Cmd.new("sleep", ["5"]).start()
defer { a.stop() }
let b = try Cmd.new("sleep", ["5"]).start()
defer { b.stop() }
assert(a.pid() != b.pid())
```

## Possible issues

### A signal to the id reaches one stage

Signaling the id with another program, such as `kill`, ends only the last stage of a pipeline. The
first stage goes on running, so a wait with a timeout runs out. `Cmd.kill` and `Cmd.stop` signal
every stage:

```ant
let pipeline = try Cmd.pipe([Cmd.new("sleep", ["30"]), Cmd.new("sleep", ["30"])]).start()
try Cmd.new("kill", ["${pipeline.pid()}"]).run()
let ctx = Context.new(timeout: Time.duration(milliseconds: 200))
println(pipeline.wait(ctx: ctx))
try pipeline.kill()
println(pipeline.wait())
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: 143 }
```

### A command is not a process

`Cmd.pid` takes the process that `Cmd.start` returns, unwrapped from its `Result`. A `Cmd` has no id
until it is started:

```ant
let sleeper = Cmd.new("sleep", ["5"])
println(sleeper.pid()) // error: Cmd.pid takes Cmd.Process first, so a Cmd does not reach it
let started = Cmd.new("sleep", ["5"]).start()
println(started.pid()) // error: Result[Cmd.Process, Error] has no field or function pid
```

## See also

[Cmd.start](start.md) · [Cmd.kill](kill.md) · [Cmd.stop](stop.md) · [Cmd.wait](wait.md) ·
[Os.pid](../Os/pid.md) · [Cmd](index.md)
