<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# print

```ant signature
func print[T](value: T): void
```

print writes String.from(value) to standard output, with no newline. Output is not buffered, so it
comes out in program order. A failed write is ignored, except a broken pipe, which ends the program
with status 141 as SIGPIPE does; File.writeLine(Os.stdout(), …) reports other errors.

<!-- end generated -->

## Details

- `print` writes no newline. End the line with [println](println.md), or put `\n` in the string.
- It writes what [String.from](../String/from.md) gives for the value, so a string is written
  without quotes, and a string inside an array or a record is quoted.
- Standard output is not buffered, so text written with `print` appears at once, before a read that
  follows it. A program that writes much output writes it through a
  [File.buffered](../File/buffered.md) handle on [Os.stdout](../Os/stdout.md) instead.
- A broken pipe ends the program with status 141, so `ant run gen.ant | head` stops once `head` has
  read enough.

## Basic examples

Write two pieces of text on one line, then end it:

```ant
print("Hello")
print(", world")
println()
```

```text
Hello, world
```

Print a prompt and read the answer on the same line:

<!-- test: stdin
Ann
-->

```ant
print("name: ")
let name = try File.readLine(Os.stdin())
println("Hello, ${name ?? "stranger"}")
```

```text
name: Hello, Ann
```

## Scope

Any value is written as [String.from](../String/from.md) renders it. A string at the top level has
no quotes; inside an array it has them:

```ant
type Point = { x: int, y: int }
type UserId = string
print(42)
print(" ")
print(1.0)
print(" ")
print(null)
print(" ")
print(Point{ x: 1, y: 2 })
print(" ")
print(UserId("u-1"))
print(" ")
print(["a", "b"])
print(" ")
print("é")
```

```text
42 1.0 null Point{ x: 1, y: 2 } UserId("u-1") ["a", "b"] é
```

An empty string writes nothing, and a `\n` inside the string ends a line:

```ant
print("")
print("one\ntwo\n")
```

```text
one
two
```

`print` takes the value through a pipe or by name, like any function:

```ant
"piped " | print()
print(value: "named\n")
```

```text
piped named
```

`print` is generic, so as a function value it needs a concrete type where a function is expected:

```ant
let show: func(int): void = print
for n in [1, 2, 3] {
  show(n)
}
println()
```

```text
123
```

## Applications

Join items with a separator as they come, writing the separator before every item but the first:

```ant
let words = ["one", "two", "three"]
for entry in Array.entries(words) {
  if entry.index > 0 {
    print(", ")
  }
  print(entry.value)
}
println()
```

```text
one, two, three
```

Print a table one cell at a time, ending each row with [println](println.md):

```ant
let rows = [[1, 20, 300], [4000, 5, 60]]
for row in rows {
  for n in row {
    print(String.padStart(String.from(n), 6))
  }
  println()
}
```

```text
     1    20   300
  4000     5    60
```

Show a dot per finished step on one line:

```ant
for _ in ["fetch", "build", "test"] {
  print(".")
}
println(" done")
```

```text
... done
```

## Properties and relations

`print(v)` followed by `println()` writes the same line as `println(v)`:

```ant
let value = Map{ "a": 1 }
print(value)
println()
println(value)
```

```text
Map{ "a": 1 }
Map{ "a": 1 }
```

`print(v)` writes the same text as [File.writeText](../File/writeText.md) on
[Os.stdout](../Os/stdout.md), except that a failed write returns no `Result`:

```ant
let value = [1, 2]
print(value)
try File.writeText(Os.stdout(), String.from(value))
```

```text
[1, 2][1, 2]
```

## Possible issues

`print` needs an argument; only [println](println.md) may be called with none:

```ant
print() // error: print needs an argument for value
```

`print` takes one value. Interpolate several into one string instead:

```ant
let name = "Ann"
let age = 31
print(name, age) // error: print takes 1 argument, got 2
```

`+` does not join a string and a number. Use interpolation:

```ant
let count = 3
print("count: " + count) // error: + needs both operands of one type, got string and int
print("count: ${count}\n")
```

Without a type to fix its type parameter, `print` is not a value:

```ant
let show = print
// error: print is generic; call it, or give the expected function type that fixes its type parameters
```

A `print` made while a [File.buffered](../File/buffered.md) handle on [Os.stdout](../Os/stdout.md)
holds output goes past the buffer and can come out first:

```ant
let out = File.buffered(Os.stdout())
try File.writeText(out, "A")
print("B")
try File.flush(out)
println()
```

```text
BA
```

Write everything through the buffered handle, or flush it before calling `print`.

## See also

[println](println.md), [eprint](eprint.md), [eprintln](eprintln.md),
[String.from](../String/from.md), [Os.stdout](../Os/stdout.md),
[File.writeText](../File/writeText.md), [File.buffered](../File/buffered.md)
