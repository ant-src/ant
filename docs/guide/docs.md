<!-- status: reviewed -->

# Documenting code

Ant documentation lives in comments. A run of `//` lines right above a top-level declaration is its
doc comment, and `//!` lines at the top of a file are the module doc. `ant doc` prints them in a
terminal, and an editor shows them on hover. `ant check` and `ant run` ignore them.

## The parts

This library documents itself at three levels: the module, a type, and each function:

<!-- test: cmd ant doc stack.ant -->

```ant [stack.ant]
//! Stack is a last-in, first-out collection of values.
//!
//! Create one with [Stack.empty], add values with [Stack.push], and take them back with
//! [Stack.pop]:
//!
//!     let s: Stack[int] = Stack.empty()
//!     s.push(1)
//!     s.pop()   // 1

// Stack holds values in the order they were pushed.
export type Stack[T] = { items: Array[T] }

// Stack.empty returns a stack with no values.
export func Stack.empty[T](): Stack[T] {
  Stack{ items: [] }
}

// Stack.push adds value to the top of s.
export func Stack.push[T](s: Stack[T], value: T): void {
  s.items.push(value)
}

// Stack.pop removes the top value of s and returns it, or returns null when s is empty.
//
//     let s: Stack[string] = Stack.empty()
//     s.pop()   // null
export func Stack.pop[T](s: Stack[T]): T | null {
  s.items.pop()
}
```

`ant doc stack.ant` prints the module doc and then the signature of each exported declaration:

```text
    Stack is a last-in, first-out collection of values.

    Create one with [Stack.empty], add values with [Stack.push], and take them back with
    [Stack.pop]:

        let s: Stack[int] = Stack.empty()
        s.push(1)
        s.pop()   // 1

type Stack[T] = { items: Array[T] }
func Stack.empty[T](): Stack[T]
func Stack.push[T](s: Stack[T], value: T): void
func Stack.pop[T](s: Stack[T]): T | null
```

Pick the comment by what it describes:

- **A module doc**, `//!`, describes the whole file: what it is for, its main types, how its
  functions fit together, an example of use. It is the first thing a reader of the module sees.
- **A doc comment**, `//` right above a declaration, describes that one declaration. Write one for
  every export.
- **A plain comment** anywhere else explains the code to the next person who edits it. Nothing shows
  it outside the file.

## Writing a doc comment

A doc comment is plain text, read in a terminal as often as in an editor. The standard library
follows a few conventions, and editors may rely on them.

### Start with the name

The first sentence is a summary that starts with the declared name, written the way a call writes
it: `Stack.pop removes ...`, not `Removes ...` or `pop removes ...`. An editor shows that sentence
alone in a completion list, so it must make sense without the rest. It ends at the first period
followed by a space or a line break.

Name parameters as they are declared, without markup: "returns null when s is empty", "panics if n <
0".

### Paragraphs, code, and links

An empty comment line, `//` or `//!` with nothing after it, separates paragraphs. Lines indented by
at least one space beyond the comment's own space are code, which an editor shows in a monospaced
block. A name in brackets, such as `[Stack.push]`, names a declaration that an editor may turn into
a link. Bracketed text that names no declaration stays as written, and so does a bracket right after
a name, as in `Array[int]`.

There is no other markup: no headings, emphasis, tables, or tags.

### Examples with values

The code in a doc comment usually shows a call and, in a comment after it, the value it returns.
`ant doc Array.map` prints the standard library's example in that form:

```text
func Array.map[T, U](xs: Array[T], f: func(T): U): Array[U]

    Array.map returns a new array holding f applied to each element of xs, in order.

        Array.map([1, 2, 3], { x * 2 })   // [2, 4, 6]
```

The reader sees the call and its result side by side, without running anything. `Stack.pop` above
uses the same form: `s.pop()   // null`.

### Module docs and scripts

A module doc starts with a summary of the module in its own terms. A module has no fixed name, since
each importer names it, so the summary uses the name that describes it, as `Stack` does above.

A script's doc says what the program does and how to run it. `//!` lines may follow the `#!` line:

<!-- test: cmd ant doc countdown.ant -->

```ant [countdown.ant]
#!/usr/bin/env ant
//! Countdown prints the numbers from 3 down to 1.
//!
//! Run it with: ant countdown.ant

for n in [3, 2, 1] {
  println(n)
}
```

