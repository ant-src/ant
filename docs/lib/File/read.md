<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.read

```ant signature
func File.read(path: string): Result[string, Error]
```

File.read reads the whole file at path as text. Bytes that are not valid UTF-8 become U+FFFD;
[`File.readBytes`](readBytes.md) keeps them as they are.

<!-- end generated -->

## Details

- `File.read` reads the whole file into one string. For a file too large to hold at once, or a
  stream, open it with [File.open](open.md) and read it with [File.readLine](readLine.md) or
  [File.readChunk](readChunk.md).
- A missing file, a directory, or a file the program may not read is `Err`, and the message names
  the path.
- To find out whether a file is text, read it with [File.readBytes](readBytes.md) and check it with
  [String.fromBytes](../String/fromBytes.md).
- A byte order mark at the start is kept, as `"\u{FEFF}"`.
- A symbolic link is followed, and the text is its target's.

## Basic examples

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
println(File.read("notes.txt"))
println(File.read("missing.txt"))
```

```text
Ok{ value: "buy milk\ncall Ann\n" }
Err{ error: Error{ message: "open missing.txt: no such file or directory" } }
```

Inside a function that returns a `Result`, `try` gives the text and returns the `Err` early:

<!-- test: file notes.txt
buy milk
call Ann
-->

```ant
func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.lines().length())
}

println(countLines("notes.txt"))
println(countLines("missing.txt"))
```

```text
Ok{ value: 2 }
Err{ error: Error{ message: "open missing.txt: no such file or directory" } }
```

## Scope

### An empty file

An empty file is `Ok` with the empty string:

```ant
try File.write("empty.txt", "")
println(File.read("empty.txt"))
```

```text
Ok{ value: "" }
```

### Unicode

The text comes back as written, whatever the script:

```ant
try File.write("greeting.txt", "Привет, 世界 👋")
let text = try File.read("greeting.txt")
println(text)
println(text.length())
```

```text
Привет, 世界 👋
12
```

### Bytes that are not UTF-8

Each ill-formed sequence becomes U+FFFD. Here a Latin-1 `é` (byte `E9`) and a euro sign cut short
(`E2 82`) become one replacement character each:

```ant
try File.writeBytes("latin1.txt", Bytes.fromArray([0x63, 0x61, 0x66, 0xE9, 0x20, 0xE2, 0x82]))
let text = try File.read("latin1.txt")
println(text)
println(text.count("\u{FFFD}"))
println(String.fromBytes(try File.readBytes("latin1.txt")))
```

```text
caf� �
2
null
```

### A directory

A directory is not a file to read:

```ant
try File.mkdir("logs")
println(File.read("logs"))
```

```text
Err{ error: Error{ message: "read logs: is a directory" } }
```

### A symbolic link

Reading through a link reads the file it points at:

<!-- test: file config.v2.toml
port = 8080
-->

```ant
try File.symlink("config.v2.toml", "config.toml")
println(File.read("config.toml"))
```

```text
Ok{ value: "port = 8080\n" }
```

## Applications

### Process the lines of a file

Split the text with [String.lines](../String/lines.md), which drops the final line break:

<!-- test: file todo.txt
buy milk
call Ann
water plants
-->

```ant
let text = try File.read("todo.txt")
for { index, value } in text.lines().entries() {
  println("${index + 1}. ${value}")
}
```

```text
1. buy milk
2. call Ann
3. water plants
```

### Load a configuration file

Read, parse, and say which file failed with `Result.wrapErr`:

<!-- test: file settings.json
{"name": "demo", "debug": true}
-->

```ant
func loadSettings(path: string): Result[Json.Value, Error] {
  let text = try (File.read(path) | Result.wrapErr("load settings"))
  Json.parse(text) | Result.wrapErr("parse ${path}")
}

