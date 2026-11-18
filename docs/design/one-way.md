<!-- status: reviewed -->

# One obvious way

When two constructs would do the same job, Ant keeps one. A job has one keyword, one operator, or
one literal, and a file has one layout. This page explains what the rule removes, where Ant still
has two forms and why, and what the rule costs.

## One construct per job

```ant
type Paid = { amount: int }
type Refunded = { amount: int, reason: string }
type Cancelled = {}
type Event = Paid | Refunded | Cancelled

func describe(e: Event): string {
  if e is Paid { "paid ${e.amount}" }
  else if e is Refunded { "refunded ${e.amount}: ${e.reason}" }
  else { "cancelled" }
}

let events: Array[Event] = [
  Paid{ amount: 30 },
  Refunded{ amount: 10, reason: "damaged" },
  Cancelled{},
]
for e in events {
  println(describe(e))
}
```

```text
paid 30
refunded 10: damaged
cancelled
```

`describe` uses the one construct Ant has for each of its jobs. `is` with `if` is the only way to
narrow a union of records: there is no `switch`, no `match`, and no `kind` field to compare. The
`if` is an expression, so its branches are the result, and there is no `?:` to choose between values
a second way. A value goes into text with `${...}`; there is no format string.

The same holds across the language:

| Job                           | Ant                         | Not in Ant                    |
| ----------------------------- | --------------------------- | ----------------------------- |
| Name a value that stays fixed | `let`                       | `const`                       |
| Name a value you reassign     | `var`                       |                               |
| Add one                       | `n += 1`                    | `n++`                         |
| Test equality                 | `==`                        | `===`                         |
| Choose between values         | an `if` expression          | `?:`, `switch`                |
| Narrow a union                | `if x is T`                 | `switch`, `match`             |
| Mark absence                  | `null`                      | `undefined`, an `Option` type |
| Report an expected failure    | `Result` and `try`          | `throw`, a `catch` block      |
| Keep data and its functions   | records, functions on types | `class`, `this`               |
| Run code concurrently         | `go`                        | `async`, `await`              |
| Combine bits                  | `Bits.or(a, b)`             | a bitwise operator            |

You learn one form for each job and meet the same form in every program, so reading code is about
what it does, not about which form its author picked.

## One form per literal

```ant
0o755        // 493
0xff == 0xFF // true
1_000_000    // 1000000
1.0 / 4.0    // 0.25
2h30m        // PT2H30M
```

An integer literal is always an `int`, and a float literal is always a `float64`: there are no
numeric suffixes, and `1` is never a float. Each base has one prefix, in lowercase. Octal is `0o`
and nothing else, because other languages read a leading `0` as octal:

```ant
let mode = 007 // error: a decimal literal other than 0 cannot start with 0
```

Prefixes and exponents are lowercase, so `0XFF` and `1E9` are errors too. Only hexadecimal digits
may be either case.

A duration is a literal when it is a constant span of hours and smaller units, such as `500ms` or
`2h30m`. Everything else, a span computed from a number or measured on the calendar, is
`Time.duration(...)`. There is no literal for days, since a day is not always 24 hours:

```ant
let week = 7d
// error: there are no duration literals for days, weeks, months or years: write Time.duration(days: 1) and the like
```

Text has two literals with two jobs. A string between double quotes has escapes and interpolation; a
raw string between backticks has neither, for paths and patterns full of backslashes. There are no
single quotes and no character literals: a character is a string of one code point.

## Habits fail loudly

```ant
const limit = 10 // error: 'const' is a reserved word
```

The words other languages use for constructs Ant leaves out are reserved: `as`, `async`, `await`,
`catch`, `class`, `const`, `enum`, `finally`, `function`, `match`, `switch`, `this`, `throw`, and
`undefined`. A habit from JavaScript stops at the word, with an error, instead of becoming a name
that means something else. A field can use any keyword or reserved word, so `event.type` and
`Map{ class: "a" }` work.

