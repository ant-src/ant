<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.open

```ant signature
func File.open(path: string): Result[File.Handle, Error]
```

File.open opens the file at path for reading. Close it with [`File.close`](close.md). A directory
opens, and a read of it is Err.

<!-- end generated -->

## Details

- A write to the handle is `Err`; open a file for writing with [File.create](create.md) or
  [File.append](append.md).
- `File.open` never creates a file. A missing file is `Err`.

## Basic examples

Open a file and read it line by line, closing it when the script ends:

<!-- test: file names.txt
Ann
Bob
Cy
-->

```ant
let f = try File.open("names.txt")
defer { f.close() }
while true {
  let line = try f.readLine() ?? break
  println("Hello, ${line}")
}
```

```text
Hello, Ann
Hello, Bob
Hello, Cy
```

A missing file is `Err`:

```ant
println(File.open("missing.txt"))
```

```text
Err{ error: Error{ message: "open missing.txt: no such file or directory" } }
```

## Scope

Reads continue where the last one stopped, whatever their kind. Read a header line, then the rest:

<!-- test: file data.txt
id,name
1,ant
2,bee
-->

```ant
let f = try File.open("data.txt")
defer { f.close() }
println(try f.readLine())
print(try f.readAll())
```

```text
id,name
1,ant
2,bee
```

The bytes of a file come back as they are through [File.readChunk](readChunk.md):

<!-- test: file magic.bin
GIF89a
-->

```ant
let f = try File.open("magic.bin")
defer { f.close() }
println(try f.readChunk(3))
println(try f.readExactly(3))
```

```text
Bytes(47 49 46)
Bytes(38 39 61)
```

An empty file opens, and its first line read is `null`:

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

## Applications

Count the lines of a file without holding it in memory, and say what failed in the error:

<!-- test: file log.txt
start
work
stop
-->

```ant
func countLines(path: string): Result[int, Error] {
  let f = try (File.open(path) | Result.wrapErr("count lines"))
  defer { f.close() }
  var n = 0
  while (try f.readLine()) != null {
    n += 1
  }
  Result.ok(n)
}

println(countLines("log.txt"))
println(countLines("nope.txt"))
```

```text
Ok{ value: 3 }
Err{ error: Error{ message: "count lines: open nope.txt: no such file or directory" } }
```

Copy a file to standard output, as `cat` does, with [File.pipe](pipe.md):

<!-- test: file poem.txt
Roses are red
-->

```ant
let f = try File.open("poem.txt")
defer { f.close() }
try File.pipe(f, Os.stdout())
```

```text
Roses are red
```

Find the first line that matches, and stop reading there:

<!-- test: file config.txt
# settings
port=8080
host=local
-->

```ant
func findLine(path: string, prefix: string): Result[string | null, Error] {
  let f = try File.open(path)
  defer { f.close() }
  while true {
    let line = try f.readLine() ?? break
    if line.startsWith(prefix) { return Result.ok(line) }
  }
  Result.ok(null)
}

println(findLine("config.txt", "port="))
println(findLine("config.txt", "user="))
```

```text
Ok{ value: "port=8080" }
Ok{ value: null }
```

## Properties and relations

Reading a whole handle with [File.readAll](readAll.md) gives what [File.read](read.md) gives:

<!-- test: file text.txt
one
two
-->

```ant
let f = try File.open("text.txt")
defer { f.close() }
assert(try f.readAll() == try File.read("text.txt"))
```

Each `File.open` gives a new handle that reads from the start of the file:

<!-- test: file abc.txt
abc
-->

```ant
let a = try File.open("abc.txt")
defer { a.close() }
let b = try File.open("abc.txt")
defer { b.close() }
assert(a != b)
assert(try a.readChunk(1) == try b.readChunk(1))
```

## Possible issues

The handle cannot be written. Open a file for writing with [File.create](create.md) or
[File.append](append.md):

<!-- test: file notes.txt
draft
-->

```ant
let f = try File.open("notes.txt")
defer { f.close() }
println(f.writeLine("more"))
```

```text
Err{ error: Error{ message: "the file is not open for writing" } }
```

## See also

[File.close](close.md) · [File.read](read.md) · [File.readLine](readLine.md) ·
[File.readAll](readAll.md) · [File.create](create.md) · [File.append](append.md) ·
[File.lock](lock.md) · [File.list](list.md)
