<!-- status: reviewed -->

# ant fmt

```sh
ant fmt <path>...           # rewrite each file in place
ant fmt --check <path>...   # list the files that are not formatted, and change nothing
ant fmt -                   # format standard input to standard output
```

`ant fmt` rewrites Ant source files into one layout. It has no options and no configuration file, so
every formatted file looks the same and a diff shows only what changed. It works on spacing,
indentation, line breaks, semicolons at the ends of statements, and trailing commas. It never
changes what a program means, never reorders anything, and never rewrites one construct into
another.

## Format files

Name the files and directories to format:

```sh
ant fmt main.ant
ant fmt .
ant fmt src lib/geo.ant
```

Take a file `dist.ant` written like this:

```text
type Point = { x: float64, y: float64 }

func dist( p:Point , q:Point ):float64 {
    let dx = p.x-q.x;
    let dy = p.y-q.y
    Math.sqrt(dx*dx+dy*dy)
}


let points = [Point{x: 0.0, y: 0.0},
  Point{ x: 3.0, y: 4.0 }]
println( dist(points[0], points[1]) )   // 5.0
```

After `ant fmt dist.ant`, it reads:

```ant [dist.ant]
type Point = { x: float64, y: float64 }

func dist(p: Point, q: Point): float64 {
  let dx = p.x - q.x
  let dy = p.y - q.y
  Math.sqrt(dx * dx + dy * dy)
}

let points = [
  Point{ x: 0.0, y: 0.0 },
  Point{ x: 3.0, y: 4.0 },
]
println(dist(points[0], points[1])) // 5.0
```

Indentation is two spaces per level, the semicolon at the end of a line is gone, operators have a
space on each side, and the two blank lines became one. The array literal has a line break inside
its brackets, so its first item moves to a line of its own, its closing bracket too, and the last
item gets a comma.

- A directory is walked recursively for files ending in `.ant`, in lexical order. Entries whose
  names start with `.` are skipped, and symbolic links met in the walk are not followed.
- A file named on the command line is formatted whatever its extension. A file ending in `.d.ant` is
  a declaration file and follows the same layout.
- `ant fmt` with no path is a usage error, so nothing is rewritten by accident. Format the current
  directory with `ant fmt .`.
- A file is written only when its contents change, and the new contents replace the old in one step:
  an interrupted run leaves either the old file or the new one.
- `ant fmt` reads only the files it is given. It does not follow imports and does not run the type
  checker, so a file with compile errors is formatted like any other.
- A file that does not parse is left untouched, and its syntax error is reported as
  `path:line:col: message [syntax]`. The other files are still formatted.

## Format standard input

`ant fmt -` reads a module from standard input and writes the formatted source to standard output.
This is the form an editor calls to format a buffer that is not saved:

```sh
$ echo 'let total = price*qty' | ant fmt -
let total = price * qty
```

A syntax error names `<stdin>` as its path:

```sh
$ echo 'let x = (' | ant fmt -
<stdin>:2:1: unexpected end of file
```

## Check formatting

`ant fmt --check` writes nothing. It prints the path of each file that is not formatted, one per
line, as it was named or found, and exits with status 1 if there is one. Here the directory holds
`dist.ant` from above, `geo.ant`, and a `draft.ant` that nobody has formatted yet:

<!-- test: file draft.ant
let origin = Point{x: 0.0, y: 0.0}
-->
<!-- test: cmd ant fmt --check . -->
<!-- test: exit 1 -->

```ant [geo.ant]
export type Point = { x: float64, y: float64 }

export func midpoint(p: Point, q: Point): Point {
  Point{ x: (p.x + q.x) / 2.0, y: (p.y + q.y) / 2.0 }
}
```

```text
draft.ant
```

`ant fmt --check -` prints `<stdin>` when standard input is not formatted.

`ant check` does not look at layout, so continuous integration runs `ant fmt --check` beside it. In
a GitHub Actions workflow:

```yaml
- run: ant fmt --check .
- run: ant check main.ant
- run: ant test .
```

`--check` has no diff output. To see what `ant fmt` would change, run it and then `git diff`.