println(loadSettings("settings.json").isOk())
println(loadSettings("other.json"))
```

```text
true
Err{ error: Error{ message: "load settings: open other.json: no such file or directory" } }
```

### Sum a column of a CSV file

Parse the text with [Csv.parse](../Csv/parse.md) and skip the header row. A row whose amount is not
an integer is an `Err`:

<!-- test: file sales.csv
item,amount
tea,4
coffee,7
cake,5
-->

```ant
func sumAmounts(path: string): Result[int, Error] {
  let text = try File.read(path)
  let rows = try Csv.parse(text)
  var total = 0
  for row in rows.slice(1, rows.length()) {
    let amount = row[1].toInt()
    if amount == null { return Result.fail("${path}: amount ${row[1]} is not an integer") }
    total += amount
  }
  Result.ok(total)
}

println(sumAmounts("sales.csv"))
```

```text
Ok{ value: 16 }
```

### Edit a file in place

Read the text, change it, and write it back with [File.write](write.md):

<!-- test: file app.conf
host = localhost
port = 8080
-->

```ant
let text = try File.read("app.conf")
try File.write("app.conf", text.replace("8080", "9090"))
println(try File.read("app.conf"))
```

```text
host = localhost
port = 9090

```

## Properties and relations

What [File.write](write.md) writes, `File.read` reads back unchanged:

```ant
let text = "line one\nline two\n"
try File.write("round.txt", text)
assert(try File.read("round.txt") == text)
```

`File.read` gives what [File.readBytes](readBytes.md) gives, decoded by
[String.fromBytesLossy](../String/fromBytesLossy.md):

```ant
try File.writeBytes("mixed.txt", Bytes.fromArray([0x6F, 0x6B, 0xFF, 0x21]))
let bytes = try File.readBytes("mixed.txt")
assert(try File.read("mixed.txt") == String.fromBytesLossy(bytes))
```

## Possible issues

### The final line break

Most text files end with a line break, so splitting on `"\n"` gives an empty last element.
[String.lines](../String/lines.md) does not:

<!-- test: file names.txt
Ann
Bob
-->

```ant
let text = try File.read("names.txt")
println(text.split("\n"))
println(text.lines())
```

```text
["Ann", "Bob", ""]
["Ann", "Bob"]
```

### A byte order mark is kept

A file saved with a byte order mark starts with `"\u{FEFF}"`, which makes a comparison of its first
line fail. Remove it with `trimPrefix`:

```ant
try File.writeBytes("bom.csv", Bytes.fromArray([0xEF, 0xBB, 0xBF, 0x69, 0x64]))
let text = try File.read("bom.csv")
println(text == "id")
println(text.trimPrefix("\u{FEFF}") == "id")
```

```text
false
true
```

### Binary data is changed

`File.read` is for text. Bytes that are not UTF-8 are replaced, so writing the text back does not
restore the file. Copy binary data with [File.readBytes](readBytes.md) or [File.copy](copy.md):

```ant
let original = Bytes.fromArray([0x89, 0x50, 0x4E, 0x47])
try File.writeBytes("image.png", original)
try File.write("copy.png", try File.read("image.png"))
println(try File.readBytes("copy.png") == original)
```

```text
false
```

### A missing file is not told apart in the Err

The `Err` does not say why the read failed, so a program that treats a missing file differently from
other failures checks [File.exists](exists.md) first. Here a missing file means no entries, while a
directory or a file the program may not read is still an `Err`:

```ant
func loadEntries(path: string): Result[Array[string], Error] {
  let text = if File.exists(path) { try File.read(path) } else { "" }
  Result.ok(text.lines())
}

println(loadEntries("history.txt"))
try File.mkdir("cache")
println(loadEntries("cache"))
```

```text
Ok{ value: [] }
Err{ error: Error{ message: "read cache: is a directory" } }
```

### The result is a `Result`

The text is inside the `Result`, so using it directly as a string is a compile error:

```ant
let size = File.read("notes.txt").length()
// error: Result[string, Error] has no field or function length
```

## See also

[File.readBytes](readBytes.md) · [File.write](write.md) · [File.open](open.md) ·
[File.readLine](readLine.md) · [File.readAll](readAll.md) · [File.exists](exists.md) ·
[String.lines](../String/lines.md) · [String.fromBytesLossy](../String/fromBytesLossy.md)
