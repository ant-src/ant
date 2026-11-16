<!-- status: draft -->
<!-- brief: Tool page: looking up names and modules. The spec's tool document stays the reference; this page explains use. -->

# ant doc

`ant doc` prints the signature and doc comment of a declaration, from the standard library or from a
module. It shows in a terminal what an editor shows when you hover over a name.

```sh
ant doc <name>              # a prelude declaration or namespace: Array.map, Array, Result, print
ant doc <file.ant>          # a module's doc and its exported declarations
ant doc <file.ant> <name>   # one exported declaration of a module: Point, Point.distance
```

## Details

- A first argument that ends in `.ant` is a module path. Any other first argument is a prelude name.
- `ant doc` finds a name without regard to case, so `array.MAP` finds `Array.map`.
- In a module, `ant doc` sees only exported declarations. It reads only the file you give it: it
  does not follow imports and does not run the type checker, so a module that parses but fails
  `ant check` is documented like any other. A file that does not parse exits with status 2.
- A signature stops before the body or initializer: a `func` after its result type (or its
  parameters when it has none), a `type` after its whole definition, an `interface` after its type
  parameters, and a `var` or `let` after its type annotation.
- The exit status is 0 when `ant doc` finds the declaration or the module, 1 when there is no such
  declaration or a module has no doc and exports nothing, and 2 when the usage is wrong or the file
  cannot be read or parsed.

## Basic examples

Look up a function of the standard library by its full name:

```sh
ant doc Array.map
```

```text
func Array.map[T, U](xs: Array[T], f: func(T): U): Array[U]

    Array.map returns a new array holding f applied to each element of xs, in order.

        Array.map([1, 2, 3], { x * 2 })   // [2, 4, 6]
```

The signature comes first, then the doc comment, indented by four spaces.

On a terminal, signatures and examples use the REPL's syntax colors and namespace headings are bold.
All output keeps its source line breaks and spacing. `NO_COLOR`, redirected output, or `TERM=dumb`
disables highlighting and bold headings. The REPL's `:doc` uses the same rendering.

Give a file to see what a module offers. `ant doc` prints the module's `//!` doc and then the
signature of each exported declaration, in source order:

<!-- test: cmd ant doc geometry.ant -->

```ant [geometry.ant]
//! Geometry computes distances on the plane.
//!
//! Make a [Point] and measure with [Point.distance].

// Point is a position on the plane.
export type Point = { x: float64, y: float64 }

// Point.distance returns the straight-line distance from a to b.
export func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt(square(a.x - b.x) + square(a.y - b.y))
}

func square(n: float64): float64 {
  n * n
}
```

```text
    Geometry computes distances on the plane.

    Make a [Point] and measure with [Point.distance].

type Point = { x: float64, y: float64 }
func Point.distance(a: Point, b: Point): float64
```

`square` is not exported, so the listing leaves it out.

Add a name after the file to print one declaration with its doc comment:

<!-- test: cmd ant doc temperature.ant fahrenheit -->

```ant [temperature.ant]
// fahrenheit converts c degrees Celsius to degrees Fahrenheit.
export func fahrenheit(c: float64): float64 {
  c * 9.0 / 5.0 + 32.0
}
```

```text
func fahrenheit(c: float64): float64

    fahrenheit converts c degrees Celsius to degrees Fahrenheit.
```

## Scope

### Namespaces

A name without a dot that names a namespace prints the namespace's own doc and then every
declaration in it. In the standard library, that lists a whole area:

```sh
ant doc Result
ant doc Time
```

In a module, a type with functions declared on it is a namespace too. `ant doc` prints the type with
its doc comment, then the signature of each function on it:

<!-- test: cmd ant doc money.ant Money -->