Where a habit has an Ant form, the error names it:

```ant
let pattern = /ant/g // error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

## One layout

`ant fmt` has no layout options, no configuration file, and no way to turn it off for part of a
file. Every formatted file in every project has the same spacing and indentation, so a diff shows
only what changed, and there is no house style to agree on. Take this file:

```text
let prices = [12, 30,
  8]
let total = prices.sum()??0
println( "total: ${ total }" )
```

`ant fmt` rewrites it as:

```ant
let prices = [
  12, 30,
  8,
]
let total = prices.sum() ?? 0
println("total: ${total}")
```

```text
total: 50
```

The formatter keeps what a reader might have chosen on purpose: where a statement breaks, the rows
of a list, every pair of parentheses, and whether `else` follows the `}` or starts the next line.
`describe` above keeps its one-line branches for that reason. It never changes what a program means
and never rewrites one construct into another.

## Where two forms remain

Two forms stay when each has its own job. `let` and `var` both name a value, but the keyword tells
the reader whether the name changes. The two string literals differ in escapes. A duration literal
is a constant; `Time.duration` takes numbers. A regular expression in the source is a literal,
`/ant/i`; a pattern from data goes through `Regex.compile`, which returns a `Result`. A lambda
`{ x * 2 }` takes at most one argument, named `x`; `func(a: int, b: int) { a + b }` names and
annotates its parameters. `x ?? d` supplies a default for `null` only; any other choice is an `if`.

One call has three spellings, on purpose:

```ant
let words = ["ant", "bee"]
words.map({ x.toUpper() })         // ["ANT", "BEE"]
Array.map(words, { x.toUpper() })  // ["ANT", "BEE"]
words | Array.map({ x.toUpper() }) // ["ANT", "BEE"]
```

These are not three constructs. `words.map(f)` and `words | Array.map(f)` both stand for
`Array.map(words, f)`, and the type checker checks them as that call, with the same inference and
the same errors. `.` reaches the functions of a value's own namespace and the pipe reaches any other
function, so a chain of transformations reads top to bottom either way. Which spelling fits a line
is a matter of style, which [Idiomatic Ant](../guide/idiomatic.md) settles: dots for local
operations, pipes for a chain, the namespace where it explains the operation.

## What it costs

- **Habits break.** `const`, `n++`, `a === b`, `?:`, and `switch` are all errors. A JavaScript
  programmer relearns each, once.
- **Some code is longer.** You combine bits with functions, `Bits.or(READ, WRITE)`, so `|` has one
  meaning in an expression: the pipe. A negative duration is `Time.negate(5s)`, not `-5s`.
- **No shortcut for formatting text.** Without a format string, fixed digits and padding are
  function calls: `String.fromFloat(total, 2)`, `String.fromInt(n, 16).padStart(8, "0")`.
- **Style is not all settled.** `ant fmt` leaves line breaks to you and keeps the call spelling you
  wrote. Two people can still write the same call two ways; the language accepts both, and only a
  convention picks one.

## See also

[`ant fmt`](../tools/fmt.md) describes the formatter, and [Idiomatic Ant](../guide/idiomatic.md) the
conventions it leaves to you. Reference:
[Keywords and reserved words](../language/lexical/reserved.md),
[Number literals](../language/literals/numbers.md),
[Duration literals](../language/literals/durations.md), [Strings](../language/literals/strings.md),
[Raw strings](../language/literals/raw-strings.md),
[Calls through `.`](../language/operators/dot-calls.md), [Pipes](../language/operators/pipe.md),
[`Bits`](../lib/Bits/index.md). The design notes
[No implicit coercion, no truthiness](no-coercion.md),
[Why everything is an expression](expressions.md), [Why no exceptions](errors-are-values.md),
[Records instead of classes](records-not-classes.md), and
[Why goroutines and not async](no-async.md) cover related decisions.
