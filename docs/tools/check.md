<!-- status: reviewed -->

# ant check

`ant check` runs the type checker over the files you name and every module they import. A directory
stands for every `.ant` file under it, and with no path `ant check` checks the current directory.

```sh
ant check main.ant
ant check .
```

When there are no errors, it exits with status 0, after printing any [warnings](#warnings).
Otherwise it prints every error and their count and exits with status 1, or with 2 when a file
cannot be loaded or the usage is wrong. Each error names its rule with a code, and
`ant check -explain code` says what the code means:

```sh
ant check -explain type-mismatch
```

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

`ant test` checks the test files it runs, with their imports, before it runs them. `ant check`
checks them too: a directory includes its test files, and the checker reads each with the modules it
imports:

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

## Checking a project

Each path is a file or a directory:

- A file is checked whatever its name, so a script installed without its `.ant` suffix can be named
  directly.
- A directory is searched for `.ant` files, test files and declaration files included, through every
  directory below it except those whose names begin with `.`. A directory with no `.ant` file under
  it is an error, so a check of the wrong directory does not pass.

```sh
ant check                  # everything under the current directory
ant check main.ant lib     # one script, and every file under lib
```

A module is checked once however many of the files import it, so an error in a shared library is
reported once.

## Reading errors

Each error starts with the file, line, and column it is about, then its message, then its code in
brackets (see [Codes](#codes)). Below it come the lines of code, numbered, with the part the error
covers underlined. The last line counts the errors. For the first example above, the output is:

```text
main.ant:3:13: + needs both operands of one type, got int | null and int [operands]
   │
 3 │ println(age + 1)
   │         ^^^^^^^

main.ant:6:3: an if without else has type void when the condition is false, which cannot be used where string is expected [missing-else]
   │
 6 │   if n > 0 { "positive" }
   │   ^^

2 errors
```

When `ant check` can say how to fix an error, lines after the code say it: `= help:` with a
sentence, `= fix:` with a change that would remove the error, and `= note:` with another place that
explains it:

```text
main.ant:2:16: Array[string] has no field or function join [no-member]
   │
 2 │ let line = tags.join(", ")
   │            ^^^^^^^^^
   = help: join is String.join, which takes the array first; call it through its namespace
   = fix: String.join(tags, ...)
```

### JSON

With `-json`, `ant check` writes each error to standard output as a JSON object on a line of its
own, and nothing else. The status is the same as without it.

```sh
ant check -json .
```

```text
{"path":"main.ant","line":3,"column":13,"severity":"error","code":"operands","message":"+ needs both operands of one type, got int | null and int","start":{"line":3,"column":9},"end":{"line":3,"column":16}}
```

`line` and `column` are where the error is reported, as the text output's first line gives them.
`severity` is `"error"`, and `code` is the rule the error breaks. An error with help has `help`, one
with fixes has `fixes`, each a `title`, whether it is `auto`, and the `edits` that make it, and one
with related places has `related`. `start` and `end` are the span it underlines, with `end` just
past its last character. Columns count bytes from 1. An error about a whole file, such as one that
cannot be read, has `line` and `column` 0 and no span.

## Warnings

A warning reports code that checks but is almost always a mistake, such as a `Result` whose error is
dropped. Warnings print like errors, with `warning:` before the message, and leave the status 0:

```ant
func save(path: string): Result[void, Error] { File.write(path, "saved\n") }

save("out.txt") // warning: the Result of save is discarded, so its Err is lost
```

Each warning goes away when the code says what it means: here `try save("out.txt")` to pass the
error on, or `_ = save("out.txt")` to drop it on purpose. There are four, listed under
[Codes](#codes): `unused-result`, `unused-import`, `unused-variable`, and `named-interpolation`. A
module with an error gets no warnings until the error is fixed, since it was not checked as written.
`-strict` makes a warning fail the check as an error does.

## Errors before checking

Some errors stop a program from loading at all. `ant check` and `ant run` report them alike: a
syntax error in any module, a module that does not exist, an import cycle, and an import of a
script. `ant run` stops there. `ant check` reads past them and checks what it can read, so one run
reports every error, and then exits with status 2. An error that only follows from one of these is
left out: a name whose declaration a syntax error hid, or an error on the line of a syntax error.
Once those are fixed, `ant check` may report errors they hid.

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

Editors with Ant support run the checker and show its errors and warnings in the file.

In CI, run `ant check -strict` beside `ant fmt --check` and `ant test`, so that a warning fails the
step as an error does:

```sh
ant fmt --check .
ant check -strict .
ant test .
```

## Codes

Every error carries a code that names the rule it breaks, printed in brackets after the message and
written as `code` under `-json`. A message may be reworded from one release to the next; its code
stays. `ant check -explain code` prints the code's entry below.

### `syntax`

The source does not parse: a token, a keyword, a reserved word, an escape, or a construct out of
place.

```ant
let n = 1 as int // error: 'as' is a reserved word
```

Fixed:

```ant
let n: int = 1
```

### `missing-module`

A module given or imported cannot be read.

```ant
import "./words.ant" { parse } // error: cannot find module "./words.ant"
```

Fixed, importing a module that exists:

```ant
import "./geometry.ant" { Point }

println(Point{ x: 1.0, y: 2.0 })
```

### `import`

An import that is not allowed: an import cycle, or an import of a script, a declaration file, or,
outside a test file, a test file.

```ant
import "./report.ant" { total }
// error: cannot import "./report.ant": it is a script, made one by the statement at report.ant:2:1, and a script cannot be imported
```

Fixed, importing a library:

```ant
import "./geometry.ant" { distance, Point }

println(distance(Point{ x: 0.0, y: 0.0 }, Point{ x: 3.0, y: 4.0 }))
```

### `undefined`

A name, type, or member of a namespace is not declared.

```ant
println(lenght("ant")) // error: lenght is not defined
```

Fixed:

```ant
println(len("ant"))
```

### `not-exported`

A name is used outside its module without `export`, or an export names a private type in its
signature.

```ant [temperature.ant]
export func celsius(f: float64): float64 { (f - 32.0) * ratio() }

func ratio(): float64 { 5.0 / 9.0 }
```

```ant
import "./temperature.ant" { celsius, ratio } // error: ratio is not exported by ./temperature.ant
```

Fixed, importing only the export:

```ant
import "./temperature.ant" { celsius }

println(celsius(212.0))
```

### `redeclared`

A name is declared twice in one scope, a field twice in one record type, a declaration takes a
prelude name, or a function on a type shares its name with a field.

```ant
type Entry = { key: string } // error: Entry is a prelude name and cannot be redeclared
```

Fixed:

```ant
type Item = { key: string }
```

### `declaration-order`

A name is read before it is initialized: a later top-level variable, a variable in its own
initializer, or a later parameter in a default.

```ant
let total = price * 2
// error: price is declared later in the module; a top-level initializer or statement may refer only to the variables above it
let price = 5
```

Fixed:

```ant
let price = 5
let total = price * 2
```

### `misplaced`

A construct is outside the place it is allowed: `return` outside a function, a function on a type
not declared at the top level of the module, or a local type that mentions what it cannot.

```ant
func String.shout(s: string): string { s.toUpper() + "!" }
// error: func String.shout: String is not a type declared at the top level of this module
```

Fixed, as a plain function:

```ant
func shout(s: string): string { s.toUpper() + "!" }
```

### `wrong-kind`

A name is used as a kind of thing it is not: a namespace or a type as a value, a variable as a type,
a value called that is not a function.

```ant
let m = Math // error: Math is a namespace, not a value
```

Fixed:

```ant
let pi = Math.PI
```

### `type-mismatch`

A value's type does not fit where it is used.

```ant
let port: int = "8080" // error: cannot use string where int is expected
```

Fixed:

```ant
let port: int = "8080".toInt() ?? 80
```

### `no-member`

A type or namespace has no field or function of that name, or a function cannot be called on that
receiver.

```ant
let s = ["a", "b"].join(",") // error: Array[string] has no field or function join
```

Fixed: `join` is `String.join`, which takes the array first.

```ant
let s = String.join(["a", "b"], ",")
```

### `needs-narrowing`

A field or function exists on some members of a union but not all, so the value must be narrowed
first.

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func size(s: Shape): float64 {
  s.radius // error: Shape has no field radius on every member: narrow the value first
}
```

Fixed:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func size(s: Shape): float64 {
  if s is Circle { s.radius } else { s.side }
}
```

### `not-nullable`

`??`, `?.`, `?.(`, or `?.[` applies to a value that cannot be `null`.

```ant
let n = 5 ?? 0 // error: ?? needs a nullable left side, got int
```

Fixed:

```ant
let n = "5".toInt() ?? 0
```

### `operands`

An operator's operands do not fit it.

```ant
let total = "3" + 4 // error: + needs both operands of one type, got string and int
```

Fixed:

```ant
let total = ("3".toInt() ?? 0) + 4
```

### `arguments`

A call's arguments, or a type's type arguments, do not fit what it takes.

```ant
let saved: Result[void, Error] = Result.ok() // error: Result.ok needs an argument for value
```

Fixed: the value of a `Result[void, E]` is `void`.

```ant
let saved: Result[void, Error] = Result.ok(void)
```

### `needs-annotation`

The checker cannot infer a type, and asks for it to be written.

```ant
let names = [] // error: cannot infer the element type of an empty array; annotate it
```

Fixed:

```ant
let names: Array[string] = []
```

### `no-common-type`

The branches, clauses, elements, entries, or returns of a construct have no type in common.

```ant
let c = true
let xs = if c { [1] } else { ["a"] }
// error: the branches of this if do not form a valid type: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array; narrow first or give the whole an expected type
```

Fixed, with an expected type:

```ant
let c = true
let xs: Array[int | string] = if c { [1] } else { ["a"] }
```

### `bound`

A type argument is outside its parameter's bound, or a bound does not provide what is used.

```ant
let sorted = [true, false].sort() // error: T would be bool, which is not a member of its bound Ordered
```

Fixed, sorting by a key that is ordered:

```ant
let sorted = [true, false].sortBy({ if x { 1 } else { 0 } })
```

### `invalid-union`

A union, written or formed by substitution, breaks a rule of unions, such as two members `is` cannot
tell apart.

```ant
type Ids = Array[int] | Array[string]
// error: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array; write Array[int | string]
```

Fixed:

```ant
type Ids = Array[int | string]
```

### `type-declaration`

A `type` declaration the language does not allow: an alias that names itself, a recursion that never
ends, or a second tag for one type.

```ant
type Point = { x: int }
type Spot = Point
// error: type Spot = Point would mint a second tag for one type; declare the fields or the base type again
```

Fixed, declaring the fields again:

```ant
type Point = { x: int }
type Spot = { x: int }
```

### `spread`

A spread in a literal of a value it cannot spread, or of a record with a field the literal's type
lacks.

```ant
type Point = { x: int, y: int }

let p = Point{ ...5, y: 3 } // error: a spread needs a value of a named record type, got int
```

Fixed:

```ant
type Point = { x: int, y: int }

let origin = Point{ x: 0, y: 0 }
let p = Point{ ...origin, y: 3 }
```

### `missing-field`

A record literal leaves out a field that has no default.

```ant
type Server = { addr: string, workers: int = 4 }

let s = Server{ workers: 8 } // error: Server literal is missing the field addr
```

Fixed:

```ant
type Server = { addr: string, workers: int = 4 }

let s = Server{ addr: ":8080", workers: 8 }
```

### `missing-else`

An `if` without `else` is used where a value is needed, including a chain of `is` tests that does
not cover a union.

```ant
func label(n: int): string {
  if n > 0 { "positive" }
  // error: an if without else has type void when the condition is false, which cannot be used where string is expected
}
```

Fixed:

```ant
func label(n: int): string {
  if n > 0 { "positive" } else { "not positive" }
}
```

### `try`

`try` on a value that is not a `Result`, or where the enclosing function does not return a
compatible `Result`.

```ant
func size(path: string): int {
  let text = try File.read(path) // error: try would return Err[Error], but the function returns int
  text.length()
}
```

Fixed, returning a `Result`:

```ant
func size(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.length())
}
```

### `assign`

An assignment to a `let`, an import, or something else that cannot be assigned.

```ant
let count = 0
count += 1 // error: cannot assign to count, a let; declare it with var to reassign it
```

Fixed:

```ant
var count = 0
count += 1
```

### `lambda`

A lambda `{ ... }` whose shape does not fit its expected type, or that has no expected type.

```ant
let double = { x * 2 }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

Fixed:

```ant
let double = func(x: int): int { x * 2 }
```

### `is-test`

An `is` test the checker cannot carry out: of an interface, an alias, or a type whose type arguments
the value does not fix.

```ant
type Id = int | string

func isId(v: Id | bool): bool {
  v is Id // error: Id is an alias and has no run-time identity; test its members instead
}
```

Fixed:

```ant
type Id = int | string

func isId(v: Id | bool): bool {
  v is int || v is string
}
```

### `useless-test`

A test whose result is known, or an `else` that cannot run.

```ant
func describe(n: int | null): string {
  if n is int { "number" } else if n == null { "none" } else { "?" }
  // error: this else cannot run: the tests before it cover every case
}
```

Fixed:

```ant
func describe(n: int | null): string {
  if n is int { "number" } else if n == null { "none" }
}
```

### `duration-equality`

`==` or `!=` between two duration literals of the same length and different fields.

```ant
let same = 60s == 1m
// error: 60s == 1m compares the fields as written, and == is false: compare the lengths with Time.compareDuration(60s, 1m)
```

Fixed:

```ant
let same = Time.compareDuration(60s, 1m) == 0
```

### `markup`

Markup holds a value or an attribute that it does not take.

```ant
let user: string | null = null
let page = <p>{user}</p>
// error: markup takes a string, a number, void, Html, or an array of them, got string | null; decide what null shows, as {x ?? "none"}, or use an if
```

Fixed:

```ant
let user: string | null = null
let page = <p>{user ?? "guest"}</p>
```

### `conversion`

A conversion to a type it does not apply to.

```ant
let n = int("42") // error: int converts a number or a named primitive type over a number, got string
```

Fixed:

```ant
let n = "42".toInt() ?? 0
```

### `void-value`

`_` is read, or a call that returns `void` is bound.

```ant
let _ = 5
println(_) // error: _ is write-only and cannot be read
```

Fixed:

```ant
let five = 5
println(five)
```

### `reflect`

`type(...)` is used where it cannot be.

```ant
let n = 5
let t = type(n) // error: type(...) takes a type, not a value; n is a value
```

Fixed:

```ant
let t = type(int)
```

### `unused-result`

A warning: a `Result`, or a union with one, is discarded, so its `Err` is lost. A statement of a
block other than its last, the last of a block whose value is discarded, such as a loop's body or
the body of a function that returns `void`, and a statement of a script's top level discard their
values. A `defer` block discards its values on purpose and is not reported.

```ant
func save(path: string): Result[void, Error] { File.write(path, "saved\n") }

save("out.txt") // warning: the Result of save is discarded, so its Err is lost
```

Fixed, propagating the `Err`, or discarding it on purpose:

```ant
func save(path: string): Result[void, Error] { File.write(path, "saved\n") }

try save("out.txt")
_ = save("again.txt")
```

### `unused-import`

A warning: a name an import binds is never used, as a value or as a type.

```ant
import "./geometry.ant" { distance, Point } // warning: distance is imported but never used

println(Point{ x: 1.0, y: 2.0 })
```

Fixed:

```ant
import "./geometry.ant" { Point }

println(Point{ x: 1.0, y: 2.0 })
```

### `unused-variable`

A warning: a `let` or `var` in a function, a block, or a test is never used. A top-level declaration
is not reported, since the module or another one may use it.

```ant
func area(w: int, h: int): int {
  let half = w / 2 // warning: half is declared with let but never used
  w * h
}
```

Fixed:

```ant
func area(w: int, h: int): int {
  w * h
}
```

### `named-interpolation`

A warning: a value of a named primitive type in a string's `${...}`, which shows it as its
constructor, `UserId("u-1")`, not as its value. Markup rejects one for that reason.

```ant
type UserId = string

let id = UserId("u-1")
println("user ${id}")
// warning: id is a UserId, which a string shows as UserId(...), not as its value
```

Fixed, converting it to its base type, or `String.from` for the constructor form on purpose:

```ant
type UserId = string

let id = UserId("u-1")
println("user ${string(id)}")
println("debug ${String.from(id)}")
```

### `internal`

The checker failed on a program it should have checked. This is a bug in `ant`; please report it
with the program.

## See also

- [ant run and the REPL](run.md)
- [ant test](test.md)
- [ant fmt](fmt.md)
- [Types](../guide/types.md)
- [Modules](../guide/modules.md)
- [Advisory types](../design/advisory-types.md)