A script exports nothing, so `ant doc` prints only its doc:

```text
    Countdown prints the numbers from 3 down to 1.

    Run it with: ant countdown.ant
```

## Looking things up with ant doc

`ant doc` takes a prelude name, a module, or a module and a name:

```sh
ant doc Array.map            # a prelude function
ant doc Array                # a namespace and everything in it
ant doc stack.ant            # a module's doc and its exports
ant doc stack.ant Stack.pop  # one export of a module
```

A name is `Name` or `Owner.name`, and `ant doc` looks it up without regard to case, so
`ant doc array.MAP` finds `Array.map`. For one declaration it prints the signature and then the doc
comment, indented by four spaces. A type name prints the type that way, followed by the signature of
every function declared on it:

<!-- test: cmd ant doc temperature.ant celsius -->

```ant [temperature.ant]
// Celsius is a temperature in degrees Celsius.
export type Celsius = { degrees: float64 }

// Celsius.toFahrenheit returns c in degrees Fahrenheit.
//
//     Celsius{ degrees: 100.0 }.toFahrenheit()   // 212.0
export func Celsius.toFahrenheit(c: Celsius): float64 {
  c.degrees * 9.0 / 5.0 + 32.0
}

// Celsius.isFreezing reports whether water freezes at c.
export func Celsius.isFreezing(c: Celsius): bool {
  c.degrees <= 0.0
}
```

```text
type Celsius = { degrees: float64 }

    Celsius is a temperature in degrees Celsius.

func Celsius.toFahrenheit(c: Celsius): float64
func Celsius.isFreezing(c: Celsius): bool
```

Ask for `Celsius.toFahrenheit` to see its doc comment and example.

`ant doc` reads only the file you give it. It does not follow imports or run the type checker, so a
module with compile errors is documented like any other. It exits with status 1 when the name is not
found or when a module has no doc and exports nothing, and with 2 when the usage is wrong or the
file cannot be read or parsed.

When a name is not found, `ant doc` suggests the nearest names it knows, so a guess from another
language leads to the right one:

```text
$ ant doc Array.join
ant doc: no declaration Array.join; did you mean Path.join or String.join?
```

When no name is close but the owner exists, the line ends with where to look instead:
`ant doc Json.decode` ends with `; ant doc Json lists its members`. The REPL's `:doc` reports the
same, without the `ant doc: ` prefix.

## Mistakes to avoid

### A blank line ends the comment

A doc comment touches its declaration. A blank line between them makes the comment an ordinary one,
and `ant doc` prints the signature alone:

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

### Block comments are not docs

A JavaScript or Java habit is a `/** */` comment with `@param` tags. Only `//` lines make a doc
comment, so this one documents nothing:

<!-- test: cmd ant doc clamp.ant clamp -->

```ant [clamp.ant]
/**
 * clamp limits n to the range lo to hi.
 * @param n the number to limit
 */
export func clamp(n: int, lo: int, hi: int): int {
  Math.max(lo, Math.min(n, hi))
}
```

```text
func clamp(n: int, lo: int, hi: int): int
```

Write `// clamp returns n limited to the range lo to hi.` instead, and name the parameters in the
sentence.

### The doc goes above, not inside

Python puts a docstring inside the function. In Ant a string at the start of a body is only a value,
and a comment inside the body is a plain comment. The doc comment goes above the declaration, or
above the `export` that starts it.

### Module docs go first

`//!` lines come before the first `import` or declaration, on consecutive lines. Other comments,
such as a license, may come before or after them, but not between them. A `//!` line anywhere else,
such as after a declaration or after code on its line, is a syntax error:

```ant
let limit = 3 //! Limits caps retries.
// error: a //! module doc line must come before the first import or declaration
println(limit)
```

### Examples in comments are not checked

`ant check` ignores comments, so it does not see the code in a doc comment. An example that stops
compiling after a change stays in the doc. Run each example when you write it, and put the behavior
it shows in a test as well.

### Only exports are listed

`ant doc` lists and finds only the exported declarations of a module. A helper without `export` has
no entry, even with a doc comment, so `ant doc` exits with status 1 for it. Its doc comment still
helps whoever reads the file.

## See also

[Comments and doc comments](../language/lexical/comments.md), [`ant doc`](../tools/doc.md),
[Modules](modules.md), [Testing](testing.md), [`export`](../language/keywords/export.md)
