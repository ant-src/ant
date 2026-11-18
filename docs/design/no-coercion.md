<!-- status: reviewed -->

# No implicit coercion, no truthiness

In Ant, a value never changes type unless the code says so. `+` does not turn a number into text, an
`int` does not quietly become a `float64`, and a condition must be a `bool`: there are no truthy or
falsy values. This page explains why, and what the rule costs.

## What other languages do quietly

| Code                           | JavaScript gives | Ant                                             |
| ------------------------------ | ---------------- | ----------------------------------------------- |
| `"items: " + 3`                | `"items: 3"`     | a compile error; write `"items: ${3}"`          |
| `"3" * 2`                      | `6`              | a compile error; parse with `"3".toInt()` first |
| `1 == "1"`                     | `true`           | a compile error                                 |
| `if (count)` with `count` 0    | skips the branch | a compile error; ask `count > 0`                |
| `port \|\| 8080` with `port` 0 | `8080`           | a compile error; `port ?? 8080` gives `0`       |

Each conversion on the left is a rule the reader has to remember and the writer has to get right.
Most of the time the rule does what was meant. When it does not, nothing fails: the program goes on
with a wrong value, and the bug shows up far from the line that caused it. Ant turns each of these
into a compile error at the line, and gives each conversion a name you write.

## Numbers never mix

```ant
let total = 10
let count = 4
total / count                   // 2
float64(total) / float64(count) // 2.5
```

Integer division truncates and float division does not, so which one runs matters. With both
operands of one kind, the kind is in plain sight: `total / count` divides integers, and the
conversions say where the code switches to floats. A language that widens `int` to `float` on its
own decides that for you, at a line that does not show it.

```ant
let total = 10
let average = total / 4.0 // error: / needs both operands of one type, got int and float64
```

Sizes do not mix either. An `int8` and an `int` are two kinds, so adding them is a compile error
until one side is converted. That matters where the size was chosen on purpose, such as a byte read
from a file or a field of a binary format: the code that leaves the small kind says so.

## A conversion says what it does

Converting is calling the target type as a function. A conversion to an integer kind never loses
information quietly:

```ant
int8(100)            // 100
int8(300)            // panic: 300 out of range for int8
int(2.5)             // panic: 2.5 has a fraction, so it does not convert to int
int(Math.trunc(2.5)) // 2
float64(7)           // 7.0
```

A conversion to an integer kind is exact or panics. Dropping the fraction is a choice between
truncating, rounding, and taking the floor, so the code makes it: `Math.trunc`, `Math.round`, or
`Math.floor` first, then `int`. A conversion to a float rounds to the nearest value and never
panics.

Text and numbers are further apart still, and the functions that cross between them have names that
say which way they go:

```ant
"42".toInt()    // 42
"4x".toInt()    // null
String.from(42) // "42"
```

`string` is not the way to render a number. It only unwraps a named type over `string`, so
`string(42)` is an error, and `String.from(42)` is the call that renders:

```ant
let s = string(42) // error: string converts only a string or a named primitive type over string, got int
```

Parsing can fail, so `toInt` returns `null` for text that is not a number, and the caller decides
what that means. A coercion has no way to say "this is not a number": JavaScript's `"4x" * 2` is
`NaN`, which then flows through every later step.

## A condition is a `bool`

```ant
func describe(count: int, name: string, user: string | null): string {
  if count { return "has items" } // error: cannot use int where bool is expected
  if name { return name }         // error: cannot use string where bool is expected
  if user { return user }
  // error: cannot use string | null where bool is expected
  // error: cannot use string | null where string is expected
  "nothing"
}
```

Truthiness answers several questions with one rule: is the number nonzero, is the string empty, is
the value there at all. The rule also differs between languages: an empty array is truthy in
JavaScript and falsy in Python. In Ant you ask the question you mean:

```ant
func describe(count: int, name: string, user: string | null): string {
  if count > 0 { return "${count} items" }
  if !name.isEmpty() { return name }
  if user != null { return user }
  "nothing"
}

println(describe(0, "", "ann"))
println(describe(2, "", null))
```

