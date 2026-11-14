<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os

Os holds the process: its arguments, environment, standard streams, exit, interrupts, and pipes.
[`Cmd`](../Cmd/index.md) runs other programs.

<!-- end generated -->

## Details

- The functions that can fail, such as [Os.cwd](cwd.md) and [Os.home](home.md), return a `Result`.
  The others, such as [Os.args](args.md) and [Os.env](env.md), never fail: a missing variable is
  `null`, not an error.
- The three streams are [File.Handle](../File/types/Handle.md) values, so the `File` functions read
  and write them. They are not buffered: `print`, `eprint`, the handles, and the program's children
  write in the order they are called. Wrap a stream in [File.buffered](../File/buffered.md) to write
  much. When the reader of standard output or standard error goes away, as `| head` does, the next
  write ends the program, whichever function made it, and the shell reports status 141.
- [Os.exit](exit.md) ends the process at once, and no deferred block runs, in any goroutine.
- A program ends at the first interrupt (Ctrl-C, SIGINT, or SIGTERM where the host has it) until it
  calls [Os.interrupt](interrupt.md). After that call, the first interrupt makes the context done
  and the program finishes on its own. A second interrupt ends the run, and so does the first one
  after the program cancels the context itself.

## Basic examples

`Os.args` gives the arguments after the program path. `Os.program` names the script for a usage
line:

<!-- test: args Ann Bob -->

```ant
let names = Os.args()
if names.isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} <name>...")
  Os.exit(2)
}
for name in names {
  println("Hello, ${name}")
}
```

```text
Hello, Ann
Hello, Bob
```

`Os.env` gives a variable's value, or `null` when it is not set, so `??` supplies a default:

```ant
let port = Os.env("APP_NO_SUCH_PORT") ?? "8080"
println("listening on ${port}")
```

```text
listening on 8080
```

`Os.stdin`, `Os.stdout`, and `Os.stderr` are handles:

<!-- test: stdin
ant
go
-->

```ant
let input = Os.stdin()
while true {
  let line = try input.readLine() ?? break
  try Os.stdout().writeLine(line.toUpper())
}
try Os.stderr().writeLine("done")
```

```text
ANT
GO
```

## Scope

The streams are not buffered, so a handle's writes and `println` come out in the order they are
called:

```ant
let out = Os.stdout()
try out.writeText("one ")
print("two ")
try out.writeLine("three")
```

```text
one two three
```

[Os.environ](environ.md) gives every variable as a map:

```ant
let env = Os.environ()
println(env.has("PATH"))
println(env.get("APP_NO_SUCH_VARIABLE"))
```

```text
true
null
```

[Os.home](home.md), [Os.configDir](configDir.md), and [Os.cacheDir](cacheDir.md) return a `Result`:
each fails when the variable it rests on is unset, or when an XDG variable holds a relative path.
Keep a script's files under a directory named after the script:

<!-- test: check-only the home directory differs between machines -->

```ant
let config = Path.join([try Os.configDir(), "greet", "config.json"])
println(config)
```

[Os.cwd](cwd.md) is the working directory, an absolute path:

```ant
let dir = try Os.cwd()
println(Path.isAbs(dir))
```

```text
true
```

[Os.uid](uid.md) is the effective user id, so `Os.uid() == 0` asks whether the program runs as root.
On Windows it is `null`, and the test is false:

<!-- test: check-only the user differs between machines -->

```ant
if Os.uid() == 0 {
  eprintln("do not run as root")
}
```

[Os.platform](platform.md) and [Os.arch](arch.md) use Go's names, so macOS is `"darwin"`:

<!-- test: check-only the platform differs between machines -->

```ant
println("${Os.platform()}/${Os.arch()}")
```

[Os.which](which.md) finds the program a `Cmd` would start, or gives `null`:

```ant
println(Os.which("no-such-program-ant") == null)
println(Os.which("sh") != null)
```

```text
true
true
```

