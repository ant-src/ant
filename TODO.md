# Work still to do

Fifteen gaps in the language, the library and the tooling. Each one is a sketch, not a decision: the
spec (`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

## Test framework

**Today:** the language has no way to write tests. The repository's own corpus uses `// expect`
trailers that compare output, which suits the implementation, not a user's program.

**Sketch:** an `ant test` command and a `Test` namespace.

```
// math_test.ant
func testAdd(t: Test): void {
  Test.equal(t, add(1, 2), 3)
}
```

```
Test.equal[T](t: Test, got: T, want: T): void      // records a failure that shows both values
Test.true(t: Test, cond: bool, message: string = ""): void
Test.fail(t: Test, message: string): void          // records a failure and goes on
Test.stop(t: Test, message: string): never         // records a failure and ends this test
Test.skip(t: Test, reason: string): never
```

**Open questions:**

- How tests are found: `*_test.ant` files and functions named `test...`, as in Go, or a
  `test "name" { }` declaration.
- Whether a test may start goroutines that outlive it. A panic in a test can be a failure of that
  test only: the runner calls each test through `Panic.catch` (section 8, Catching a panic).
- Filtering (`-run pattern`), running tests in parallel, the output format, and the exit status.
- Whether test files can reach a module's unexported names (section 10).

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
  and a misspelled name would be a run-time panic. A record of defaults the parser fills in would be
  checked, but needs a way to iterate a record's fields.
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

## Default values for records

**Today:** a record literal must list exactly the fields of its type (section 4, Records), so a
record with many fields is verbose to build, and adding a field breaks every literal.

**Sketch:** as in Go, a field left out of a literal takes the zero value of its type: `0`, `0.0`,
`""`, `false`, `null` for a nullable type, an empty array, map or set. An unknown field stays a
compile error, so typos are still caught.

```
type Server = { addr: string, workers: int, verbose: bool, tags: Array[string] }

var s = Server{ addr: ":8080" }   // workers: 0, verbose: false, tags: []
```

**Open questions:**

- Fields with no zero value: a record type, a union that is not nullable, a type parameter, a
  function type, a sync object. Either such a field must always be listed, or the type has no zero
  value and Go's rule does not carry over.
- Declared defaults instead of, or as well as, zero values: `type Server = { workers: int = 4 }`,
  evaluated per literal like parameter defaults (section 6). This would also answer the record of
  defaults in Flags.
- Whether a missing field should stay an error under a flag or lint, since it catches a forgotten
  field after one is added.
- A zero value for a whole type, `Server{}`, and whether `ant run` (which keeps no field types)
  needs the field types at run time to fill in zeros.

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
  `${T.name(s)}: ${T.area(s)}`
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

## Shebang line

**Today:** a script runs only as `ant run script.ant`. A `#!` line is a syntax error (section 2), so
an Ant script cannot be made executable and run as `./script.ant`, as a Python or shell script can.

**Sketch:** a first line that starts with `#!` is ignored, as in JavaScript and Python. Only the
first line of the file, and only from its first byte.

```
#!/usr/bin/env -S ant run
//! Syncs ~/.ssh/cerrion-config with on-prem.conf.

println(Os.args())
```

**Open questions:**

- `#!/usr/bin/env ant` would run `ant script.ant`, which is not a command today. Either the line
  needs `env -S ant run`, or `ant <file.ant>` becomes a short form of `ant run`.
- Whether `ant fmt` keeps the line as written, and where the module doc's `//!` lines may start
  after it (section 2 puts them before the first `import` or declaration).
- Whether a `#!` line in a library is an error, since only a script can be run (section 10).

## Running a command on the terminal

**Today:** `Os.exec` captures the child's output whole and feeds it a fixed `stdin`, so a script
cannot hand the terminal to `ssh`, `git`, `make`, an editor or a pager. A wrapper script that ends
by starting such a program cannot be written in Ant.

**Sketch:** a call whose child shares the program's three standard streams and returns its exit
status.

```
Os.run(program: string, args: Array[string] = [], dir: string | null = null,
       env: Map[string, string] | null = null, ctx: Context | null = null): Result[int, Error]
```

**Open questions:**

- The name, and whether it is a parameter of `Os.exec` instead (`stdin`, `stdout` and `stderr` each
  captured or inherited, or a `File.Handle`), as Go's `exec.Cmd` has it.
- Signals: while the child runs, Ctrl-C should reach the child and not end the script, as a shell
  does. How this meets `Os.interrupt()`.
- Flushing the program's buffered output before the child writes to the same stream.
- Whether a `syscall.Exec`-style call that replaces the process is wanted too; it skips deferred
  blocks, as `Os.exit` does.

## Exit status with deferred blocks

**Today:** a script exits with `0`, or `1` on an `Err` (section 10, Scripts). Any other status needs
`Os.exit`, which skips deferred blocks, so a script that passes on a child's status, as a shell
wrapper does, leaves its temporary files behind or removes them by hand before exiting.

**Sketch:** one of:

- the script body may return an `int`, the exit status, after its deferred blocks run;
- an `Os.exitAfterDefers(status)`-style call that unwinds the main goroutine first.

**Open questions:**

- Whether `return 3` in a script body reads as an exit status or as a mistake, given that today
  `return 5` is a compile error.
- What happens to other goroutines, as with `Os.exit`.

## Temporary files and permissions

**Today:** `File.tempDir` exists but there is no temporary file, and no way to set a file's mode, so
a script that writes a secret cannot keep it `0600`. `File.Info` has no mode either, so a script
cannot ask whether a directory is writable, as `[[ -w /dev/shm ]]` does.

**Sketch:**

```
File.tempFile(prefix: string, dir: string | null = null): Result[string, Error]   // created empty with mode 0600; the host's temporary directory by default
File.chmod(path: string, mode: int): Result[void, Error]
type Info = { name: string, size: int, isDir: bool, modified: Time.Instant, mode: int }
```

**Open questions:**

- Whether `tempDir` promises mode `0700` too.
- Modes on Windows, where only the read-only bit means anything.
- A mode parameter on `File.write` and `File.create` as well as, or instead of, `chmod`, which
  leaves a moment when the file has the default mode.

## Platform

**Today:** a script that does something different on macOS and Linux runs `uname` through `Os.exec`.

**Sketch:**

```
Os.platform(): string   // "darwin", "linux", "windows", as Go's GOOS
Os.arch(): string       // "arm64", "amd64", as Go's GOARCH
```

**Open questions:**

- Go's names or a union type of literals; what the JS backend reports.

## The program's name

**Today:** `Os.args()` leaves out the program path, so a usage line cannot name the script, as
`$(basename "$0")` does, and the name is written into the message by hand.

**Sketch:** `Os.program(): string`, the path of the script as given to `ant run`.

**Open questions:**

- What it is in the REPL, under `ant build`, and when run through a shebang line.

## Writing to standard error

**Today:** `File.writeLine(Os.stderr(), msg)` is the only way, and a warning or a usage line is
common enough in scripts that it is long.

**Sketch:** `eprint` and `eprintln` in the prelude, as `print` and `println` to standard error.

**Open questions:**

- Whether the two belong in the prelude or as `Os.eprintln`, and how they interleave with buffered
  standard output.

## Pushing several values

**Today:** `Array.push` takes one value, so building an argument list of pairs, such as `-o X`,
needs `xs = Array.concat(xs, ["-o", x])`.

**Sketch:** `Array.pushAll[T](xs: Array[T], ys: Array[T]): void // mutates`, or a variadic `push`.

**Open questions:**

- Ant has no variadic functions, so `pushAll` is the smaller change.

## Formatting argument lists

**Today:** `ant fmt` puts each element of a long array literal on a line of its own, so
`["-o", "IdentitiesOnly=yes", "-o", ...]` turns into one value per line and the pairs are lost.

**Open questions:**

- Keep a multi-line literal's line breaks as the author wrote them, filling each line, as `gofmt`
  does for composite literals, rather than one element per line.

## Command output as text

**Today:** a call to a command for its output checks `status`, reads `stdout`, and trims the final
newline every time, which a small helper does in each script.

**Sketch:**

```
Os.output(program: string, args: Array[string] = [], ...): Result[string, Error]   // stdout without its final newline; Err on a nonzero status, with the status and stderr in the message
```

**Open questions:**

- How much of `stderr` goes in the message, and whether a caller that needs the status uses
  `Os.exec` instead.
