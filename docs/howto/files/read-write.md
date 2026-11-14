<!-- status: reviewed -->

# Read and write a whole file

Read a file into one string, change it, and write it back.

1. Read the file with [File.read](../../lib/File/read.md). It returns the whole file as a string.
2. Change the text, here with [String.replace](../../lib/String/replace.md).
3. Write it with [File.write](../../lib/File/write.md). It creates the file, or replaces what the
   file held.
4. Use [`try`](../../language/keywords/try.md) on both calls, so a script stops at a missing file or
   a failed write, writes the error to standard error, and exits with status 1.

<!-- test: file todo.txt
[ ] buy milk
[ ] call Ann
-->

```ant
let text = try File.read("todo.txt")
try File.write("todo.txt", text.replace("[ ] call Ann", "[x] call Ann"))
print(try File.read("todo.txt"))
```

```text
[ ] buy milk
[x] call Ann
```

For a file too large to hold in memory, read it a line at a time instead:
[Read a large file line by line](lines.md).

## Write a new file without replacing one

Pass `replace: false` to create the file only when nothing is at the path. When something is, the
write is an `Err` and the file keeps what it held:

<!-- test: file report.txt
first report
-->

```ant
println(File.write("report.txt", "second report", replace: false))
print(try File.read("report.txt"))
```

```text
Err{ error: Error{ message: "open report.txt: file exists" } }
first report
```

## Say what the program was doing

In a function that returns a `Result`, use `try` to return the `Err` to the caller, and
[Result.wrapErr](../../lib/Result/wrapErr.md) to say what the function was doing. The error already
names the file:

```ant
func countWords(path: string): Result[int, Error] {
  let text = try (File.read(path) | Result.wrapErr("count words"))
  Result.ok(text.fields().length())
}

try File.write("poem.txt", "so much depends\nupon a red wheel\nbarrow\n")
println(countWords("poem.txt"))
println(countWords("missing.txt"))
```

```text
Ok{ value: 8 }
Err{ error: Error{ message: "count words: open missing.txt: no such file or directory" } }
```

[String.fields](../../lib/String/fields.md) splits the text at runs of white space.

## Create the directory first

`File.write` does not create missing directories. Create them with
[File.mkdir](../../lib/File/mkdir.md), which makes the missing parents too and is `Ok` when the
directory already exists, and build the path with [Path.join](../../lib/Path/join.md):

```ant
let dir = Path.join(["out", "2026"])
let path = Path.join([dir, "summary.txt"])
println(File.write(path, "done\n"))
try File.mkdir(dir)
try File.write(path, "done\n")
print(try File.read(path))
```

```text
Err{ error: Error{ message: "open out/2026/summary.txt: no such file or directory" } }
done
```

## Add to the end of a file

`File.write` replaces what the file held. To add to it, open it with
[File.append](../../lib/File/append.md), which creates the file when it is missing, and close it
with `defer`:

```ant
try File.write("log.txt", "started\n")
let log = try File.append("log.txt")
defer { log.close() }
try log.writeLine("finished")
print(try File.read("log.txt"))
```

```text
started
finished
```

## Read and write bytes

For a file that is not text, such as an image, use [File.readBytes](../../lib/File/readBytes.md) and
[File.writeBytes](../../lib/File/writeBytes.md). They keep the bytes as they are. This program
writes the first bytes of a PNG file and checks for them with
[Bytes.startsWith](../../lib/Bytes/startsWith.md):

```ant
let signature = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
try File.writeBytes("logo.png", signature)
let data = try File.readBytes("logo.png")
println(data.length())
println(data.startsWith(signature))
```

```text
8
true
```

[Bytes.fromArray](../../lib/Bytes/fromArray.md) makes the bytes from numbers.

## Check that a file is text

`File.read` replaces bytes that are not valid UTF-8 with U+FFFD, `�`, so the read never fails on
such a file. To find out whether a file is text, read its bytes and convert them with
[String.fromBytes](../../lib/String/fromBytes.md), which returns `null` for bytes that are not
UTF-8:

```ant
try File.writeBytes("name.txt", Bytes.fromArray([0x41, 0x6e, 0xff]))
println(try File.read("name.txt"))
if String.fromBytes(try File.readBytes("name.txt")) == null {
  println("name.txt is not UTF-8")
}
```

```text
An�
name.txt is not UTF-8
```

## See also

[File.read](../../lib/File/read.md) · [File.write](../../lib/File/write.md) ·
[File.readBytes](../../lib/File/readBytes.md) · [File.writeBytes](../../lib/File/writeBytes.md) ·
[File.append](../../lib/File/append.md) · [File.mkdir](../../lib/File/mkdir.md) ·
[Path.join](../../lib/Path/join.md) · [Result.wrapErr](../../lib/Result/wrapErr.md) ·
[String.replace](../../lib/String/replace.md) · [String.fields](../../lib/String/fields.md) ·
[String.fromBytes](../../lib/String/fromBytes.md) · [Bytes.fromArray](../../lib/Bytes/fromArray.md)
· [Bytes.startsWith](../../lib/Bytes/startsWith.md) · [Read a large file line by line](lines.md) ·
[Replace a file without a partial write](replace-safely.md)