[Os.openPipe](openPipe.md) gives two connected handles. Close the writing end, and the reading end
is at end of file after what was written:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
try pipe.write.writeLine("through the pipe")
try pipe.write.close()
println(try pipe.read.readLine())
println(try pipe.read.readLine())
```

```text
through the pipe
null
```

[Os.pid](pid.md) is this process's id, so the program can signal itself. After a call to
`Os.interrupt`, a SIGINT makes the context done instead of ending the run:

```ant
let stop = Os.interrupt()
try Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"]))
stop.wait()
println("interrupted: ${stop.reason()}")
```

```text
interrupted: interrupt
```

## Functions by job

### Types

- [Os.Pipe](types/Pipe.md): the two handles of a pipe

### Arguments and identity

- [Os.args](args.md): the arguments after the program path
- [Os.program](program.md): the program path, as the command gave it
- [Os.pid](pid.md): this process's id
- [Os.uid](uid.md): the effective user id, or `null` on Windows

### Environment

- [Os.env](env.md): one variable, or `null` when it is not set
- [Os.environ](environ.md): every variable, as a map

### Host and directories

- [Os.platform](platform.md), [Os.arch](arch.md): the operating system and processor, by Go's names
- [Os.hostname](hostname.md): the host's name
- [Os.cwd](cwd.md): the working directory
- [Os.home](home.md): the user's home directory
- [Os.configDir](configDir.md), [Os.cacheDir](cacheDir.md): where programs keep configuration and
  what can be made again
- [Os.which](which.md): the path of the program a `Cmd` would start

### Streams and pipes

- [Os.stdin](stdin.md), [Os.stdout](stdout.md), [Os.stderr](stderr.md): the standard streams, as
  handles
- [Os.openPipe](openPipe.md): two connected handles

### Ending the process

- [Os.exit](exit.md): end the process now with a status
- [Os.interrupt](interrupt.md): a context done at Ctrl-C, SIGINT, or SIGTERM

## Applications

A filter in the Unix style reads standard input and writes standard output through a buffer, which
the final close writes out:

<!-- test: stdin
apples
pears
-->

```ant
let out = File.buffered(Os.stdout())
defer { out.close() }
let input = Os.stdin()
var n = 0
while true {
  let line = try input.readLine() ?? break
  n += 1
  try out.writeLine("${n}\t${line}")
}
try out.close()
```

```text
1	apples
2	pears
```

Pick a command by platform rather than running `uname`:

<!-- test: check-only opening a file depends on the desktop -->

```ant
let opener = if Os.platform() == "darwin" { "open" } else { "xdg-open" }
try Cmd.run(Cmd.new(opener, ["report.html"]))
```

Read a setting from the environment and reject a bad value with a usage error:

<!-- test: exit 2 -->

```ant
let text = Os.env("APP_NO_SUCH_WORKERS") ?? "zero"
let workers = text.toInt()
if workers == null {
  eprintln("APP_NO_SUCH_WORKERS must be a number, not ${text}")
  Os.exit(2)
}
println("${workers} workers")
```

Work until Ctrl-C, then finish what is in hand. Passing the context to
[Time.sleep](../Time/sleep.md) ends the pause at once:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
let stop = Os.interrupt()
var rounds = 0
while !stop.isDone() {
  rounds += 1
  Time.sleep(10s, ctx: stop)
}
println("saving after ${rounds} round")
```

```text
saving after 1 round
```

## Properties and relations

`Os.env` agrees with `Os.environ`, and `Os.interrupt` gives the same context at every call:

```ant
assert(Os.env("PATH") == Os.environ().get("PATH"))
assert(Os.env("APP_NO_SUCH_VARIABLE") == null)
assert(Os.interrupt() == Os.interrupt())
```

## Possible issues

A directory is a `Result`, not a string. Use `try` first:

```ant
let path = Path.join([Os.home(), ".greetrc"])
// error: cannot use Result[string, Error] where string is expected
```

`Os.args` leaves out the program path, so the first argument is at index 0, not 1:

<!-- test: args input.txt -->

```ant
println(Os.args())
println(Os.args().get(1))
```

```text
["input.txt"]
null
```

`Os.exit` runs no deferred block, so a deferred close of a buffered stream never writes it. Close
the buffer before the call:

<!-- test: exit 1 -->

```ant
let out = File.buffered(Os.stdout())
defer { out.close() }
try out.writeLine("lost")
Os.exit(1)
```

## See also

[Cmd](../Cmd/index.md) · [File](../File/index.md) · [Path](../Path/index.md) ·
[Context](../Context/index.md) · [print](../prelude/print.md) ·
[Read command-line arguments](../../howto/cli/args.md) ·
[Read environment variables and standard directories](../../howto/cli/env.md) ·
[Handle Ctrl-C](../../howto/cli/interrupt.md)