```text
ann
2 items
```

The explicit test pays for itself in the last guard: `user != null` is a test the type checker
understands, so `user` is a `string` inside the branch and the function can return it. `if`,
`while`, `!`, `&&`, and `||` all take `bool` and nothing else, so `&&` and `||` always give a
`bool`; `port || 8080` is not a way to pick a default.

## Absence is not falsy

```ant
let saved: int | null = 0
saved ?? 8080 // 0
```

In JavaScript, `saved || 8080` replaces a saved `0` with `8080`, because `0` is falsy, and an empty
string or `false` would be replaced the same way. Ant keeps absence apart from every other value.
`null` is the only absent value, and `??` replaces only `null`, so a zero, an empty string, and
`false` are kept as the real values they are.

## Equality does not convert

```ant
let count = 3
let same = count == "3"
// error: int and string cannot be compared with ==: neither is assignable to the other
```

An `int` and a `string` are never equal, so comparing them is a mistake the checker reports. To
compare across kinds, convert one side, and the code then says which comparison it means:
`String.from(count) == "3"` compares text, and `"3".toInt() == count` compares numbers. Since `==`
never converts, Ant needs no second, strict equality: there is no `===`.

## Your own types stay apart

The same rule covers named primitive types. A `Meters` and a `Seconds` are both `float64`
underneath, but neither turns into the other or into a plain `float64`:

```ant
type Meters = float64
type Seconds = float64

let distance = Meters(100.0)
let time = Seconds(9.58)
let speed = distance / time // error: / needs both operands of one type, got Meters and Seconds
```

Dividing meters by seconds gives neither, so the code unwraps both, as
`float64(distance) / float64(time)`. A type that converted to its base on its own would let a user
ID through where an account ID was expected, which is what the named type is there to prevent.

## The one implicit rendering

String interpolation renders any value as text: `"items: ${count}"`. It calls `String.from`, the
only function the language calls on its own: inside `${...}` of a string, and on a number in an HTML
expression such as `<p>{count}</p>`. The braces around the value mark the conversion. The print
functions take any value and render it the same way. Rendering for output is something a program
does all the time, and it does not change the value: `count` is still an `int` after the string is
built.

## Not even at run time

Types are advisory: `ant run` does not check a program first. A program that skips `ant check` still
never coerces. Where the checker would have reported `+` on an `int` and a `string`, the interpreter
panics at that line instead of producing `"items: 3"`, and the same holds for an `int` used as a
condition. `==` across kinds is `false`, never `true` by conversion. An unchecked program never
carries a converted value on.

## What it costs

- **More conversions in mixed arithmetic.** Code that mixes integer counts and float measurements
  converts at each step, as in `float64(total) / float64(count)`.
- **Literals have one type.** An integer literal is always an `int` and a float literal always a
  `float64`. `1` is never a float, so a `float64` needs `1.0`, and `int8(5) + 1` is a compile error:
  write `int8(5) + int8(1)`.
- **Some conversions panic.** `int(x)` on a fraction, `NaN`, or a value out of range stops the
  program. That is the point, since the alternative is a wrong number, but it means you pick the
  rounding before you convert.
- **Longer conditions.** `if count > 0` and `if !name.isEmpty()` are longer than `if count` and
  `if name`. They also say which question the code asks.
- **Unwrapping named types.** Library functions take the base kinds, so a `Meters` is unwrapped with
  `float64(m)` before it goes to `Math.sqrt`.

## See also

The [types guide](../guide/types.md) and the [null guide](../guide/null.md) show the rules in use.
Reference: [arithmetic operators](../language/operators/arithmetic.md),
[equality](../language/operators/equality.md),
[logical operators](../language/operators/logical.md),
[`??`](../language/operators/null-coalescing.md), [`bool`](../language/types/bool.md),
[`int`](../lib/prelude/int.md), [`float64`](../lib/prelude/float64.md),
[`String.toInt`](../lib/String/toInt.md), [`String.from`](../lib/String/from.md). The design notes
[One obvious way](one-way.md) and [Types are advisory](advisory-types.md) cover related decisions.