```ant [money.ant]
// Money is an amount in cents of one currency.
export type Money = { cents: int, currency: string }

// Money.add returns the sum of a and b, which must share a currency.
export func Money.add(a: Money, b: Money): Money {
  assert(a.currency == b.currency)
  Money{ cents: a.cents + b.cents, currency: a.currency }
}

// Money.format returns m as text, such as "12.50 EUR".
export func Money.format(m: Money): string {
  let fraction = String.from(m.cents % 100).padStart(2, "0")
  "${m.cents / 100}.${fraction} ${m.currency}"
}
```

```text
type Money = { cents: int, currency: string }

    Money is an amount in cents of one currency.

func Money.add(a: Money, b: Money): Money
func Money.format(m: Money): string
```

### Case

Case does not matter when you look a name up, so you need not remember how a name is capitalized:

<!-- test: cmd ant doc units.ant DISTANCE.TOMILES -->

```ant [units.ant]
// Distance is a length in meters.
export type Distance = { meters: float64 }

// Distance.toMiles returns d in miles.
export func Distance.toMiles(d: Distance): float64 {
  d.meters / 1609.344
}
```

```text
func Distance.toMiles(d: Distance): float64

    Distance.toMiles returns d in miles.
```

### Interfaces

An interface prints its own doc and the signature of each of its functions. Name one function as
`Interface.function` to see its doc:

<!-- test: cmd ant doc log.ant Sink.writeLine -->

```ant [log.ant]
// Sink is anything a log line can be written to.
export interface Sink {
  // writeLine writes line and a line break to s.
  writeLine(s: Self, line: string): Result[void, Error]
}
```

```text
writeLine(s: Self, line: string): Result[void, Error]

    writeLine writes line and a line break to s.
```

### Constants and variables

A `let` or `var` prints with its type annotation, if it has one, and never its initializer:

<!-- test: cmd ant doc server.ant -->

```ant [server.ant]
// port is the TCP port the server listens on.
export let port: int = 8080

// retries is how many times a failed request is sent again.
export var retries = 3
```

```text
let port: int
var retries
```

### Scripts

A script exports nothing, so `ant doc` prints its `//!` doc, which says what the program does and
how to run it:

<!-- test: cmd ant doc greet.ant -->

```ant [greet.ant]
#!/usr/bin/env ant
//! Greet prints a greeting for each name it is given.
//!
//! Run it with: ant greet.ant Ann Bob

for name in Os.args() {
  println("Hello, ${name}")
}
```

```text
    Greet prints a greeting for each name it is given.

    Run it with: ant greet.ant Ann Bob
```

### Declaration files

A file whose name ends in `.d.ant` declares names bound outside the language. Every declaration in
it is visible, so `ant doc` lists them all:

<!-- test: cmd ant doc host.d.ant -->

```ant [host.d.ant]
// Host
//
// Host describes the machine the program runs on.

// Host.name returns the name of the machine.
func Host.name(): string

// Host.cores returns the number of processor cores.
func Host.cores(): int
```

```text
func Host.name(): string
func Host.cores(): int
```

`Host` is a namespace that no type declares, so its doc is the section label at the top of the file:
a comment line with the name alone, an empty comment line, and a paragraph that starts with the
name, set off from the declarations by a blank line. Looking up the namespace prints its name, then
that doc, then its declarations:

<!-- test: cmd ant doc host.d.ant Host -->

```ant [host.d.ant]
// Host
//
// Host describes the machine the program runs on.

// Host.name returns the name of the machine.
func Host.name(): string

// Host.cores returns the number of processor cores.
func Host.cores(): int
```

```text
Host

    Host describes the machine the program runs on.

func Host.name(): string
func Host.cores(): int
```

## Writing doc comments

A doc comment is plain text, read in a terminal as often as in an editor. The standard library
follows these conventions, and editors may rely on them:

- The first sentence is a summary that starts with the declared name, as written in a call. An
  editor shows it alone in a completion list, so it stands on its own. It ends at the first period
  followed by a space or a line break.
- An empty `//` line separates paragraphs.
- Lines indented past the comment's own space are code, which an editor shows as Ant in a monospaced
  block.
- A name in brackets, such as `[Math.min]`, names a declaration, and an editor may turn it into a
  link. Brackets right after a name, as in `Array[int]`, are type arguments, not a link.