| Exit status | When                                                                |
| ----------- | ------------------------------------------------------------------- |
| `0`         | every file is formatted, or was rewritten                           |
| `1`         | with `--check`, at least one file is not formatted                  |
| `2`         | the usage is wrong, or a file could not be read, parsed, or written |

A `2` outranks a `1`: a `--check` run that finds both a syntax error and an unformatted file exits
with 2.

## What it keeps

`ant fmt` keeps every choice a reader might have made on purpose.

- **Line breaks.** You decide where a statement breaks. `ant fmt` never joins two lines of an
  expression and never splits one, so there is no maximum line width. It changes line breaks only
  between statements and inside brackets.
- **Order.** Declarations, statements, record fields, arguments, and imports stay in written order.
  Imports are never sorted, since their order is the order modules are initialized.
- **Spelling.** Names, number literals such as `0xFF` and `1_000`, duration literals such as `90m`,
  the text of strings and their escapes, raw strings, and regular expressions are printed exactly as
  written. Inside a string, only the code between `${` and `}` is formatted.
- **Parentheses.** Every pair you wrote stays, even where precedence makes it redundant, and none is
  added.
- **Comments.** Every comment stays, in order, with its text. Only the space before it, its
  indentation, and spaces at the ends of its lines change.

A pipeline or an expression broken after an operator keeps its lines, each continuation line one
level deeper than the line the statement starts on:

```ant
type User = { name: string, age: int }

let users = [User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 15 }]
let adults = users
  | Array.filter({ x.age >= 18 })
  | Array.map({ x.name })
let subtotal = 40
let shipping = 5
let total = subtotal +
  shipping
println(adults)
println(total)
```

```text
["Ann"]
45
```

A list written with a break inside its brackets keeps the rows you wrote, such as a flag and its
value:

```ant
let args = [
  "-o", "IdentitiesOnly=yes",
  "-p", "2222",
]
println(args.length())
```

```text
4
```

A line break inside an argument, such as in a lambda, does not break the argument list, so the other
arguments stay on the call's line:

```ant
let doubled = Array.map([1, 2, 3], {
  x * 2
})
println(doubled)
```

```text
[2, 4, 6]
```

An `else` stays where you put it: after the `}` on the same line, or at the start of the next line,
which keeps a chain of one-line branches readable:

```ant
let n = 7
let size = if n < 5 { "small" }
else if n < 10 { "medium" }
else { "large" }
println(size)
```

```text
medium
```

## What it changes

| Before                 | After                     | Rule                                               |
| ---------------------- | ------------------------- | -------------------------------------------------- |
| `f( a,b )`             | `f(a, b)`                 | no space inside brackets, one after a comma        |
| `let a = b+c*d`        | `let a = b + c * d`       | one space around binary operators                  |
| `Point{x: 1.0,y: 2.0}` | `Point{ x: 1.0, y: 2.0 }` | one space inside the braces of a literal           |
| `"sum: ${ a+b }"`      | `"sum: ${a + b}"`         | code in an interpolation is formatted, text is not |
| `let n = 1; let m = 2` | two lines                 | outside a one-line block, `;` becomes a line break |
| `func f() { g( ) ; }`  | `func f() { g() }`        | a one-line block stays on one line                 |

Beyond these, a run of blank lines becomes one blank line, and a block or a file never starts or
ends with a blank line. Indentation is two spaces, never a tab. Lines end in `\n` and never in
spaces, and the file ends with exactly one line break. Trailing comments on consecutive lines line
up one space after the longest code:

```ant
let xs = [3, 1, 2]
xs.length() // 3
xs.sum()    // 6
```

## What it does not do

- It does not wrap long lines or join short ones.
- It does not reflow HTML markup: it reindents the lines you wrote and never moves an element's
  children onto lines of their own.
- It does not add blank lines, not even between top-level declarations.
- It does not sort imports, declarations, or record fields.
- It does not remove redundant parentheses, simplify expressions, or rename anything.
- It does not align anything except runs of trailing comments.
- It does not format Ant code in Markdown files.
- It has no options beyond `--check`, no configuration file, and no way to turn it off for part of a
  file.

## See also

- [ant check](check.md)
- [ant test](test.md)
- [ant run and the REPL](run.md)
- [Idiomatic Ant](../guide/idiomatic.md)
