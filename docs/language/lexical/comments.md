<!-- status: reviewed -->

# Comments and doc comments

A comment is text the type checker and the interpreter skip. `//` starts a line comment, which runs
to the end of its line, and `/* */` encloses a block comment. A run of line comments right above a
top-level declaration is its doc comment, and `//!` lines document the whole module; `ant doc` and
editors show them.

## Details

- Block comments do not nest: a block comment ends at the first `*/`.
- `//` and `/*` inside a string literal are part of the string, not a comment.
- A doc comment is a run of `//` lines, each alone on its line, directly above a top-level `var`,
  `let`, `func`, or `type`, or the `export` before one. A blank line, a block comment, a `//!` line,
  or a comment after code ends the run.
- `//!` lines come before the first `import` or declaration, on consecutive lines. Other comments,
  such as a license, may come before or after them, but not between.
- A file's first line may start with `#!`, so a script runs as a command. `#!` anywhere else is a
  syntax error.

## Basic examples

A line comment runs to the end of its line. It can stand alone or follow code:

```ant
// Count the guests.
let guests = ["Ann", "Bob"]
let count = guests.length() // two names
println(count)
```

```text
2
```

A block comment can sit inside an expression or span several lines:

```ant
let total = 10 + /* tip */ 2
/*
  The bill is split
  between two people.
*/
println(total / 2)
```

```text
6
```

A doc comment is the run of line comments right above a top-level declaration. `ant doc` prints it
under the declaration's signature:

<!-- test: cmd ant doc greeting.ant greet -->

```ant [greeting.ant]
// greet returns a greeting for name.
export func greet(name: string): string {
  "Hello, ${name}"
}
```

```text
func greet(name: string): string

    greet returns a greeting for name.
```

## Scope

### Comments between values

A block comment can sit between the parts of an expression, such as to label one argument of a call:

```ant
Math.max(/* lower */ 1, /* upper */ 5) // 5
10 + /* tip */ 2                       // 12
```

### Comments at line ends

A comment at the end of a line does not change where the expression ends. A line that ends with an
operator continues onto the next line, comment or not:

```ant
let total = 1 + // first
  2
println(total)
```

```text
3
```

### Comments in markup

Inside markup, a comment is a block comment in braces, `{/* ... */}`. It renders nothing:

```ant
let name = "Ann"
println(<p>{/* greeting */}Hello, {name}</p>)
```

```text
<p>Hello, Ann</p>
```

### Strings

`//` and `/*` in a string literal are characters of the string:

```ant
"https://ant-lang.org"    // "https://ant-lang.org"
"/* kept */".length()     // 10
`C:\Program Files // raw` // "C:\\Program Files // raw"
```

### Commenting out code

A block comment turns off several lines at once, as long as they hold no `*/`:

```ant
let steps: Array[string] = []
steps.push("build")
/*
steps.push("test")
steps.push("deploy")
*/
println(steps)
```

```text
["build"]
```

### Module docs and section labels

`//!` lines document the module. Here a license comment comes first and a blank line separates it,
so it documents nothing. `// Distances` is followed by a blank line, so it labels a section of the
file instead of documenting `distance`. In a doc comment, an empty `//` line separates paragraphs,
lines indented past the comment's own space are code, and a name in brackets, such as `[Point]`,
names a declaration that an editor may turn into a link:

<!-- test: cmd ant doc geometry.ant -->

```ant [geometry.ant]
// Copyright 2026 The Ant Authors. MIT license.

//! Geometry computes distances on the plane.
//!
//! Build a [Point] and pass two of them to [distance].

// Point is a position on the plane.
export type Point = { x: float64, y: float64 }

// Distances

// distance returns the straight-line distance from a to b.
//
//     distance(Point{ x: 0.0, y: 0.0 }, Point{ x: 3.0, y: 4.0 })   // 5.0
export func distance(a: Point, b: Point): float64 {
  Math.sqrt(square(a.x - b.x) + square(a.y - b.y))
}

func square(n: float64): float64 {
  n * n
}
```

For a module, `ant doc` prints the module doc and the signature of each exported declaration:

```text
    Geometry computes distances on the plane.

    Build a [Point] and pass two of them to [distance].

type Point = { x: float64, y: float64 }
func distance(a: Point, b: Point): float64
```

### Scripts

A script starts with `#!` on its first line so that it runs as a command. Its `//!` lines say what
the program does and how to run it:

```ant
#!/usr/bin/env ant
//! Greet prints a greeting.
//!
//! Run it with: ant greet.ant

println("Hello")
```

```text
Hello
```

## Possible issues

Block comments do not nest. The first `*/` closes the comment, and the rest of the line is read as
code:

```ant
/* outer /* inner */ still */
// error: unterminated regular expression
```

So a block comment cannot turn off code that already holds one. Use line comments for that:

```ant
let price = 10
// let discount = /* percent */ 5
println(price)
```

```text
10
```

A block comment needs its `*/`. Without one, it runs to the end of the file:

```ant
/* draft notes // error: unterminated block comment
println("done")
```

A blank line between a comment and a declaration ends the run, so the comment documents nothing and
`ant doc` prints only the signature:

<!-- test: cmd ant doc area.ant area -->

```ant [area.ant]
// area returns the area of a w by h rectangle.

export func area(w: float64, h: float64): float64 {
  w * h
}
```

```text
func area(w: float64, h: float64): float64
```

A `//!` line after the first declaration, or after code on its line, is a syntax error. Put the
module doc at the top of the file:

```ant
let limit = 3 //! Limits caps retries.
// error: a //! module doc line must come before the first import or declaration
println(limit)
```

An HTML comment, `<!-- -->`, is a syntax error in markup. Write `{/* ... */}` instead:

```ant
let page = <p><!-- note -->Hi</p> // error: an HTML comment is written {/* ... */} in markup
```

`//` always starts a comment, so an empty regular expression cannot be a literal, and neither can a
pattern that starts with `*`. Write the empty pattern with `Regex.compile("")`:

```ant
Regex.compile("") // Ok{ value: /(?:)/ }
```

`#!` works only from the first byte of the file. On any later line it is an error:

```ant
println("start")
#!/usr/bin/env ant // error: unexpected character '#'
```

## See also

[`ant doc`](../../tools/doc.md), [Writing docs](../../guide/docs.md), [Identifiers](identifiers.md),
[Line endings](line-endings.md), [Keywords and reserved words](reserved.md),
[`export`](../keywords/export.md), [`ant run`](../../tools/run.md)
