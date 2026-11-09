<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# eprintln

```ant signature
func eprintln[T](value: T): void
```

eprintln writes String.from(value) and a newline to standard error, for a warning or a usage line.
eprintln() with no argument writes only the newline.

<!-- end generated -->

## Details

- Standard error is not buffered, so `eprintln` and [println](println.md) come out in the order the
  program calls them.
- `eprintln` ignores a failed write, except a write to a stream whose reader has gone (a broken
  pipe), which ends the program with status 141. To learn of a failed write, call
  [File.writeLine](../File/writeLine.md) on [Os.stderr](../Os/stderr.md), which returns a `Result`.

## Basic examples

Write a warning to standard error. Standard output holds only what [println](println.md) writes:

```ant
println("total: 42")
eprintln("warning: 3 rows skipped")
```

```text
total: 42
```

Print a usage line and exit with a nonzero status when an argument is missing:

<!-- test: exit 2 -->

```ant
let args = Os.args()
if len(args) != 1 {
  eprintln("usage: greet <name>")
  Os.exit(2)
}
println("Hello, ${args[0]}")
```

## Scope

Any value works. A string is written as it is; other values are written as
[String.from](../String/from.md) renders them:

```ant
type Point = { x: int, y: int }

eprintln("plain text")
eprintln(42)
eprintln(3.0)
eprintln(null)
eprintln([1, 2, 3])
eprintln(Point{ x: 1, y: 2 })
eprintln(Map{ "a": 1 })
```

Standard error then holds:

```text
plain text
42
3.0
null
[1, 2, 3]
Point{ x: 1, y: 2 }
Map{ "a": 1 }
```

With no argument, `eprintln` writes an empty line:

```ant
eprintln("first")
eprintln()
eprintln("after a blank line")
```

Interpolate to build a message from several values:

```ant
let path = "data.csv"
let line = 17
eprintln("${path}:${line}: missing column")
```

At the end of a pipe, the piped value becomes the argument:

```ant
let skipped = 3
"skipped ${skipped} files" | eprintln()
```

## Applications

Report an error from a `Result` and keep going:

```ant
let paths = ["missing.txt"]
for path in paths {
  let r = File.read(path)
  if r is Err {
    eprintln("error: ${r.error.message}")
    continue
  }
  println(r.value)
}
println("done")
```

```text
done
```

Keep the data on standard output clean, so `ant run report.ant > report.csv` writes only the rows
and the warning still reaches the terminal:

```ant
let rows = [["Ann", "31"], ["Bob", ""]]
for row in rows {
  if row[1] == "" {
    eprintln("warning: no age for ${row[0]}")
    continue
  }
  println(String.join(row, ","))
}
```

```text
Ann,31
```

End a line of progress dots that [eprint](eprint.md) wrote:

```ant
for _ in Iter.range(0, 5) {
  eprint(".")
}
eprintln()
eprintln("5 steps done")
```

## Properties and relations

`eprintln(x)` writes the same text as `eprint("${x}\n")`, since interpolation renders with
[String.from](../String/from.md) too:

```ant
let x = [1, 2]
assert("${x}" == String.from(x))
assert(String.from("text") == "text")
assert(String.from(["a", "b"]) == "[\"a\", \"b\"]")
```

## Possible issues

`eprintln` takes one value. Interpolate to write several:

```ant
let name = "Ann"
let age = 31
eprintln(name, age) // error: eprintln takes 0 to 1 arguments, got 2
```

`eprintln` is generic, so as a function value it needs a function type that fixes its type
parameter:

```ant
let warn = eprintln
// error: eprintln is generic; call it, or give the expected function type that fixes its type parameters
```

```ant
let warn: func(string): void = eprintln
warn("disk almost full")
```

The call with no argument exists only as a call. The function value takes one argument:

```ant
let blank: func(): void = eprintln // error: cannot use eprintln where func(): void is expected
```

`eprintln` is not a function on a type, so it cannot be called through `.`:

```ant
"oops".eprintln() // error: string has no field or function eprintln
```

## See also

[eprint](eprint.md), [println](println.md), [print](print.md), [Os.stderr](../Os/stderr.md),
[File.writeLine](../File/writeLine.md), [Os.exit](../Os/exit.md), [String.from](../String/from.md)
