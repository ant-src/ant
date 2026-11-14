<!-- status: reviewed -->

# Read standard input line by line

This program prints each line of its standard input with its number. Save it as `number.ant` and
pipe text into it with `printf 'apple\nbanana\n' | ant run number.ant`:

<!-- test: stdin
apple
banana
-->
<!-- test: cmd ant run number.ant -->

```ant [number.ant]
let input = Os.stdin()
var n = 0
while true {
  let line = try input.readLine() ?? break
  n += 1
  println("${n}: ${line}")
}
```

```text
1: apple
2: banana
```

1. Get the handle of standard input with [Os.stdin](../../lib/Os/stdin.md).
2. Call [File.readLine](../../lib/File/readLine.md) in a loop. Each call returns the next line
   without its `\n` or `\r\n`, an empty line as `""`, and `null` at the end of the input.
3. End the loop at `null` with `?? break`. `try` ends the script on a read error, writes the error
   to standard error and exits with status 1.
4. Handle the line, here with [println](../../lib/prelude/println.md).

## Write many lines fast

Standard output is not buffered. For a large input, write through
[File.buffered](../../lib/File/buffered.md) and close the buffered handle at the end with
[File.close](../../lib/File/close.md), which writes out what it holds:

<!-- test: stdin
apple
banana
-->

```ant
let input = Os.stdin()
let out = Os.stdout().buffered()
defer { out.close() }
while true {
  let line = try input.readLine() ?? break
  try out.writeLine(line.toUpper())
}
try out.close()
```

```text
APPLE
BANANA
```

[File.writeLine](../../lib/File/writeLine.md) writes each line to the buffer, and
[String.toUpper](../../lib/String/toUpper.md) upper-cases each line. The deferred close writes out
what the buffer holds when a read error ends the script; the final `try out.close()` reports an
error from writing it out.

## Read a file or standard input

Read the file named by the first argument, or standard input when there is none, and print the
length of each line. Get the arguments with [Os.args](../../lib/Os/args.md), check for none with
[Array.isEmpty](../../lib/Array/isEmpty.md), open the file with [File.open](../../lib/File/open.md)
and measure each line with [String.length](../../lib/String/length.md). Closing standard input does
nothing, so one `defer` serves both:

<!-- test: stdin
apple
banana
-->

```ant
let args = Os.args()
let input = if args.isEmpty() { Os.stdin() } else { try File.open(args[0]) }
defer { input.close() }
while true {
  let line = try input.readLine() ?? break
  println(line.length())
}
```

```text
5
6
```

## Read all of standard input at once

To get the whole input as one string instead of a loop, call
[File.readAll](../../lib/File/readAll.md). It waits for the end of the input:

<!-- test: stdin
apple
banana
-->

```ant
let text = try Os.stdin().readAll()
println(text.length())
```

```text
13
```
