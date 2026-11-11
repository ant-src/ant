<!-- status: reviewed -->

# String literals and interpolation

Double quotes write a `string`: `"hello"`. Inside the quotes, a backslash starts an escape, and
`${...}` interpolates the value of any expression.

## Details

- The escapes are `\n`, `\t`, `\r`, `\\`, `\"`, `\$`, and `\u{...}` with a hexadecimal Unicode
  scalar value. Any other backslash sequence is a compile error.
- `${...}` renders its value as `String.from` does, so any type can appear there. Braces nest inside
  it, so a lambda, a record literal, or another string fits.
- A `$` not followed by `{` stands for itself. Write `\$` for a literal `$` before `{`.
- A string may span lines. Each line break is one `\n` in the value, and the indentation of the
  following lines stays as written.
- Ant has no character literal, and single quotes do not write a string. A one-character string such
  as `"a"` takes its place. For text with backslashes or quotes, see [raw strings](raw-strings.md).

## Basic examples

```ant
"hello"       // "hello"
"line\nbreak" // "line\nbreak"
"\u{1F600}"   // "😀"
```

Interpolate a name and an expression:

```ant
let name = "Ant"
let count = 3
"hello, ${name}"            // "hello, Ant"
"total: ${count * 2} items" // "total: 6 items"
```

## Scope

### Escapes

Each escape writes one character. `println` shows the text the escapes stand for:

```ant
println("a\tb")
println("say \"hi\"")
println("C:\\temp")
println("cost: \${price}")
println("caf\u{e9}")
```

```text
a	b
say "hi"
C:\temp
cost: ${price}
café
```

### Values of every type

`${...}` renders a number, a `bool`, `null`, a record, or a collection. A string inside a collection
keeps its quotes:

```ant
type Point = { x: int, y: int }
let tags = ["new", "sale"]
"${1.5} ${true} ${null}"    // "1.5 true null"
"at ${Point{ x: 1, y: 2 }}" // "at Point{ x: 1, y: 2 }"
"tags: ${tags}"             // "tags: [\"new\", \"sale\"]"
"ratio: ${2.0}"             // "ratio: 2.0"
```

### Expressions inside `${...}`

Any expression fits: a call through `.`, a lambda, a pipe, an `if` expression, `??`, or another
string:

```ant
let words = ["ant", "bee"]
let n = words.length()
let user: string | null = null
"${words.map({ x.toUpper() })}"             // "[\"ANT\", \"BEE\"]"
"${words | String.join(" and ")}"           // "ant and bee"
"hi, ${user ?? "guest"}"                    // "hi, guest"
"${n} item${if n == 1 { "" } else { "s" }}" // "2 items"
```

### A dollar sign

A `$` without a `{` after it needs no escape:

```ant
let price = 5
"$${price}"  // "$5"
"100% of $5" // "100% of $5"
```

### Several lines

A line break inside the quotes is part of the string, and so is the indentation after it:

```ant
func letter(name: string): string {
  "Dear ${name},
  thank you."
}
println(letter("Ann"))
```

```text
Dear Ann,
  thank you.
```

Start the following lines at the left margin when the text must not be indented:

```ant
let usage = "usage: ant run <file>
flags:
  -h  show help"
println(usage)
println(usage.lines().length())
```

```text
usage: ant run <file>
flags:
  -h  show help
3
```

## Applications

Align the columns of a table:

```ant
type Item = { name: string, qty: int }
let items = [Item{ name: "apple", qty: 3 }, Item{ name: "fig", qty: 12 }]
for item in items {
  let qty = String.from(item.qty)
  println("${item.name.padEnd(8)}${qty.padStart(4)}")
}
```

```text
apple      3
fig       12
```

Build an error message from the values that caused it:

```ant
func checkAge(age: int): Result[int, Error] {
  if age < 0 { return Result.fail("age must not be negative, got ${age}") }
  Result.ok(age)
}
checkAge(-2) // Err{ error: Error{ message: "age must not be negative, got -2" } }
```

## Properties and relations

Interpolation is `String.from` of the value, joined with the text around it:

```ant
let n = 42
let xs = [1, 2]
assert("n=${n}" == "n=" + String.from(n))
assert("${xs}" == String.from(xs))
```

An escape and the character it stands for give the same string:

```ant
assert("\u{e9}" == "é")
assert("\u{41}" == "A")
assert("a
b" == "a\nb")
```

`\u{...}` writes one code point, so `e` followed by `\u{301}` is two code points and has length 2,
although it looks like `é`, which `\u{e9}` writes as one:

```ant
"\u{e9}".length()      // 1
"e\u{301}".length()    // 2
"e\u{301}" == "\u{e9}" // false
```

## Possible issues

### An unknown escape

Only the listed escapes exist. JavaScript's `\x41` and C's `\0` are compile errors:

```ant
let tab = "\x09" // error: bad escape '\x'
```

JavaScript's four-digit form needs braces in Ant:

```ant
let a = "\u0041" // error: bad escape: \u needs a {...} code
```

```ant
"\u{0041}" // "A"
```

### A surrogate is not a scalar value

```ant
let half = "\u{D800}" // error: bad escape: \u{D800} is not a Unicode scalar value
```

### Single quotes

```ant
let letter = 'a' // error: unexpected character '\''
```

### `+` does not convert

`+` joins a string only to another string; it does not turn a number into text. Interpolate the
number instead:

```ant
let count = 3
let label = "items: " + count // error: + needs both operands of one type, got string and int
```

```ant
let count = 3
"items: ${count}" // "items: 3"
```

### `null` renders as text

A nullable value interpolates as `null` rather than failing. Give a default with `??`:

```ant
let nickname: string | null = null
"hi, ${nickname}"             // "hi, null"
"hi, ${nickname ?? "friend"}" // "hi, friend"
```

### `${` meant for another program

A shell or template variable in a string is interpolated by Ant, and fails when Ant has no such
name. Escape the `$`, or use a [raw string](raw-strings.md). Both give the same string, which the
REPL shows with the `$` escaped:

```ant
let script = "echo ${HOME}" // error: HOME is not defined
```

```ant
"echo \${HOME}" // "echo \${HOME}"
`echo ${HOME}`  // "echo \${HOME}"
```

When a name with that spelling is in scope, Ant substitutes its value with no error, and the other
program never sees `${...}`:

```ant
let HOME = "/srv"
"echo ${HOME}" // "echo /srv"
```

## See also

[string](../types/string.md) · [Raw strings](raw-strings.md) ·
[String.from](../../lib/String/from.md) · [String functions](../../lib/String/index.md) ·
[Regular expression literals](regex.md) · [Strings tour](../../tour/04-strings.md) ·
[Strings guide](../../guide/strings.md)
