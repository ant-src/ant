<!-- status: reviewed -->

# Identifiers

An identifier names a variable, function, type, parameter, or field. It starts with an ASCII letter
or `_` and continues with ASCII letters, digits, or `_`.

## Details

- Identifiers are case-sensitive: `value` and `Value` are two names. Case means nothing else; it
  does not decide what a module exports, [export](../keywords/export.md) does.
- Letters outside ASCII, such as `é` or `名`, are not identifier characters.
- `_` on its own is the discard. It can be bound or assigned wherever a name can, any number of
  times, and the value is dropped. Reading it is a compile error in `ant check` and a panic in
  `ant run`. In the REPL, reading `_` gives the result the last input printed; writing it still
  discards.
- [Keywords and reserved words](reserved.md) cannot name a variable, function, type, or parameter,
  but any of them can name a field.
- The standard library names, such as `println`, `Array`, and `Result`, are ordinary identifiers
  bound by the prelude. No declaration in any scope may take one of them.

## Basic examples

```ant
let count = 3
let maxRetries = 5
let item2 = 4
count + maxRetries + item2 // 12
```

Case makes two different names:

```ant
let value = 1
let Value = 2
value + Value // 3
```

`_` discards a value that a destructuring or a loop would otherwise bind:

```ant
let [_, second, _] = [10, 20, 30]
second // 20
```

## Scope

Any mix of case, digits after the first character, and underscores is a name. The common styles all
work, and none of them means anything to the language:

```ant
let userName = "ann"
let MAX_SIZE = 64
let _draft = true
let point3d = [1, 2, 3]
type HttpRequest = { url: string }
let request = HttpRequest{ url: "/" }
request.url // "/"
```

A `for` loop that repeats an action and does not need the element binds `_`:

```ant
for _ in Iter.range(0, 3) {
  println("tick")
}
```

```text
tick
tick
tick
```

Parameters may all be `_`, so a function can ignore the arguments a caller passes:

```ant
func always(_: string, _: int): bool {
  true
}
always("x", 1) // true
```

A keyword or reserved word names a field after `.` or `?.`, and before `:` in a record type, a
record or map literal, and a record destructuring:

```ant
type Event = { type: string, class: string }
let event = Event{ type: "click", class: "button" }
let maybe: Event | null = event
event.type            // "click"
event.class           // "button"
maybe?.type           // "click"
Map{ null: 1, if: 2 } // Map{ "null": 1, "if": 2 }
```

Rename such a field to bind it as a variable:

```ant
type Event = { type: string }
let event = Event{ type: "click" }
let { type: kind } = event
kind // "click"
```

## Applications

A library exports exactly the names `export` marks, whatever their case. A lowercase `area` is
exported and a capitalized `Unit` is too:

```ant [shapes.ant]
export func area(width: int, height: int): int {
  width * height
}

export let Unit = "cm"
```

```ant
import "./shapes.ant" { area, Unit }

println("${area(3, 4)} ${Unit}")
```

```text
12 cm
```

## Possible issues

Names that differ only in case are different names, so a wrong case is an undefined name:

```ant
let total = 10
println(Total) // error: Total is not defined
```

A letter outside ASCII is not part of an identifier:

```ant
let café = "open" // error: unexpected character 'é'
```

An identifier cannot start with a digit or contain `-`. The first reads as a number, the second as
subtraction:

```ant
let 2nd = 2 // error: malformed number literal
```

```ant
let max-size = 64 // error: expected '=', got '-'
```

A keyword or reserved word cannot be a variable or parameter name:

```ant
let class = "button" // error: 'class' cannot be a variable name
```

```ant
func describe(type: string): string { type }
// error: 'type' is a keyword and cannot be a parameter name
```

The destructuring shorthand would bind a variable named after the field, so it fails for a keyword
field. Rename it instead:

```ant
type Event = { type: string }
let event = Event{ type: "click" }
let { type } = event // error: 'type' is a keyword and cannot be bound: write { type: name }
```

A prelude name cannot be declared again, even in an inner scope:

```ant
func label(): string {
  let println = "print" // error: println is a prelude name and cannot be redeclared
  println
}
```

`_` can only be written:

```ant
let _ = 5
let n = _ // error: _ is write-only and cannot be read
```

## See also

[Keywords and reserved words](reserved.md) · [Comments](comments.md) · [let](../keywords/let.md) ·
[var](../keywords/var.md) · [export](../keywords/export.md)
