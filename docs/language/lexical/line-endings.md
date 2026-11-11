<!-- status: reviewed -->

# Line endings

A newline ends a statement in Ant, as a semicolon does. A few clear signs continue a statement onto
the next line: an operator or an open bracket at the end of a line, or `.`, `?.`, `|`, or `else` at
the start of the next one.

## Details

- A line feed `\n`, a carriage return and line feed `\r\n`, and a carriage return `\r` alone are
  each one line break. A file saved with any of them means the same program.
- A line break inside a string or a raw string is one `\n` in the value, whatever the file uses.
- A statement continues when its line ends with a token that cannot end an expression, such as a
  binary operator, `,`, `(`, `[`, `{`, or `=`, or when the next non-blank line starts with `.`,
  `?.`, `|`, or `else`. Nothing else continues it.
- Inside parentheses, square brackets, and braces that are not a block, newlines are ignored.
- A `return` at the end of a line returns `void`, even when the next line starts with `.`, `?.`,
  `|`, or `else`.
- A trailing comment does not change where a statement ends.

## Basic examples

Each line is a statement:

```ant
let width = 3
let height = 4
println(width * height)
```

```text
12
```

An operator at the end of a line continues the expression on the next:

```ant
let total = 120 +
  45 +
  7
println(total)
```

```text
172
```

A pipe at the start of a line continues the expression above:

```ant
let names = ["ann", "bob", "cy"]
let line = names
  | Array.map({ x.toUpper() })
  | String.join(", ")
println(line)
```

```text
ANN, BOB, CY
```

## Scope

### Semicolons

A semicolon ends a statement too, so two statements can share a line:

```ant
let n = 5
if n > 3 { print("big"); println("!") }
```

```text
big!
```

### Dot calls on new lines

A line that starts with `.` or `?.` continues the line above, so a chain of dot calls can put one
call on each line:

```ant
let title = "  line endings  "
  .trim()
  .toUpper()
println(title)

type User = { name: string }
let user: User | null = null
let name = user
  ?.name
println(name)
```

```text
LINE ENDINGS
null
```

### `else` on the next line

A line that starts with `else` continues the `if` above it:

```ant
let n = 5
let size = if n < 3 { "small" }
else { "large" }
println(size)
```

```text
large
```

### Inside brackets

Inside parentheses, square brackets, and the braces of a literal, newlines are ignored. Items are
separated by commas, and a trailing comma is allowed:

```ant
type Point = { x: int, y: int }

let primes = [
  2,
  3,
  5,
]
let origin = Point{
  x: 0,
  y: 0,
}
println(primes)
println(origin)
println(Math.max(
  primes.length(),
  origin.x,
))
```

```text
[2, 3, 5]
Point{ x: 0, y: 0 }
3
```

The braces of a block are different: there, a newline separates the statements.

### Union types

A `|` at the start of a line continues a union type too:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Triangle = { base: float64, height: float64 }
type Shape = Circle
  | Square
  | Triangle

let shape: Shape = Square{ side: 2.0 }
println(shape)
```

```text
Square{ side: 2.0 }
```

### Strings across lines

A string or raw string literal may span lines. Each line break in it is one `\n` in the value:

```ant
"roses
are red" // "roses\nare red"
"a
b" == "a\nb" // true
`a
b` == "a\nb" // true
```

### Trailing comments

A comment at the end of a line does not change how the line ends. The `+` still continues the
expression:

```ant
let price = 100 + // base
  20
println(price)
```

```text
120
```

## Possible issues

### A leading operator starts a new statement

Only `.`, `?.`, `|`, and `else` continue a line from the start of the next one. A `-` there is a
unary minus in a statement of its own, so `total` below is `10`:

```ant
let total = 10
-3
println(total)
```

```text
10
```

Other binary operators cannot start a statement, so the line is a syntax error:

```ant
let ready = true
  && false // error: unexpected '&&'
```

Put the operator at the end of the line instead:

```ant
let ready = true &&
  false
println(ready)
```

```text
false
```

### A block's `{` on the next line

The `{` of a block must sit on the same line as its `if`, `while`, or `for` header. A newline ends
the header expression, even inside parentheses:

```ant
let n = 5
if n > 3 // error: expected '{' after the if header, got newline
{ println("big") }
```

Put the `{` on the same line as the header: `if n > 3 {`.

### A line that starts with `(` or `[`

A line that starts with `(` or `[` begins a new statement. It never continues a call or an index on
the line above:

```ant
let xs = [10, 20]
let first = xs
[0]
println(first)
```

```text
[10, 20]
```

Keep the `[0]` or the `(` of a call on the same line as what it applies to: `xs[0]`.

### `return` at the end of a line

A `return` at the end of a line returns `void`. The next line is a statement of its own:

```ant
func double(n: int): int {
  return // error: cannot use void where int is expected
  n * 2
}
```

A line that starts with `.` does not continue a bare `return` either, and since no statement starts
with `.`, it is a syntax error:

```ant
func shout(s: string): string {
  return
  .toUpper() // error: unexpected '.'
}
```

Start the value on the same line as `return`.

### A regular expression on one line

A regular expression literal cannot hold a line break:

```ant
let pattern = /one
// error: unterminated regular expression
two/
```

Write `\n` in the pattern to match a line break.

## See also

- [Comments](comments.md)
- [Pipe](../operators/pipe.md)
- [Dot calls](../operators/dot-calls.md)
- [Optional chaining](../operators/optional-chaining.md)
- [if](../keywords/if.md)
- [return](../keywords/return.md)
- [String literals](../literals/strings.md)
- [Regular expressions](../literals/regex.md)