- A namespace that no type declares gets its doc from a section label in a declaration file: a
  comment line with the name alone, an empty `//` line, and a paragraph that starts with the name,
  as in `host.d.ant` above.
- Parameters appear by their declared names, without markup, as in "panics if lo > hi".

There are no headings, emphasis, or tables. This comment follows every convention:

<!-- test: cmd ant doc range.ant clamp -->

```ant [range.ant]
// clamp returns n limited to the range from lo to hi. It panics if lo > hi.
//
// A value inside the range comes back unchanged:
//
//     clamp(5, 0, 10)    // 5
//     clamp(-3, 0, 10)   // 0
//
// To bound n on one side only, use [Math.min] or [Math.max].
export func clamp(n: int, lo: int, hi: int): int {
  Math.clamp(n, lo, hi)
}
```

```text
func clamp(n: int, lo: int, hi: int): int

    clamp returns n limited to the range from lo to hi. It panics if lo > hi.

    A value inside the range comes back unchanged:

        clamp(5, 0, 10)    // 5
        clamp(-3, 0, 10)   // 0

    To bound n on one side only, use [Math.min] or [Math.max].
```

A module doc starts with a summary of the module in its own terms, since each importer gives the
module a name of its own: `//! Geometry computes distances on the plane.` It goes on to what a
reader of the whole module needs: its main types, how its functions fit together, an example of use.

## Possible issues

A declaration that is not exported is not found. `ant doc tax.ant rate` prints
`ant doc: no declaration rate` and exits with status 1:

<!-- test: cmd ant doc tax.ant rate -->
<!-- test: exit 1 -->

```ant [tax.ant]
let rate = 0.25

// withTax returns price with tax added.
export func withTax(price: float64): float64 {
  price * (1.0 + rate)
}
```

`ant doc` does not follow imports either. `checkout.ant` imports `withTax`, but `withTax` is not a
declaration of `checkout.ant`, so looking it up there fails. Ask the module that declares it,
`ant doc tax.ant withTax`:

<!-- test: cmd ant doc checkout.ant withTax -->
<!-- test: exit 1 -->

```ant [checkout.ant]
import "./tax.ant" { withTax }

// total returns the price of all items, with tax.
export func total(prices: Array[float64]): float64 {
  withTax(prices.sum() ?? 0.0)
}
```

A `let` or `var` that destructures has no name of its own. The module lists it, but neither `width`
nor `height` finds it:

<!-- test: cmd ant doc screen.ant width -->
<!-- test: exit 1 -->

```ant [screen.ant]
// width and height are the size of the terminal in characters.
export let [width, height] = [80, 24]
```

Give each name a declaration of its own when it needs a doc of its own.

A module with no `//!` doc that exports nothing has nothing to show. `ant doc` prints
`ant doc: helpers.ant has no doc and exports nothing` and exits with status 1:

<!-- test: cmd ant doc helpers.ant -->
<!-- test: exit 1 -->

```ant [helpers.ant]
func double(n: int): int {
  n * 2
}
```

Without `.ant`, the argument is a prelude name, not a file: `ant doc geometry` looks for a prelude
declaration named `geometry` and exits with status 1. Write the whole file name,
`ant doc geometry.ant`.

A file that does not parse cannot be documented. Unlike a type error, which `ant doc` ignores
because it does not run the checker, a syntax error stops it. `ant doc broken.ant` reports the error
at its position, `broken.ant:1:16: expected a parameter name, got '{'`, and exits with status 2:

```ant [broken.ant]
export func f( { // error: expected a parameter name, got '{'
```

A blank line between a comment and a declaration ends the doc comment, so `ant doc` prints only the
signature. See [Comments and doc comments](../language/lexical/comments.md).

## See also

[Comments and doc comments](../language/lexical/comments.md), [Documenting code](../guide/docs.md),
[`export`](../language/keywords/export.md), [`ant check`](check.md)
