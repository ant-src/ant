<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.writeText

```ant signature
func File.writeText(f: File.Handle, text: string): Result[void, Error]
```

File.writeText writes text to f as UTF-8, without adding a newline.

<!-- end generated -->

## Details

- `File.writeText` adds nothing to `text`: no newline, no separator. Use
  [File.writeLine](writeLine.md) to end a line.
- The write reaches the OS before `File.writeText` returns, all of it or `Err`. Only a handle from
  [File.buffered](buffered.md) holds writes back, until its buffer is full, a flush, or a close.
- A handle that is not open for writing, such as one from [File.open](open.md) or one already
  closed, gives `Err`.

## Basic examples

Write text to a new file:

```ant
let f = try File.create("hello.txt")
try File.writeText(f, "Hello, Ant")
try f.close()
println(File.read("hello.txt"))
```

```text
Ok{ value: "Hello, Ant" }
```

Print a prompt on standard output and read the answer on the same line:

<!-- test: stdin
Ann
-->

```ant
try File.writeText(Os.stdout(), "Name: ")
let name = try Os.stdin().readLine() ?? "stranger"
println("Hello, ${name}")
```

```text
Name: Hello, Ann
```

## Scope

Several calls add to what the handle has written, with nothing between them. An empty string writes
nothing:

```ant
let f = try File.create("parts.txt")
try f.writeText("ant")
try f.writeText("")
try f.writeText("hill")
try f.close()
println(File.read("parts.txt"))
```

```text
Ok{ value: "anthill" }
```

Text with line breaks inside is written as it is:

```ant
let f = try File.create("poem.txt")
try f.writeText("Roses are red\nViolets are blue\n")
try f.close()
print(try File.read("poem.txt"))
```

```text
Roses are red
Violets are blue
```

Unicode text is written as UTF-8. `é` takes two bytes and `🐜` four:

```ant
let f = try File.create("ant.txt")
try f.writeText("é🐜")
try f.close()
println(File.readBytes("ant.txt"))
```

```text
Ok{ value: Bytes(c3 a9 f0 9f 90 9c) }
```

A handle from [File.append](append.md) writes after what the file already holds:

<!-- test: file log.txt
start;
-->

```ant
let f = try File.append("log.txt")
try f.writeText("more;")
try f.close()
print(try File.read("log.txt"))
```

```text
start;
more;
```

A connection is a handle too. This client sends a request in two pieces and reads the reply:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
go {
  let peer = Net.accept(listener).unwrap()
  defer { peer.close() }
  let line = peer.readLine().unwrap() ?? ""
  peer.writeText("got ${line}\n").unwrap()
}
let conn = try Net.dial(listener.address())
defer { conn.close() }
try conn.writeText("GET ")
try conn.writeText("/index\n")
println(try conn.readLine())
```

```text
got GET /index
```

## Applications

Show progress on one line, a dot per step, and end the line when the work is done:

```ant
let out = Os.stdout()
try out.writeText("Copying")
for _ in Iter.range(0, 3) {
  try out.writeText(".")
}
try out.writeLine(" done")
```

```text
Copying... done
```

Write a report to any handle, so the same function prints it or saves it:

```ant
type Item = { name: string, count: int }

func report(out: File.Handle, items: Array[Item]): Result[void, Error] {
  for item in items {
    try out.writeText("${item.name}: ${item.count}\n")
  }
  Result.ok(void)
}

let items = [Item{ name: "ant", count: 3 }, Item{ name: "bee", count: 12 }]
try report(Os.stdout(), items)
let f = try File.create("report.txt")
defer { f.close() }
try report(f, items)
try f.close()
println(File.read("report.txt"))
```

```text
ant: 3
bee: 12
Ok{ value: "ant: 3\nbee: 12\n" }
```

Write many small pieces through [File.buffered](buffered.md), so they reach the OS together at the
close:

```ant
let out = File.buffered(try File.create("squares.txt"))
for n in Iter.range(1, 6) {
  try out.writeText("${n * n} ")
}
try out.close()
println(File.read("squares.txt"))
```

```text
Ok{ value: "1 4 9 16 25 " }
```

## Properties and relations

`File.writeText` with a `"\n"` at the end writes what [File.writeLine](writeLine.md) writes:

```ant
let a = try File.create("a.txt")
try a.writeText("line\n")
try a.close()
let b = try File.create("b.txt")
try b.writeLine("line")
try b.close()
assert(try File.read("a.txt") == try File.read("b.txt"))
```

Writing pieces one after another gives the file [File.write](write.md) gives for the joined text:

```ant
let pieces = ["one", ",", "two"]
let f = try File.create("pieces.txt")
for piece in pieces {
  try f.writeText(piece)
}
try f.close()
try File.write("joined.txt", pieces | String.join(""))
assert(try File.read("pieces.txt") == try File.read("joined.txt"))
```

## Possible issues

`File.writeText` adds no line break, so the next output continues the same line:

```ant
try Os.stdout().writeText("first")
println("second")
```

```text
firstsecond
```

A handle from [File.open](open.md) is for reading, and a closed handle takes no more writes. Writing
to either is `Err`:

<!-- test: file notes.txt
notes
-->

```ant
let r = try File.open("notes.txt")
defer { r.close() }
println(r.writeText("more"))
let w = try File.create("closed.txt")
try w.close()
println(w.writeText("late"))
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
Err{ error: Error{ message: "the file is closed" } }
```

`File.writeText` takes a string. Write `Bytes` with [File.writeChunk](writeChunk.md):

```ant
let f = try File.create("data.bin")
try File.writeText(f, Bytes.fromArray([0xFF]))
// error: cannot use Bytes where string is expected
```

## See also

[File.writeLine](writeLine.md) · [File.writeChunk](writeChunk.md) · [File.write](write.md) ·
[File.create](create.md) · [File.append](append.md) · [File.buffered](buffered.md) ·
[File.close](close.md) · [Os.stdout](../Os/stdout.md)
