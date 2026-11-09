<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.from

```ant signature
func String.from[T](value: T): string
```

String.from renders any value as text, in the format print, println, and string interpolation use. A
string renders as itself at the top level and quoted with escapes inside a container; a float
renders as the shortest decimal that reads back to it, with .0 when integral.

<!-- end generated -->

## Details

- `print`, `println`, and every `${...}` in a string render through `String.from`.
- A float renders as the shortest decimal that reads back to the same value, with `.0` when
  integral. At or above `1e21`, or below `1e-6`, it takes the exponent form, such as `1e+21`.
- A string renders as itself at the top level. Inside an array, a map, a set, or a record it is
  quoted and escaped, so it reads back as a literal.
- A value already being rendered renders as `...`, so a cyclic structure terminates.
- A user type cannot change how `String.from` renders it.

## Basic examples

Render a number as text:

```ant
String.from(42)  // "42"
String.from(1.0) // "1.0"
```

A string inside a container is quoted, so the elements stay apart:

```ant
String.from(["a", "b"]) // "[\"a\", \"b\"]"
```

`println` prints the same text:

```ant
println(["a", "b"])
```

```text
["a", "b"]
```

## Scope

### Numbers

Every integer type renders in decimal:

```ant
String.from(-7)          // "-7"
String.from(byte(255))   // "255"
String.from(Math.minInt) // "-9223372036854775808"
```

A float renders as the shortest decimal that reads back to it. An integral float keeps `.0`, so it
never looks like an `int`:

```ant
String.from(1.5)       // "1.5"
String.from(100.0)     // "100.0"
String.from(0.1 + 0.2) // "0.30000000000000004"
String.from(1.0 / 3.0) // "0.3333333333333333"
```

A `float32` renders as the shortest decimal for a `float32`:

```ant
String.from(float32(0.1)) // "0.1"
```

Very large and very small magnitudes take the exponent form, without `.0`:

```ant
String.from(1e20)     // "100000000000000000000.0"
String.from(1e21)     // "1e+21"
String.from(0.000001) // "0.000001"
String.from(1.5e-7)   // "1.5e-7"
```

`-0.0` keeps its sign, and the non-finite values have names:

```ant
String.from(-0.0)           // "-0.0"
String.from(Math.NaN)       // "NaN"
String.from(Math.Infinity)  // "Infinity"
String.from(-Math.Infinity) // "-Infinity"
```

### Strings

At the top level a string is returned as it is, the empty string included:

```ant
String.from("hi") // "hi"
String.from("")   // ""
```

Inside a container a string is quoted. The program prints the rendering, so you see it without the
quotes a case adds:

```ant
println(["", "say \"hi\"", `C:\temp`])
```

```text
["", "say \"hi\"", "C:\\temp"]
```

`\n`, `\t`, and `\r` are written as escapes, and so is a `$` before `{`, so the rendering reads back
as a string literal. A `$` before anything else stays as it is:

```ant
println(["one\ntwo\tthree", "a\rb", "\${name}", "$5"])
```

```text
["one\ntwo\tthree", "a\rb", "\${name}", "$5"]
```

A code point that `String.isPrint` rejects, such as a control character, a non-breaking space, or a
zero-width joiner, is written `\u{...}` in lower-case hexadecimal. An emoji joined from several code
points shows each of them:

```ant
println(["\u{7f}", "a\u{a0}b", "👨‍👩‍👧"])
```

```text
["\u{7f}", "a\u{a0}b", "👨\u{200d}👩\u{200d}👧"]
```

### bool, null, and void

```ant
String.from(true) // "true"
String.from(null) // "null"
```

`void` renders as its literal too:

```ant
var nothing = println("first")
println(String.from(nothing))
```

```text
first
void
```

### Arrays, maps, and sets

Entries are separated by `, `. Map keys that are strings are always quoted, and a set renders as the
call that builds it:

```ant
println([[1, 2], [3]])
println(Map{ apples: 3, "green pears": 5 })
println(Set.of(["x", "y"]))
```

```text
[[1, 2], [3]]
Map{ "apples": 3, "green pears": 5 }
Set.of(["x", "y"])
```

Empty containers render with nothing inside the brackets:

```ant
var none: Array[int] = []
var empty: Map[string, int] = Map{}
var noTags: Set[string] = Set.of([])
println([String.from(none), String.from(empty), String.from(noTags)])
```

```text
["[]", "Map{}", "Set.of([])"]
```

### Records and unions

A record renders with its tag and its fields in declaration order:

```ant
type User = { name: string, tags: Array[string] }
println(User{ name: "Ann", tags: ["admin"] })
```

```text
User{ name: "Ann", tags: ["admin"] }
```

A value of a union renders as the member it holds, and an empty record keeps its braces:

```ant
type Circle = { radius: float64 }
type Unknown = {}
type Shape = Circle | Unknown
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Unknown{}]
println(shapes)
```

```text
[Circle{ radius: 1.0 }, Unknown{}]
```

A `Result` is a union of records, so it renders as one:

```ant
var parsed: Result[int, string] = Result.ok(1)
String.from(parsed) // "Ok{ value: 1 }"
```

### Named primitive types

A named primitive renders as the call that builds it, so a `UserId` never passes for a plain string:

```ant
type UserId = string
String.from(UserId("u-1")) // "UserId(\"u-1\")"
```

The constructor itself renders as the type's name:

```ant
type UserId = string
String.from(UserId) // "UserId"
```

### Bytes and Html

`Bytes` renders as two lower-case hexadecimal digits per byte:

```ant
var empty: Array[int] = []
String.from(Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])) // "Bytes(89 50 4e 47)"
String.from(Bytes.fromArray(empty))                    // "Bytes()"
```

`Html` renders as its markup, with text escaped. An empty `Html` renders as `<></>` inside a
container:

```ant
var name = "<Ann>"
println(<p class="greeting">Hi, {name}</p>)
println([<b>bold</b>, <></>])
```

```text
<p class="greeting">Hi, &lt;Ann&gt;</p>
[<b>bold</b>, <></>]
```

### Functions, patterns, and time values

A function renders as `func`, and a regular expression as a literal that reads back to it:

```ant
String.from(func(n: int): int { n + 1 }) // "func"
String.from(String.length)               // "func"
String.from(/a\/b/i)                     // "/a\\/b/i"
```

A time value renders as its RFC 9557 string:

```ant
println([Time.fromEpochMillis(0), Time.plainDate(2026, 11, 8)])
```

```text
[1970-01-01T00:00:00Z, 2026-11-08]
```

### Sync objects, handles, and types

A sync object, a `Random`, or a handle renders as its type's name:

```ant
String.from(Mutex.new()) // "Mutex"
```

A type value renders in annotation syntax:

```ant
String.from(type(Array[int])) // "Array[int]"
```

### Cyclic values

A value already being rendered renders as `...`, so a structure that contains itself still prints:

```ant
type Node = { value: int, next: Array[Node] }
var node = Node{ value: 1, next: [] }
node.next = [node]
println(node)
```

```text
Node{ value: 1, next: [...] }
```

### Calls through `.` and pipes

On a string, `.from` is `String.from`, and a pipe passes the value as the first argument:

```ant
"ant".from()          // "ant"
3 | String.from()     // "3"
String.from(value: 3) // "3"
```

## Applications

Interpolation renders each `${...}` with `String.from`, so a log line needs no conversions:

```ant
type Order = { id: int, total: float64 }
var order = Order{ id: 7, total: 19.5 }
println("order ${order.id}: ${order.total} (${order})")
```

```text
order 7: 19.5 (Order{ id: 7, total: 19.5 })
```

Pass `String.from` to `Array.map` to turn numbers into text, then join them:

```ant
var widths = [3, 14, 159]
String.join(Array.map(widths, String.from), " | ") // "3 | 14 | 159"
```

Right-align a column of numbers by rendering each one first:

```ant
for n in [7, 42, 1024] {
  println(String.padStart(String.from(n), 6))
}
```

```text
     7
    42
  1024
```

Print a debug view of a whole structure, quotes and all, to see where a string has stray spaces:

```ant
var fields = String.split("Ann , 31", ",")
println(fields)
```

```text
["Ann ", " 31"]
```

## Properties and relations

Interpolation and `String.from` agree on every value:

```ant
var values = [1.5, 100.0, -0.0]
for v in values {
  assert("${v}" == String.from(v))
}
```

A string at the top level renders as itself:

```ant
for s in ["", "a b", "say \"hi\""] {
  assert(String.from(s) == s)
}
```

A finite float reads back from its rendering with `String.toFloat`:

```ant
for x in [0.1, 1.0 / 3.0, 1e21, 1.5e-7, -0.0] {
  assert(String.toFloat(String.from(x)) == x)
}
```

For an `int`, `String.from` writes the same digits as `String.fromInt` in base 10:

```ant
for n in [0, -42, Math.maxInt] {
  assert(String.from(n) == String.fromInt(n))
}
```

## Possible issues

A string is quoted inside a container, so rendering an array of words is not the way to join them.
Use `String.join`:

```ant
var words = ["red", "green"]
String.from(words)       // "[\"red\", \"green\"]"
String.join(words, ", ") // "red, green"
```

A named primitive renders as its constructor call. Convert it with `string(...)` to get the text:

```ant
type UserId = string
var id = UserId("u-1")
"user ${id}"         // "user UserId(\"u-1\")"
"user ${string(id)}" // "user u-1"
```

The rendering of a float shows every digit it needs to read back, so money and measurements print
with noise. Use `String.fromFloat` for a fixed number of decimals:

```ant
var total = 0.1 + 0.2
"${total}"                 // "0.30000000000000004"
String.fromFloat(total, 2) // "0.30"
```

A function named `from` on your own type does not change how `String.from`, `println`, or
interpolation render it. Call it yourself:

```ant
type Point = { x: float64, y: float64 }
func Point.from(p: Point): string { "(${p.x}, ${p.y})" }
var p = Point{ x: 1.0, y: 2.0 }
"${p}"   // "Point{ x: 1.0, y: 2.0 }"
p.from() // "(1.0, 2.0)"
```

`String.from` is not JSON: records keep their tag and maps their `Map` prefix. Use `Json.stringify`
for data another program reads:

```ant
type Point = { x: float64, y: float64 }
var p = Point{ x: 1.0, y: 2.0 }
String.from(p)    // "Point{ x: 1.0, y: 2.0 }"
Json.stringify(p) // "{\"x\":1.0,\"y\":2.0}"
```

`T` comes from the argument, so a lambda, which needs an expected function type, cannot be passed:

<!-- test: error a lambda needs an expected function type, but T is expected -->

```ant
var text = String.from({ x + 1 })
```

An empty literal fixes no element type either. Annotate it first:

<!-- test: error cannot infer the element type of an empty array -->

```ant
var text = String.from([])
```

`.from` reaches `String.from` only on a string. On a number it is a compile error; call
`String.from` directly:

<!-- test: error int has no field or function from -->

```ant
var n = 42
var text = n.from()
```

## See also

[String.fromInt](fromInt.md), [String.fromFloat](fromFloat.md), [String.join](join.md),
[String.toFloat](toFloat.md), [String.isPrint](isPrint.md), [println](../prelude/println.md),
[print](../prelude/print.md), [eprint](../prelude/eprint.md), [eprintln](../prelude/eprintln.md),
[Json.stringify](../Json/stringify.md)
