<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.exit

```ant signature
func Os.exit(status: int): never
```

Os.exit ends the process at once with status. No deferred block runs, in any goroutine. A Unix
parent sees the status modulo 256, so Os.exit(256) reads as success; a portable program exits with 0
to 125.

<!-- end generated -->

## Details

- `Os.exit` never returns. Its type is `never`, so a call fits wherever a value is expected, such as
  after `??` or in one branch of an `if`.
- No deferred block runs, in the calling goroutine or any other. Clean up first, or return an `Err`
  from the script, which runs deferred blocks and exits with status `1`.
- The standard streams are not buffered, so what `print`, `println` and `eprintln` wrote before the
  call is out. A [`File.buffered`](../File/buffered.md) handle that is not flushed loses what its
  buffer holds.
- [`Panic.catch`](../Panic/catch.md) does not stop it.
- In `ant test`, it ends the whole command with its status, and no further test runs.

## Basic examples

`Os.exit(2)` reports a usage error after a message on standard error:

<!-- test: exit 2 -->

```ant
if Os.args().isEmpty() {
  eprintln("usage: greet <name>")
  Os.exit(2)
}
println("Hello, ${Os.args()[0]}")
```

Nothing after the call runs:

<!-- test: exit 3 -->

```ant
println("start")
Os.exit(3)
println("never printed")
```

```text
start
```

`Os.exit(0)` ends the program early with success:

<!-- test: args --version -->

```ant
if Os.args().contains("--version") {
  println("greet 1.0")
  Os.exit(0)
}
println("Hello")
```

```text
greet 1.0
```

## Scope

Because the type is `never`, `Os.exit` can supply the missing value after `??`:

<!-- test: exit 2 -->

```ant
let name = Os.args().get(0) ?? Os.exit(2)
println("Hello, ${name}")
```

It can also be one branch of an `if` expression, and the result keeps the other branch's type:

<!-- test: args 8080 -->

```ant
let port = if Os.args().length() == 1 { Os.args()[0] } else { Os.exit(2) }
println("listening on ${port}")
```

```text
listening on 8080
```

A function that ends in `Os.exit` on bad input needs no value for that path:

<!-- test: exit 2 -->

```ant
func portOf(s: string): int {
  s.toInt() ?? Os.exit(2)
}

println(portOf("80"))
println(portOf("http"))
```

```text
80
```

A call in any goroutine ends the whole process, the main goroutine included:

<!-- test: exit 7 -->

```ant
let stop: Channel[bool] = Channel.new()
go {
  println("worker exits")
  Os.exit(7)
}
Channel.receive(stop)
println("never printed")
```

```text
worker exits
```

## Applications

A filter in the style of `grep`: status `0` when some line matches, `1` when none does, and `2` for
a usage error:

<!-- test: args ant -->
<!-- test: stdin
ant colony
bee hive
ant hill
-->

```ant
let word = Os.args().get(0) ?? Os.exit(2)
let text = try Os.stdin().readAll()
let hits = text.split("\n").filter({ x.contains(word) })
if hits.isEmpty() {
  Os.exit(1)
}
for line in hits {
  println(line)
}
```

```text
ant colony
ant hill
```

The same program with no matching line exits with `1` and prints nothing:

<!-- test: args ant -->
<!-- test: stdin
bee hive
-->
<!-- test: exit 1 -->

```ant
let word = Os.args().get(0) ?? Os.exit(2)
let text = try Os.stdin().readAll()
let hits = text.split("\n").filter({ x.contains(word) })
if hits.isEmpty() {
  Os.exit(1)
}
for line in hits {
  println(line)
}
```

## Properties and relations

Returning an `Err` from a script also exits with status `1`, but it runs the deferred blocks first
and writes the error to standard error:

<!-- test: exit 1 -->

```ant
defer { println("cleanup") }
println("start")
return Result.fail("bad input")
```

```text
start
cleanup
```

`Os.exit(1)` at the same point skips the deferred block:

<!-- test: exit 1 -->

```ant
defer { println("cleanup") }
println("start")
Os.exit(1)
```

```text
start
```

`Panic.catch` catches panics only, so `Os.exit` passes through it:

<!-- test: exit 4 -->

```ant
let result = Panic.catch({ Os.exit(4) })
println("never printed: ${result}")
```

## Possible issues

Deferred blocks do not run, so a temporary directory removed by `defer` stays behind. Remove it
before the call:

<!-- test: exit 1 -->

```ant
let dir = try File.tempDir("job-*")
defer { File.removeAll(dir) }
if Os.args().isEmpty() {
  File.removeAll(dir)
  Os.exit(1)
}
```

The same holds in every goroutine: a worker's deferred block does not run when another goroutine
exits:

<!-- test: exit 5 -->

```ant
let started: Channel[bool] = Channel.new()
let stop: Channel[bool] = Channel.new()
go {
  defer { println("worker cleanup") }
  Channel.send(started, true)
  Channel.receive(stop)
}
Channel.receive(started)
println("exiting")
Os.exit(5)
```

```text
exiting
```

What a buffered handle holds is lost, because `Os.exit` does not flush it:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("lost")
Os.exit(0)
```

Flush it first:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("kept")
try out.flush()
Os.exit(0)
```

```text
kept
```

In a test, `Os.exit` ends `ant test` with its status, and the tests after it do not run:

<!-- test: cmd ant test . -->
<!-- test: exit 3 -->

```ant [exit_test.ant]
test "adds" {
  assert(1 + 1 == 2)
}

test "stops the run" {
  Os.exit(3)
}

test "never runs" {
  assert(false)
}
```

## See also

[Os.args](args.md) · [Os.interrupt](interrupt.md) · [Panic.catch](../Panic/catch.md) ·
[File.buffered](../File/buffered.md) · [panic](../prelude/panic.md) ·
[defer](../../language/keywords/defer.md)
