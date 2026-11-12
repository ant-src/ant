<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.flush

```ant signature
func File.flush(f: File.Handle): Result[void, Error]
```

File.flush writes out the buffer of a buffered handle, and of every buffered handle under it, so
that what was written to f has been handed to the OS. On a handle that is not buffered it does
nothing and is Ok.

<!-- end generated -->

## Details

- Only a handle from [File.buffered](buffered.md) holds writes back. A write to any other handle
  reaches the OS before it returns, so nothing needs a flush.
- A buffer is written out when a write does not fit, at `File.flush`, and at [File.close](close.md),
  and at nothing else: not at a read, a wait, or the end of the program.
- Over buffered handles stacked on each other, `File.flush` writes out every layer, the outer one
  first.
- The first error from writing a buffer out is kept, and every later write and flush returns it.

## Basic examples

Lines written to a buffered handle stay in its buffer until `File.flush` writes them out:

```ant
let out = File.buffered(try File.create("log.txt"))
defer { File.close(out) }
try out.writeLine("started")
println("before: ${try File.read("log.txt") | String.length()} bytes")
try File.flush(out)
print(try File.read("log.txt"))
```

```text
before: 0 bytes
started
```

Flush buffered standard output with `.`:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("one")
try out.writeLine("two")
try out.flush()
```

```text
one
two
```

## Scope

On a handle that is not buffered, such as a file from [File.create](create.md) or standard output,
`File.flush` does nothing and is `Ok`:

```ant
let f = try File.create("plain.txt")
defer { File.close(f) }
println(f.flush())
println(Os.stdout().flush())
```

```text
Ok{ value: void }
Ok{ value: void }
```

A buffered handle over another buffered handle with a smaller buffer is a second layer. One flush of
the outer handle writes out both:

```ant
let inner = File.buffered(try File.create("layers.txt"), size: 16)
let outer = File.buffered(inner, size: 4096)
defer { File.close(outer) }
try outer.writeLine("through both buffers")
try outer.flush()
print(try File.read("layers.txt"))
```

```text
through both buffers
```

A buffered connection from [Net.dial](../Net/dial.md) sends nothing until it is flushed. Flush the
request, then read the reply through the same handle:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { Net.close(listener) }
go {
  let peer = Net.accept(listener).unwrap()
  defer { File.close(peer) }
  let name = peer.readLine().unwrap() ?? ""
  peer.writeLine("hello, ${name}").unwrap()
}
let conn = File.buffered(try Net.dial(listener.address()))
defer { File.close(conn) }
try conn.writeLine("ant")
try conn.flush()
println(try conn.readLine())
```

```text
hello, ant
```

## Applications

A function that reports progress flushes after each line, so the line shows up at once whether the
caller passes a buffered handle or not:

```ant
func progress(f: File.Handle, step: int): Result[void, Error] {
  try f.writeLine("step ${step} done")
  f.flush()
}

try progress(Os.stdout(), 1)
let out = File.buffered(Os.stdout())
try progress(out, 2)
```

```text
step 1 done
step 2 done
```

Write a table in many small pieces through one buffer, and flush once at the end:

```ant
let stock = Map{ "ant": 3, "bee": 12 }
let out = File.buffered(Os.stdout())
for { key, value } in stock {
  try out.writeText(key)
  try out.writeText(": ")
  try out.writeLine("${value}")
}
try out.flush()
```

```text
ant: 3
bee: 12
```

## Properties and relations

After a flush, the file holds everything written to the buffered handle. A second flush writes
nothing more:

```ant
let out = File.buffered(try File.create("twice.txt"))
defer { File.close(out) }
try out.writeText("abc")
try out.flush()
assert(try File.read("twice.txt") == "abc")
try out.flush()
assert(try File.read("twice.txt") == "abc")
```

[File.close](close.md) of a buffered handle flushes it too, so a flush right before a close is not
needed:

```ant
let out = File.buffered(try File.create("closed.txt"))
try out.writeText("kept")
try File.close(out)
assert(try File.read("closed.txt") == "kept")
```

## Possible issues

A buffered handle that is neither flushed nor closed when the program ends loses what its buffer
holds. This script writes nothing to `report.txt`:

```ant
let out = File.buffered(try File.create("report.txt"))
for row in ["ant", "bee", "cat"] {
  try out.writeLine(row)
}
println(try File.read("report.txt") | String.length())
```

```text
0
```

A `print` or `println` goes past the buffer of `File.buffered(Os.stdout())`, so it can come out
before what the buffer still holds. Flush before printing another way:

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("from the buffer")
println("from println")
try out.flush()
```

```text
from println
from the buffer
```

A write to the handle under the buffer goes past the buffer too, and lands before what the buffer
holds:

```ant
let f = try File.create("mixed.txt")
let out = File.buffered(f)
defer { File.close(out) }
try out.writeLine("through the buffer")
try f.writeLine("straight to the file")
try out.flush()
print(try File.read("mixed.txt"))
```

```text
straight to the file
through the buffer
```

Flush `out` before writing to `f` directly.

## See also

[File.buffered](buffered.md) · [File.close](close.md) · [File.writeLine](writeLine.md) ·
[File.writeText](writeText.md) · [Os.stdout](../Os/stdout.md)
