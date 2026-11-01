# Work still to do

Four gaps in the language, the library and the tooling. Each one is a sketch, not a decision: the
spec (`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

## Flags

**Today:** `Os.args()` returns the raw arguments, and every program parses them by hand.

**Sketch:** a `Flag` namespace that parses `Os.args()` against a declared set of flags and prints
usage on `-h` or `--help`.

```
var flags = Flag.parse(Os.args(), [
  Flag.string("addr", ":8080", "address to listen on"),
  Flag.int("workers", 4, "number of workers"),
  Flag.bool("verbose", false, "log every request"),
])
// Result[Flag.Parsed, Error]

Flag.getString(p: Flag.Parsed, name: string): string
Flag.getInt(p: Flag.Parsed, name: string): int
Flag.getBool(p: Flag.Parsed, name: string): bool
Flag.rest(p: Flag.Parsed): Array[string]           // the positional arguments
```

**Open questions:**

- Getting a typed result without reflection. Lookups by name check nothing until the program runs,
  and a misspelled name would be a run-time panic. A record whose fields declare defaults (section
  4, Field defaults) would be checked, but filling it from `Os.args()` needs a way to iterate a
  record's fields.
- The syntax accepted: `--name=value`, `--name value`, `-n`, grouped short flags, `--` to stop, and
  `--no-verbose`.
- Subcommands such as `tool serve --addr :80`.
- What `--help` prints, and whether a parse error exits with status 2 as Go's `flag` does or is left
  to the program.

## Native code through Go

**Today:** `ant run` walks the tree, which keeps the interpreter simple but slow, and there is no
faster way to run a program.

**Sketch:** an `ant build main.ant -o main` command. It checks the program, generates a Go package
from the checked tree, and runs `go build` on it. The output is one static binary. The generated
code links against the runtime the interpreter already uses (`value`, `internal/lib`, `temporal`),
so the library is written once, and Go supplies goroutines, the garbage collector and the scheduler.

- Records become Go structs. Named primitives become defined types. A union becomes an interface, or
  a tagged struct where the members are few and small.
- Functions become Go functions. Closures become Go closures. Generic functions become Go generics,
  or `value.Value` where Go's generics cannot express them, such as a constraint over a union.
- `go`, `Channel`, `select`, `defer` and `panic` map onto their Go counterparts. `select` needs
  `reflect.Select` only when the number of heads isn't fixed.
- An Ant panic becomes a Go panic carrying the same message, and the `ant` exit status and stderr
  format stay the same.

**Open questions:**

- The runtime packages are `internal`, so the generated code has to build inside the `ant` module,
  in a temporary directory with a `replace`, or the runtime must move to public packages with a
  stable API.
- Whether a Go toolchain on the build machine is acceptable, or `ant build` ships one.
- Where typed Go code and `value.Value` meet: crossing at every library call costs boxing, and
  deciding how far typed code reaches is most of the speed.
- Parity: every program in `tests/run` must give the same output, stderr and exit status under
  `ant run` and `ant build`, so the corpus harness runs both.
- Debugging: positions in panics must point at the `.ant` source, not the generated Go. Go's
  `//line` directives can do this.

## Interfaces

**Today:** "There are no interfaces" (section 4, Generics). A generic function can bound `T` by a
record shape, but cannot ask for functions on `T`, so code over "anything printable" or "anything
with an area" passes the functions in by hand.

**Sketch:** an interface lists functions a type must have in its namespace (section 6, Functions on
types). There is no method call, so the required "methods" are the `T.name` functions, and a type
satisfies the interface structurally, as in Go, without declaring it.

```
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

func describe[T: Shape](s: T): string {
  "${T.name(s)}: ${T.area(s)}"
}

describe(Circle{ radius: 2.0 })   // Circle.area and Circle.name exist, so Circle is a Shape
```

**Open questions:**

- Go's model or type classes: is `Shape` only a bound on a type parameter (static dispatch through
  `T.area`), or also a type of its own, so `Array[Shape]` can hold circles and rects and dispatch at
  run time by the tag? The second needs `ant run` to find `Circle.area` from a value's tag.
- Prelude types: their namespaces are closed, so `int` can never satisfy a user interface unless the
  rule changes.
- Required fields alongside functions, which anonymous record types already cover.
- Whether `String.from`, equality and ordering become interfaces a user type can implement.

## Globbing

**Today:** `File.list` and `File.walk` return every name, and a script filters them by hand. There
is no `*.txt`.

**Sketch:**

```
Path.match(pattern: string, name: string): bool              // pure: does name match pattern
File.glob(pattern: string): Result[Array[string], Error]     // the existing paths that match, in lexical order
```

**Open questions:**

- The syntax: `*`, `?`, `[a-z]`, and whether `**` crosses directories (Go's `filepath.Match` has no
  `**`; zsh, bash's `globstar`, Python's `glob(recursive=True)` and Node's `fs.glob` do), and
  `{a,b}` alternatives.
- Whether a leading `*` matches dot files, which the shell does not.
- Links: whether `**` descends through them, which `File.walk` does not.
- Case on case-insensitive file systems, and `\` as an escape versus the Windows separator.
- No match: an empty array, not the pattern back as the shell gives it.
