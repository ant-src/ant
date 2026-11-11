<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# println

```ant signature
func println[T](value: T): void
```

println writes String.from(value) and a newline to standard output. println() with no argument
writes only the newline.

<!-- end generated -->

## Details

- A string is written without quotes; a string inside an array, a map, or a record keeps its quotes.
- Standard output is not buffered, so each call writes at once, in program order with
  [print](print.md) and [eprintln](eprintln.md). A program that writes many lines can buffer them
  with [File.buffered](../File/buffered.md).
- A failed write is ignored, except a broken pipe, which ends the program with status 141. Write
  with [File.writeLine](../File/writeLine.md) to get a `Result` instead.

## Basic examples

Write a line:

```ant
println("Hello, Ant")
```

```text
Hello, Ant
```

Interpolate values into the line:

```ant
let name = "Ann"
let age = 31
println("${name} is ${age}")
```

```text
Ann is 31
```

## Scope

Numbers, booleans, and `null`:

```ant
println(42)
println(3.0)
println(1.0 / 3.0)
println(true)
println(null)
```

```text
42
3.0
0.3333333333333333
true
null
```

A string is written as it is. Inside a collection, strings are quoted:

```ant
println("say \"hi\"")
println(["say", "hi"])
```

```text
say "hi"
["say", "hi"]
```

Records, union members, and named primitive types are written with their type name:

```ant
type Point = { x: int, y: int }
type User = { name: string, age: int }
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
type Count = int

let shape: Shape = Circle{ radius: 1.5 }
println(Point{ x: 1, y: 2 })
println(User{ name: "Ann", age: 31 })
println(shape)
println(Count(3))
```

```text
Point{ x: 1, y: 2 }
User{ name: "Ann", age: 31 }
Circle{ radius: 1.5 }
Count(3)
```

Arrays, maps, sets, `Result`s, and `Bytes`:

```ant
let parsed: Result[int, Error] = Ok{ value: 7 }
println([1, 2, null])
println(Map{ "a": 1, "b": 2 })
println(Set.of([1, 2]))
println(parsed)
println("hi".toBytes())
```

```text
[1, 2, null]
Map{ "a": 1, "b": 2 }
Set.of([1, 2])
Ok{ value: 7 }
Bytes(68 69)
```

An `Html` value is written as its markup. Interpolated text is escaped:

```ant
let name = "Ann & Bob"
println(<p>Hello, {name}</p>)
```

```text
<p>Hello, Ann &amp; Bob</p>
```

With no argument, `println` writes an empty line. An empty string does the same:

```ant
println("first")
println()
println("")
println("last")
```

```text
first


last
```

A newline inside the string starts a new line of output:

```ant
println("one\ntwo")
```

```text
one
two
```

`println` is a function value too. Give it a concrete type where a function is expected:

```ant
let show: func(int): void = println
for n in [1, 2, 3] {
  show(n * n)
}
```

```text
1
4
9
```

## Applications

Print a small report, one line per entry:

```ant
let stock = Map{ "apples": 12, "pears": 0, "plums": 7 }
for { key, value } in stock {
  let note = if value == 0 { " (sold out)" } else { "" }
  println("${key.padEnd(8, " ")}${value}${note}")
}
```

```text
apples  12
pears   0 (sold out)
plums   7
```

Print a table with a separator line:

```ant
let header = ["name", "age"]
let rows = [["Ann", "31"], ["Bob", "27"]]
println(String.join(header, " | "))
println("-".repeat(10))
for row in rows {
  println(String.join(row, " | "))
}
```

```text
name | age
----------
Ann | 31
Bob | 27
```

Write many lines through a buffer, and flush it at the end:

```ant
let out = Os.stdout().buffered()
for n in Array.range(1, 4) {
  try out.writeLine("line ${n}")
}
try out.flush()
```

```text
line 1
line 2
line 3
```

## Properties and relations

`println(v)` writes the same text as [print](print.md) of `v` followed by `"\n"`:

```ant
let value = [1, 2]
println(value)
print(value)
print("\n")
```

```text
[1, 2]
[1, 2]
```

The text `println` writes before the newline is the text [String.from](../String/from.md) returns:

```ant
type Point = { x: int, y: int }
assert(String.from(Point{ x: 1, y: 2 }) == "Point{ x: 1, y: 2 }")
assert(String.from("hi") == "hi")
assert(String.from(null) == "null")
```

`println(v)` writes what [File.writeLine](../File/writeLine.md) writes on
[Os.stdout](../Os/stdout.md), except that a failed write returns no `Result`:

```ant
let value = 42
println(value)
try File.writeLine(Os.stdout(), String.from(value))
```

```text
42
42
```

## Possible issues

`println` takes at most one argument. Join the parts with interpolation instead:

```ant
let a = 1
let b = 2
println(a, b) // error: println takes 0 to 1 arguments, got 2
```

```ant
let a = 1
let b = 2
println("${a} ${b}")
```

```text
1 2
```

As a function value, `println` is generic, so it needs a type where it is used:

```ant
let show = println
// error: println is generic; call it, or give the expected function type that fixes its type parameters
```

The no-argument form exists only in a call. A `func(): void` cannot hold `println`:

```ant
let blank: func(): void = println // error: cannot use println where func(): void is expected
```

```ant
let blank: func(): void = { println() }
blank()
```

`println` writes to standard output, so a pipe or a `>` redirect captures it. Write warnings and
progress with [eprintln](eprintln.md) so they stay out of the captured output.

## See also

[print](print.md), [eprintln](eprintln.md), [eprint](eprint.md), [String.from](../String/from.md),
[File.writeLine](../File/writeLine.md), [File.buffered](../File/buffered.md),
[Os.stdout](../Os/stdout.md), [string interpolation](../../language/literals/strings.md)
