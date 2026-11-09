<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# eprint

```ant signature
func eprint[T](value: T): void
```

eprint writes String.from(value) to standard error, with no newline. Neither stream is buffered, so
the two interleave in program order.

<!-- end generated -->

## Details

- `eprint` writes to standard error, not standard output, so a pipe or a `>` redirect does not
  capture it. Use it for progress, warnings, and prompts that are not part of a program's result.
- It writes no newline. End the line with [eprintln](eprintln.md).
- It writes what [String.from](../String/from.md) gives for the value, so a string is written
  without quotes.
- A failed write is ignored, except a broken pipe, which ends the program with status 141.

## Basic examples

Start a line of status on standard error and finish it with [eprintln](eprintln.md) once the work is
done. Only the result reaches standard output, which the text block below shows:

```ant
eprint("loading config... ")
let port = 8080
eprintln("done")
println("port: ${port}")
```

```text
port: 8080
```

Prompt on standard error, so that standard output holds only the answer:

<!-- test: stdin
Ann
-->

```ant
eprint("name: ")
let name = try File.readLine(Os.stdin())
println("Hello, ${name ?? "stranger"}")
```

```text
Hello, Ann
```

## Scope

Any value is written as [String.from](../String/from.md) renders it: numbers, records, arrays,
`null`, and strings without quotes:

```ant
type Point = { x: int, y: int }
eprint(42)
eprint(" ")
eprint(Point{ x: 1, y: 2 })
eprint(" ")
eprint([1, 2, null])
eprint(" ")
eprint("é\n")
```

Standard error receives `42 Point{ x: 1, y: 2 } [1, 2, null] é` and a newline.

An empty string writes nothing:

```ant
eprint("")
```

`eprint` is a function value too. Give it a concrete type where a function is expected:

```ant
let warn: func(string): void = eprint
warn("disk almost full\n")
```

## Applications

Show progress on standard error while standard output carries the data. A pipe receives only the
names; a person at the terminal sees a dot per file as well:

```ant
let files = ["a.txt", "b.txt", "c.txt"]
for file in files {
  eprint(".")
  println(String.toUpper(file))
}
eprintln(" done")
```

```text
A.TXT
B.TXT
C.TXT
```

Report a usage error and exit with a nonzero status:

<!-- test: exit 2 -->

```ant
let args = Os.args()
if len(args) == 0 {
  eprint("usage: ")
  eprintln("${Path.base(Os.program())} <file>...")
  Os.exit(2)
}
println(args[0])
```

## Properties and relations

`eprint(v)` writes the same text as [File.writeText](../File/writeText.md) on
[Os.stderr](../Os/stderr.md), except that a failed write returns no `Result`:

```ant
let value = [1, 2]
// Each of the next two lines writes [1, 2] to standard error.
eprint(value)
try File.writeText(Os.stderr(), String.from(value))
```

## Possible issues

`eprint` needs an argument. Only [println](println.md) and [eprintln](eprintln.md) may be called
with none:

```ant
eprint() // error: eprint needs an argument for value
```

In a test, [ant test](../../tools/test.md) collects what `eprint` writes and shows it only when the
test fails, or with `-v`. A passing test prints nothing.

A result written with `eprint` is lost to a pipe or a `>` redirect. Write what a program produces
with [print](print.md) or [println](println.md), and keep `eprint` for messages about it.

## See also

[eprintln](eprintln.md), [print](print.md), [println](println.md), [Os.stderr](../Os/stderr.md),
[File.writeText](../File/writeText.md), [String.from](../String/from.md)
