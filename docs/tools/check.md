<!-- status: reviewed -->

# ant check

`ant check` runs the type checker over a file and every module it imports.

```sh
ant check main.ant
```

When the program has no errors, it prints nothing and exits with status 0. Otherwise it prints
every error and their count and exits with status 1, or with 2 when the file cannot be loaded or
the usage is wrong.

## What it checks

`ant check` reports every compile error in the program: a value of the wrong type, a nullable value
used before a check for `null`, an `if` without `else` where a value is needed.

```ant
let ages = Map{ "Ann": 31 }
let age = ages.get("Bob")
println(age + 1) // error: + needs both operands of one type, got int | null and int

func label(n: int): string {
  if n > 0 { "positive" }
  // error: an if without else has type void when the condition is false, which cannot be used where string is expected
}
```

The checker reads code that nothing calls, too. `ant run` would run this program and print `hi`, but
`ant check` rejects it:

```ant
func unused(): int {
  "oops" // error: cannot use string where int is expected
}

println("hi")
```

### Imported modules

The checker follows every `import`, so one command checks the whole program. A library checks on its
own too:

<!-- test: cmd ant check geometry.ant -->

```ant [geometry.ant]
export type Point = { x: float64, y: float64 }

export func distance(a: Point, b: Point): float64 {
  let dx = a.x - b.x
  let dy = a.y - b.y
  Math.sqrt(dx * dx + dy * dy)
}
```

A call across modules is checked against the exported signature, here a `float64` where `distance`
takes a `Point`:

```ant
import "./geometry.ant" { distance, Point }

let origin = Point{ x: 0.0, y: 0.0 }
println(distance(origin, 5.0)) // error: cannot use float64 where Point is expected
```

### Test files

`ant test` runs tests without checking them, so check a test file with `ant check` before you run
it. The checker reads the test file and the modules it imports:

```ant [geometry_test.ant]
import "./geometry.ant" { distance, Point }

test "distance to itself is zero" {
  let p = Point{ x: 1.0, y: 2.0 }
  assert(distance(p) == 0.0) // error: distance needs an argument for b
}
```

### Declaration files

`ant check` also accepts a declaration file, a `.d.ant` file of signatures. Every type a signature
names must exist, every default must fit its parameter's type, each name is declared once, and no
declaration may take a prelude name:

```ant [shout.d.ant]
// Shout.loud returns text in capitals.
func Shout.loud(text: string): Strin // error: unknown type Strin

// Shout.loud returns text in capitals, with marks.
func Shout.loud(text: string, marks: int): string // error: Shout.loud is already declared
```

## Reading errors

Each error starts with the file, line, and column it is about, and then its message. Below it come
the lines of code, numbered, with the part the error covers underlined. The last line counts the
errors. For the first example above, the output is:

```text
main.ant:3:13: + needs both operands of one type, got int | null and int
   │
 3 │ println(age + 1)
   │         ^^^^^^^

main.ant:6:3: an if without else has type void when the condition is false, which cannot be used where string is expected
   │
 6 │   if n > 0 { "positive" }
   │   ^^

2 errors
```

## Errors before checking

Some errors stop a program from loading at all. `ant check` and `ant run` report them alike, and
neither goes further: a syntax error in any module, a module that does not exist, an import cycle,
and an import of a script.

```ant [report.ant]
let total = 3
println("total: ${total}")
```

```ant
import "./report.ant" { total }
// error: cannot import "./report.ant": it is a script, made one by the statement at report.ant:2:1, and a script cannot be imported

println(total)
```

A module with a statement at its top level is a script, and no module can import a script. Move the
shared code into a library that both files import.

## ant check and ant run

`ant run` does not check types. It parses the program and runs it, so a program with compile errors
still runs. It panics when it reaches an operation that cannot work on the values it holds, or a
name that does not exist. The first example above, run with `ant run`, stops at line 3 with a panic,
followed by its trace:

```text
panic: + needs two numbers of one kind or two strings, got null and int
```

`ant check` finds the same mistake without running anything, on every path, including the paths that
one run does not take. No flag makes `ant run` check first. Run `ant check` yourself, or let your
editor and CI run it.

The checker does not trace calls, though. A top-level `let` or `var` read through a function call
before its initializer has run passes `ant check`, and `ant run` panics when it reaches the read:

```ant
let greeting = welcome()
let name = "Ann"

func welcome(): string {
  "hello, ${name}" // panic: name is read before its initializer has run
}

println(greeting)
```

Declare `name` before the line that calls `welcome`.

## In editors and CI

Editors with Ant support run the checker and show its errors in the file.

In CI, run `ant check` on each script beside `ant fmt --check` and `ant test`. A non-zero status
fails the step:

```sh
ant fmt --check .
ant check main.ant
ant check geometry_test.ant
ant test .
```

`ant check` takes one file. Check each script, and each test file, with a command of its own; a
library that one of them imports is checked with it.

## See also

- [ant run and the REPL](run.md)
- [ant test](test.md)
- [ant fmt](fmt.md)
- [Types](../guide/types.md)
- [Modules](../guide/modules.md)
- [Advisory types](../design/advisory-types.md)
