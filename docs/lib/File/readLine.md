<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readLine

```ant signature
func File.readLine(f: File.Handle, ctx: Context | null = null): Result[string | null, Error]
```

File.readLine reads the next line from f as text, without its \\n or \\r\\n, and returns null at end
of file. When ctx is done first it is Err with ctx's reason, and what it had read stays for the next
read of f.

<!-- end generated -->

## Details

- The value is `Ok(null)` only at the end of the input. An empty line is `Ok("")`, and a last line
  without a line break is returned whole.
- `readLine` decodes what it reads as text: bytes that are not UTF-8 become U+FFFD. Read binary data
  with [File.readChunk](readChunk.md) or [File.readExactly](readExactly.md).
- All reads of a handle share one buffer, so `readLine` can be followed by any other read, which
  continues after the line.
- A context interrupts a read that waits: a pipe, a connection, a terminal. A regular file never
  waits, so there the context is looked at only before the read begins.

## Basic examples

Read a file line by line until `File.readLine` returns `null`:

<!-- test: file names.txt
Ann
Bob
Cy
-->

```ant
let f = try File.open("names.txt")
defer { f.close() }
while true {
  let line = try File.readLine(f) ?? break
  println("Hello, ${line}")
}
```

```text
Hello, Ann
Hello, Bob
Hello, Cy
```

Read one line of standard input:

<!-- test: stdin
Ann
-->

```ant
print("Name: ")
let name = try File.readLine(Os.stdin()) ?? "stranger"
println("Hello, ${name}")
```

```text
Name: Hello, Ann
```

## Scope

Each call returns the next line. `\r\n` ends a line as `\n` does, an empty line is `""`, a last line
without a line break comes back whole, and the end of the file is `null`:

```ant
try File.write("lines.txt", "one\r\ntwo\n\nlast")
let f = try File.open("lines.txt")
defer { f.close() }
for _ in Iter.range(0, 5) {
  println(f.readLine())
}
```

```text
Ok{ value: "one" }
Ok{ value: "two" }
Ok{ value: "" }
Ok{ value: "last" }
Ok{ value: null }
```

An empty file has no lines:

<!-- test: file empty.txt
-->

```ant
let f = try File.open("empty.txt")
defer { f.close() }
println(f.readLine())
```

```text
Ok{ value: null }
```

Bytes that are not UTF-8 become U+FFFD:

```ant
try File.writeBytes("latin1.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xE9, 0x0A]))
let f = try File.open("latin1.txt")
defer { f.close() }
println(try f.readLine())
```

```text
caf�
```

A connection is a handle too, so a line protocol is a loop over `readLine`. This server answers each
line in upper case:

```ant
func serve(conn: File.Handle): Result[void, Error] {
  defer { conn.close() }
  while true {
    let line = try conn.readLine() ?? break
    try conn.writeLine(line.toUpper())
  }
  Result.ok(void)
}

let listener = try Net.listen("127.0.0.1:0")
go { Result.unwrap(serve(Result.unwrap(Net.accept(listener)))) }
let conn = try Net.dial(listener.address())
defer { conn.close() }
for word in ["ping", "hello"] {
  try conn.writeLine(word)
  println(try conn.readLine())
}
```

```text
PING
HELLO
```

## Named arguments

`ctx` gives up on a read that waits. Here the line is not complete when the timeout passes; the read
is `Err`, and the part it had read stays for the next read:

```ant
let pipe = try Os.openPipe()
defer { pipe.read.close() }
try pipe.write.writeText("par")
let ctx = Context.new(timeout: 50ms)
println(pipe.read.readLine(ctx: ctx))
try pipe.write.writeLine("tial")
try pipe.write.close()
println(pipe.read.readLine())
```

```text
Err{ error: Error{ message: "timeout" } }
Ok{ value: "partial" }
```

## Applications

Sum the numbers on standard input, one per line, and name the line that is not a number:

<!-- test: stdin
3
4
x
-->

```ant
func sum(input: File.Handle): Result[int, Error] {
  var total = 0
  var n = 0
  while true {
    let line = try input.readLine() ?? break
    n += 1
    let value = line.trim().toInt() ?? return Result.fail("line ${n}: not a number: ${line}")
    total += value
  }
  Result.ok(total)
}

println(sum(Os.stdin()))
```

```text
Err{ error: Error{ message: "line 3: not a number: x" } }
```

Read `key=value` settings into a map, skipping blank lines and comments:

<!-- test: file app.conf
# server
port=8080

host=localhost
-->

```ant
func readConfig(path: string): Result[Map[string, string], Error] {
  let f = try (File.open(path) | Result.wrapErr("config"))
  defer { f.close() }
  let settings: Map[string, string] = Map{}
  while true {
    let line = (try f.readLine() ?? break).trim()
    if line.isEmpty() || line.startsWith("#") { continue }
    let pair = line.cut("=") ?? continue
    settings[pair.before] = pair.after
  }
  Result.ok(settings)
}

println(readConfig("app.conf"))
```

```text
Ok{ value: Map{ "port": "8080", "host": "localhost" } }
```

Read a header line, then the rest of the file in one piece with [File.readAll](readAll.md):

<!-- test: file report.csv
name,score
Ann,9
Bob,7
-->

```ant
let f = try File.open("report.csv")
defer { f.close() }
let header = try f.readLine() ?? ""
println(header.split(","))
print(try f.readAll())
```

```text
["name", "score"]
Ann,9
Bob,7
```

## Properties and relations

Reading every line gives what [String.lines](../String/lines.md) gives for the whole file:

<!-- test: file poem.txt
Roses are red

Violets are blue
-->

```ant
let f = try File.open("poem.txt")
defer { f.close() }
let lines: Array[string] = []
while true {
  lines.push(try f.readLine() ?? break)
}
assert(lines == (try File.read("poem.txt")).lines())
```

## Possible issues

The line is `string | null`. Handle the end of the file before you use it:

<!-- test: file words.txt
ant
-->

```ant
let f = try File.open("words.txt")
let line = try f.readLine()
println(line.trim())
// error: string | null has no field or function trim on every member: narrow the value first
```

A context that is already done stops a read of a regular file before it begins:

<!-- test: file data.txt
data
-->

```ant
let f = try File.open("data.txt")
defer { f.close() }
let ctx = Context.new()
ctx.cancel()
println(f.readLine(ctx: ctx))
println(f.readLine())
```

```text
Err{ error: Error{ message: "cancelled" } }
Ok{ value: "data" }
```

A handle opened for writing cannot be read:

```ant
let out = try File.create("out.txt")
defer { out.close() }
println(out.readLine())
```

```text
Err{ error: Error{ message: "the file is not open for reading" } }
```

## See also

[File.open](open.md) · [File.readAll](readAll.md) · [File.read](read.md) ·
[File.readChunk](readChunk.md) · [File.writeLine](writeLine.md) · [String.lines](../String/lines.md)
· [Os.stdin](../Os/stdin.md)
