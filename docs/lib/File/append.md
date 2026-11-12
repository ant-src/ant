<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.append

```ant signature
func File.append(path: string): Result[File.Handle, Error]
```

File.append opens the file at path for writing at its end, creating it when it does not exist.

<!-- end generated -->

## Details

- Writes go after what the file already holds; `File.append` never truncates the file.
- A new file gets mode `0o666` less the process's umask, `0o644` under the usual umask `022`. A file
  that exists keeps its mode.
- A symbolic link at `path` is followed, and the writes go to its target.
- Each write is handed to the OS before it returns. For many small writes, wrap the handle in
  [File.buffered](buffered.md) and close it.

## Basic examples

Open a log file, add a line, and close it. The deferred close covers early exits, and the final
close reports whether the line was written:

```ant
func log(path: string, message: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeLine(message)
  file.close()
}

println(log("app.log", "started"))
println(log("app.log", "stopped"))
println(File.read("app.log"))
```

```text
Ok{ value: void }
Ok{ value: void }
Ok{ value: "started\nstopped\n" }
```

The first call created `app.log`; the second added its line after the first.

## Scope

### An existing file

What the file held stays, and the new text follows it:

```ant
func addLine(path: string, line: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeLine(line)
  file.close()
}

File.write("todo.txt", "buy milk\n")
println(addLine("todo.txt", "call Ann"))
println(File.read("todo.txt"))
```

```text
Ok{ value: void }
Ok{ value: "buy milk\ncall Ann\n" }
```

### Text and bytes

The handle takes every kind of write: `writeLine` ends the line, `writeText` does not, and
`writeChunk` writes bytes as they are:

```ant
func record(path: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeText("total: ")
  try file.writeLine("42")
  try file.writeChunk("done\n".toBytes())
  file.close()
}

println(record("totals.txt"))
println(File.read("totals.txt"))
```

```text
Ok{ value: void }
Ok{ value: "total: 42\ndone\n" }
```

### A symbolic link

Appending through a link writes to the file the link points at:

```ant
func appendVia(link: string, line: string): Result[void, Error] {
  let file = try File.append(link)
  defer { file.close() }
  try file.writeLine(line)
  file.close()
}

File.write("history.txt", "one\n")
File.symlink("history.txt", "latest.txt")
println(appendVia("latest.txt", "two"))
println(File.read("history.txt"))
```

```text
Ok{ value: void }
Ok{ value: "one\ntwo\n" }
```

### Modes

A file that exists keeps its mode, so a private file stays private:

```ant
func appendSecret(path: string, line: string): Result[int, Error] {
  try File.write(path, "first\n")
  try File.chmod(path, 0o600)
  let file = try File.append(path)
  defer { file.close() }
  try file.writeLine(line)
  try file.close()
  let info = try File.stat(path)
  Result.ok(info.mode)
}

let mode = Result.unwrap(appendSecret("token.txt", "second"))
println(String.fromInt(mode, base: 8))
```

```text
600
```

## Applications

### Many lines at once

Wrap the handle in [File.buffered](buffered.md) when a loop writes many small lines, and close the
buffered handle, which writes out the buffer and closes the file:

```ant
func appendAll(path: string, lines: Array[string]): Result[void, Error] {
  let out = File.buffered(try File.append(path))
  defer { out.close() }
  for line in lines {
    try out.writeLine(line)
  }
  out.close()
}

println(appendAll("events.log", ["login ann", "login bob", "logout ann"]))
println(File.read("events.log"))
```

```text
Ok{ value: void }
Ok{ value: "login ann\nlogin bob\nlogout ann\n" }
```

### A CSV file with a header

Write the header only when the file is new, then append a row on every call:

```ant
func addRow(path: string, name: string, score: int): Result[void, Error] {
  let isNew = !File.exists(path)
  let file = try File.append(path)
  defer { file.close() }
  if isNew {
    try file.writeLine("name,score")
  }
  try file.writeLine("${name},${score}")
  file.close()
}

println(addRow("scores.csv", "Ann", 31))
println(addRow("scores.csv", "Bob", 27))
println(File.read("scores.csv"))
```

```text
Ok{ value: void }
Ok{ value: void }
Ok{ value: "name,score\nAnn,31\nBob,27\n" }
```

## Properties and relations

After an append, the file holds what it held before followed by the new text:

```ant
func appendText(path: string, text: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeText(text)
  file.close()
}

File.write("notes.txt", "a\n")
let before = Result.unwrap(File.read("notes.txt"))
Result.unwrap(appendText("notes.txt", "b\n"))
assert(Result.unwrap(File.read("notes.txt")) == before + "b\n")
```

On a file that does not exist, appending gives what [File.write](write.md) gives:

```ant
func appendText(path: string, text: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeText(text)
  file.close()
}

Result.unwrap(appendText("appended.txt", "hello\n"))
Result.unwrap(File.write("written.txt", "hello\n"))
assert(File.read("appended.txt") == File.read("written.txt"))
```

Opening and closing without a write creates an empty file and leaves an existing one unchanged:

```ant
func openAndClose(path: string): Result[void, Error] {
  let file = try File.append(path)
  file.close()
}

Result.unwrap(openAndClose("empty.txt"))
assert(Result.unwrap(File.read("empty.txt")) == "")
File.write("kept.txt", "kept\n")
Result.unwrap(openAndClose("kept.txt"))
assert(Result.unwrap(File.read("kept.txt")) == "kept\n")
```

## Possible issues

### Running twice adds the lines twice

`File.append` never starts over, so a program that writes a report with it doubles the report on its
second run:

```ant
func writeReport(path: string): Result[void, Error] {
  let file = try File.append(path)
  defer { file.close() }
  try file.writeLine("report")
  file.close()
}

writeReport("report.txt")
writeReport("report.txt")
println(File.read("report.txt"))
```

```text
Ok{ value: "report\nreport\n" }
```

Use [File.create](create.md) or [File.write](write.md) for a file that should hold only this run's
output.

### No `replace` parameter

`File.append` never replaces, so it takes no `replace:` argument:

```ant
let file = File.append("out.txt", replace: false) // error: File.append has no parameter replace
```

### A buffered handle that is never closed

What a buffered handle holds is written only at `flush` or `close`. Without either, the lines never
reach the file:

```ant
func appendLater(path: string): Result[void, Error] {
  let out = File.buffered(try File.append(path))
  out.writeLine("lost")
}

println(appendLater("later.txt"))
println(File.read("later.txt"))
```

```text
Ok{ value: void }
Ok{ value: "" }
```

Close the buffered handle, as in [Many lines at once](#many-lines-at-once).

## See also

[File.create](create.md), [File.write](write.md), [File.writeLine](writeLine.md),
[File.writeText](writeText.md), [File.buffered](buffered.md), [File.close](close.md),
[File.open](open.md), [File.Handle](types/Handle.md)
