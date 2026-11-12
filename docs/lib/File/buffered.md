<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.buffered

```ant signature
func File.buffered(f: File.Handle, size: int = 4096): File.Handle
```

File.buffered returns a handle over f whose writes collect in a buffer of size bytes, written to f
when a write does not fit, at [`File.flush`](flush.md), and at [`File.close`](close.md), and at
nothing else: a buffered handle neither flushed nor closed when the run ends loses what it holds. It
is for a loop of many small writes. f itself is returned when it is buffered with at least size
bytes; a size below 1 gives the default. Reads of the handle read f. The first error writing the
buffer out is kept and returned by every later write and flush, and by the first close. A write to f
itself goes past the buffer.

<!-- end generated -->

## Details

- `File.close` of the buffered handle writes the buffer out and closes `f`. Check its result: a
  deferred close cannot report an error. Over a standard stream, `File.close` writes the buffer out
  and leaves the stream open.

## Basic examples

Buffer the lines written to a file, and close it to write them out:

```ant
func save(path: string, lines: Array[string]): Result[void, Error] {
  let out = File.buffered(try File.create(path))
  defer { out.close() }
  for line in lines { try out.writeLine(line) }
  out.close()
}

println(save("notes.txt", ["first", "second"]))
println(File.read("notes.txt"))
```

```text
Ok{ value: void }
Ok{ value: "first\nsecond\n" }
```

Buffer standard output for a loop of many small writes, and flush it at the end:

```ant
let out = File.buffered(Os.stdout())
for i in Iter.range(1, 4) {
  out.writeLine("line ${i}")
}
out.flush()
```

```text
line 1
line 2
line 3
```

## Scope

The file stays empty until the buffer is written out:

```ant
let out = File.buffered(Result.unwrap(File.create("log.txt")))
out.writeLine("started")
println(File.read("log.txt"))
out.flush()
println(File.read("log.txt"))
out.close()
```

```text
Ok{ value: "" }
Ok{ value: "started\n" }
```

Closing a buffered standard output writes it out and leaves standard output open:

```ant
let out = File.buffered(Os.stdout())
out.writeLine("buffered")
println(out.close())
println("stdout still open")
```

```text
buffered
Ok{ value: void }
stdout still open
```

Reads of a buffered handle read `f`, so one handle can serve both:

<!-- test: file names.txt
Ann
Bob
-->

```ant
let names = File.buffered(Result.unwrap(File.open("names.txt")))
defer { names.close() }
println(names.readLine())
println(names.readLine())
```

```text
Ok{ value: "Ann" }
Ok{ value: "Bob" }
```

Buffering a buffered handle again with no larger size returns the same handle:

```ant
let out = File.buffered(Result.unwrap(File.create("twice.txt")))
println(File.buffered(out) == out)
println(File.buffered(out, size: 8192) == out)
out.close()
```

```text
true
false
```

## Named arguments

`size` sets how many bytes the buffer holds. A write that does not fit writes the buffer out:

```ant
let out = File.buffered(Result.unwrap(File.create("small.txt")), size: 8)
out.writeText("abc")
println(File.read("small.txt"))
out.writeText("defghijk")
println(Result.unwrap(File.read("small.txt")).startsWith("abc"))
out.close()
println(File.read("small.txt"))
```

```text
Ok{ value: "" }
true
Ok{ value: "abcdefghijk" }
```

A `size` below 1 gives the default:

```ant
let out = File.buffered(Os.stdout(), size: 0)
out.writeText("held until flush\n")
out.flush()
```

```text
held until flush
```

## Applications

Write a CSV report row by row, and report a failure to finish it:

```ant
type Sale = { item: string, amount: int }

func writeReport(path: string, sales: Array[Sale]): Result[void, Error] {
  let out = File.buffered(try File.create(path))
  defer { out.close() }
  try out.writeLine("item,amount")
  for sale in sales { try out.writeLine("${sale.item},${sale.amount}") }
  out.close() | Result.wrapErr("write ${path}")
}

let sales = [Sale{ item: "tea", amount: 3 }, Sale{ item: "cake", amount: 5 }]
println(writeReport("sales.csv", sales))
println(File.read("sales.csv"))
```

```text
Ok{ value: void }
Ok{ value: "item,amount\ntea,3\ncake,5\n" }
```

Accept any handle and flush it without knowing whether it is buffered:

```ant
func greet(out: File.Handle, names: Array[string]): Result[void, Error] {
  for name in names { try out.writeLine("Hello, ${name}") }
  out.flush()
}

println(greet(Os.stdout(), ["Ann"]))
println(greet(File.buffered(Os.stdout()), ["Bob"]))
```

```text
Hello, Ann
Ok{ value: void }
Hello, Bob
Ok{ value: void }
```

## Properties and relations

`File.flush` and `File.close` on the buffered handle are `Ok`; `File.flush` on a handle that is not
buffered does nothing and is `Ok`:

```ant
let f = Result.unwrap(File.create("props.txt"))
let out = File.buffered(f)
assert(out.writeText("x") is Ok)
assert(f.flush() is Ok)
assert(Result.unwrap(File.read("props.txt")) == "")
assert(out.flush() is Ok)
assert(Result.unwrap(File.read("props.txt")) == "x")
assert(out.close() is Ok)
assert(out.close() is Ok)
```

## Possible issues

A buffered handle neither flushed nor closed loses what it holds when the program ends. This program
prints nothing:

```ant
let out = File.buffered(Os.stdout())
out.writeLine("lost")
```

A `println` goes past the buffer of a buffered standard output, so it comes out first:

```ant
let out = File.buffered(Os.stdout())
out.writeLine("written first")
println("printed second")
out.flush()
```

```text
printed second
written first
```

A write after the close is `Err`, and a second close does nothing:

```ant
let out = File.buffered(Result.unwrap(File.create("closed.txt")))
println(out.close())
println(out.writeLine("late"))
println(out.close())
```

```text
Ok{ value: void }
Err{ error: Error{ message: "the file is closed" } }
Ok{ value: void }
```

## See also

[File.flush](flush.md) · [File.close](close.md) · [File.create](create.md) ·
[File.writeLine](writeLine.md) · [File.writeText](writeText.md) · [Os.stdout](../Os/stdout.md)
