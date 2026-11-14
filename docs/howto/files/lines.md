<!-- status: reviewed -->

# Read a large file line by line

This program counts the lines of a log file that start with `ERROR`. It holds one line in memory at
a time, so the size of the file does not matter. Save it as `count.ant` and run it with
`ant run count.ant app.log`:

<!-- test: file app.log
INFO start
ERROR disk full
INFO retry
ERROR timeout
INFO done
-->
<!-- test: cmd ant run count.ant app.log -->

```ant [count.ant]
let args = Os.args()
if args.isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} <file>")
  Os.exit(2)
}
let input = try File.open(args[0])
defer { input.close() }
var errors = 0
while true {
  let line = try input.readLine() ?? break
  if line.startsWith("ERROR") {
    errors += 1
  }
}
println("${errors} errors")
```

```text
2 errors
```

1. Get the file name with [Os.args](../../lib/Os/args.md). When there is none, write a usage line
   with [eprintln](../../lib/prelude/eprintln.md) and exit with [Os.exit](../../lib/Os/exit.md).
   [Os.program](../../lib/Os/program.md) gives the script path, and
   [Path.base](../../lib/Path/base.md) keeps its last element.
2. Open the file with [File.open](../../lib/File/open.md), and close it with
   [File.close](../../lib/File/close.md) in a `defer` right after.
3. Call [File.readLine](../../lib/File/readLine.md) in a loop. Each call returns the next line
   without its `\n` or `\r\n`, an empty line as `""`, and `null` at the end of the file.
4. End the loop at `null` with `?? break`. `try` ends the script on a read error, writes the error
   to standard error and exits with status 1.
5. Handle the line, here with [String.startsWith](../../lib/String/startsWith.md).

## Stop at the first match

Return from inside the loop when the line you want turns up. The rest of the file is never read, and
the deferred close runs on every return. [String.contains](../../lib/String/contains.md) tests each
line. The function returns the line number and the line in `Result.ok`, or `Result.ok(null)` when no
line matches, and the caller turns null into a default with `??`:

<!-- test: file app.log
INFO start
ERROR disk full
INFO retry
ERROR timeout
INFO done
-->

```ant
func firstMatch(path: string, word: string): Result[string | null, Error] {
  let input = try File.open(path)
  defer { input.close() }
  var n = 0
  while true {
    let line = try input.readLine() ?? break
    n += 1
    if line.contains(word) {
      return Result.ok("${n}: ${line}")
    }
  }
  Result.ok(null)
}

println(try firstMatch("app.log", "timeout") ?? "no match")
println(try firstMatch("app.log", "panic") ?? "no match")
```

```text
4: ERROR timeout
no match
```

## Read a header line first

Read the first line before the loop, then loop over the rest. Every read continues where the last
one stopped. Here the loop sums the second column of a CSV file:
[String.cut](../../lib/String/cut.md) splits each line at the first comma,
[String.toInt](../../lib/String/toInt.md) parses the number, and `?? continue` skips a line without
one:

<!-- test: file sizes.csv
name,bytes
index.html,5120
logo.png,20480
-->

```ant
let input = try File.open("sizes.csv")
defer { input.close() }
let header = try input.readLine() ?? ""
println("columns: ${header}")
var total = 0
while true {
  let line = try input.readLine() ?? break
  let cut = line.cut(",") ?? continue
  total += cut.after.toInt() ?? continue
}
println("total: ${total}")
```

```text
columns: name,bytes
total: 25600
```

To parse quoted fields as well, read the whole file with [Csv.parse](../../lib/Csv/parse.md)
instead.

## Write the matching lines to another file

Writes to a file are not buffered: each `writeLine` goes to the OS. For many small writes, wrap the
output in [File.buffered](../../lib/File/buffered.md). Create the file with
[File.create](../../lib/File/create.md) and end with an explicit close, which writes out what the
buffer holds and reports an error the deferred close cannot:

<!-- test: file app.log
INFO start
ERROR disk full
INFO retry
ERROR timeout
INFO done
-->

```ant
let input = try File.open("app.log")
defer { input.close() }
let out = File.buffered(try File.create("errors.log"))
defer { out.close() }
while true {
  let line = try input.readLine() ?? break
  if line.startsWith("ERROR") {
    try out.writeLine(line)
  }
}
try out.close()
print(try File.read("errors.log"))
```

```text
ERROR disk full
ERROR timeout
```

## Read lines from standard input

To read the lines piped into the script, as in `grep -v DEBUG app.log | ant run count.ant`, loop
over [Os.stdin](../../lib/Os/stdin.md) the same way. The program does not open standard input, so it
does not close it either:

<!-- test: stdin
INFO start
ERROR disk full
INFO retry
ERROR timeout
INFO done
-->

```ant
var errors = 0
while true {
  let line = try Os.stdin().readLine() ?? break
  if line.startsWith("ERROR") {
    errors += 1
  }
}
println("${errors} errors")
```

```text
2 errors
```

## Read a small file at once

When the file fits in memory, read it whole with [File.read](../../lib/File/read.md) and split it
with [String.lines](../../lib/String/lines.md), and count the matches with
[Array.countWhere](../../lib/Array/countWhere.md):

<!-- test: file app.log
INFO start
ERROR disk full
INFO retry
ERROR timeout
INFO done
-->

```ant
let errors = try File.read("app.log") | String.lines() | Array.countWhere({ x.startsWith("ERROR") })
println("${errors} errors")
```

```text
2 errors
```

`File.read` holds the whole file and `String.lines` all of its lines, so a file of several gigabytes
needs that much memory. The `readLine` loop does not.
