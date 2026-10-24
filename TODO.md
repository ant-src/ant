# Work still to do

Eight gaps in the language, the library and the tooling. Each one is a sketch, not a decision: the
spec (`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

## `min` and `max`

**Today:** `Array.min` and `Array.max` take an array and return `T | null`, and `Math` has no
two-argument form. Picking the larger of two values means `Array.max([a, b]) ?? a` or writing the
`max` that section 4 uses as its generics example.

**Sketch:** two prelude functions over `Ordered` that take exactly two arguments, so the result is
never `null`. Longer lists stay with `Array.min` and `Array.max`, since there are no variadic
parameters (section 6).

```
min[T: Ordered](a: T, b: T): T
max[T: Ordered](a: T, b: T): T
```

- `max(1, 2.0)` is a compile error, as in the section 4 example: both arguments are one `T`.
- NaN behaves as it does in `Array.min` and `Array.max`: any NaN argument gives NaN.
- A tie returns the first argument, as `Array.minBy` returns the first element with the least key.

**Open questions:**
- Section 4 defines its own `max` in an example. Whether a module may shadow a prelude name, and
  the example changes to another name if not.
- Whether named primitives such as `Meters` are accepted and keep their tag, as arithmetic on them
  does (section 4), or must be unwrapped first, as library functions require.
- Whether `-0.0` and `0.0` are ordered, as Go's `min` and `max` order them, or tie.
- Whether a `clamp(x, lo, hi)` joins them.

## `println`

**Today:** `print` always ends with a newline. Writing without one, such as a prompt or a progress
dot, means `File.writeText(Os.stdout(), String.from(x))`, which returns a `Result` and takes only a
string.

**Sketch:** split the two, as Go and Rust do. `print` writes the value with no newline, and the new
`println` adds one.

```
print[T](value: T): void                         // String.from(value) to standard output
println[T](value: T): void                       // String.from(value) and a newline
```

- Both write to the buffer `Os.stdout()` shares (stdlib, Os), so they interleave with it in program
  order, and a read of standard input flushes them first.

**Open questions:**
- This changes what `print` means in every existing program: 245 `.ant` files in the repository
  call it, along with the spec's examples and the `// expect` corpus. The other way round, keeping
  `print` and adding a newline-free name such as `write`, breaks nothing.
- Whether `println()` with no argument prints an empty line. Without default arguments for a generic
  `T`, that needs a second signature or `println("")`.
- Whether a standard-error pair (`eprint`, `eprintln`) joins them, or `File.writeLine(Os.stderr(),
  msg)` is enough.
- Whether `println` flushes at each newline, or output is flushed only before a read and at exit.

## Unicode

**Today:** `String` has the string-wide checks (`isDigit`, `isLetter`, `isSpace`, `isUpper`,
`isLower`), case mapping, and `toCodePoints` / `fromCodePoints`. Go's `unicode` and `unicode/utf8`
packages cover more than that: they test a single code point, name its category and script, fold
case, and measure and validate UTF-8. Ant has none of these, so a lexer, a slug function or a
case-insensitive comparison has to hand-roll tables.

**Sketch:** a `Unicode` namespace over single code points, given as the `int` scalar values
`String.toCodePoints` returns, plus a few `String` additions for the whole-string forms.

```
Unicode.isLetter(c: int): bool                     // likewise isDigit, isNumber, isSpace, isUpper,
                                                   // isLower, isTitle, isPunct, isSymbol, isMark,
                                                   // isControl, isGraphic, isPrint
Unicode.category(c: int): string                   // the two-letter general category, such as "Lu"
Unicode.inScript(c: int, script: string): bool     // "Latin", "Han", "Cyrillic", ...
Unicode.toUpper(c: int): int
Unicode.toLower(c: int): int
Unicode.toTitle(c: int): int
Unicode.simpleFold(c: int): int                    // the next code point in its case-fold orbit
Unicode.isScalar(c: int): bool                     // 0 ..= 0x10FFFF and not a surrogate
Unicode.utf8Length(c: int): int                    // 1 to 4
Unicode.version(): string                          // the Unicode version the tables follow

String.equalFold(a: string, b: string): bool       // equal under simple case folding, as Go's EqualFold
String.toTitle(s: string): string
String.byteLength(s: string): int                  // UTF-8 bytes, without String.toBytes
String.isValidUtf8(bytes: Array[byte]): bool
String.normalize(s: string, form: string): string  // "NFC", "NFD", "NFKC" or "NFKD"; panics on another form
```

- Every function that takes a code point, except `isScalar`, panics on an `int` that is not a scalar
  value.
- The tables are the Unicode version Go's `unicode` package ships, so `ant run` and `ant build` agree.

**Open questions:**
- Whether a code point gets its own type, such as a named primitive `Rune` over `int32` (section
  4), so an ordinary `int` or a byte cannot be passed by mistake. `String.toCodePoints` and `for`
  over a string would then need to change too.
- Whether categories and scripts are strings, which a typo turns into a run-time panic, or tagged
  unions the checker can see.
- Normalization is not in Go's standard library (`golang.org/x/text/unicode/norm`), nor are
  grapheme clusters and display width. Which of them belong in the library, given the tables they
  carry.
- Whether the existing `String.is...` checks keep their "every code point, false on `""`" meaning or
  are defined through `Unicode`, and whether `String.toUpper` maps one code point to one, as Go's
  does, or applies the full mappings that turn `ß` into `SS`.

## Seeded random

**Today:** `Math.random` and `Math.randomInt` cannot be seeded, and `Crypto.randomBytes` reads the
operating system. A simulation, a shuffled test fixture or a game cannot be replayed.

**Sketch:** a `Random` value made from a seed. `Math.random` stays as it is.

```
Random.new(seed: int): Random
Random.float(r: Random): float64                   // uniform in [0, 1)
Random.int(r: Random, from: int, to: int): int     // from inclusive, to exclusive; panics if to <= from
Random.shuffle[T](r: Random, xs: Array[T]): void   // mutates, in place
Random.pick[T](r: Random, xs: Array[T]): T | null  // null for an empty array
```

**Open questions:**
- The spec has to name the generator (PCG, SplitMix64, xoshiro) and how `int` and `shuffle` draw
  from it. Otherwise the same seed gives different sequences under `ant run` and `ant build`.
- Whether two goroutines sharing one `Random` race (a data race like any shared mutation,
  section 9) or take turns behind a lock.
- Whether `Random.new()` with no seed exists, seeded from the operating system.

## Error wrapping

**Today:** `Error` is `{ message: string }`. A function that adds context builds a new message by
hand (`Error{ message: "load config: " + e.message }`), and the original error is lost as a value.

**Sketch:** a cause chain on the library's error, with helpers to build and walk it.

```
type Error = { message: string, cause: Error | null }

Error.new(message: string): Error
Error.wrap(e: Error, message: string): Error               // cause is e
Error.root(e: Error): Error                                // the end of the chain
Error.chain(e: Error): Array[Error]                        // e first
Result.context[T](r: Result[T, Error], message: string): Result[T, Error]  // wraps an Err
```

**Open questions:**
- Adding a field breaks every `Error{ message: "..." }` literal unless records get field defaults,
  or the cause lives somewhere else.
- Whether a cause can be any error type, such as a user's tagged union. `cause: Error | null` keeps
  it simple but wraps only library errors.
- How an `Error` with a cause renders, and what `message` means then: only its own text, or the
  chain joined with `: ` as Go's `%w` gives.
- Whether `try` gets a form that adds context on the way out.

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
- How tests are found: `*_test.ant` files and functions named `test...`, as in Go, or a `test "name"
  { }` declaration.
- Whether a panic in a test is a failure of that test only, and whether a test may start goroutines
  that outlive it.
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
  and a misspelled name would be a run-time panic. A record of defaults the parser fills in would be checked,
  but needs a way to iterate a record's fields.
- The syntax accepted: `--name=value`, `--name value`, `-n`, grouped short flags, `--` to stop, and
  `--no-verbose`.
- Subcommands such as `tool serve --addr :80`.
- What `--help` prints, and whether a parse error exits with status 2 as Go's `flag` does or is
  left to the program.

## Native code through Go

**Today:** `ant run` walks the tree, which keeps the interpreter simple but slow, and there is no
faster way to run a program.

**Sketch:** an `ant build main.ant -o main` command. It checks the program, generates a Go package
from the checked tree, and runs `go build` on it. The output is one static binary. The generated
code links against the runtime the interpreter already uses (`value`, `internal/lib`, `temporal`),
so the library is written once, and Go supplies goroutines, the garbage collector and the
scheduler.

- Records become Go structs. Named primitives become defined types. A union becomes an interface,
  or a tagged struct where the members are few and small.
- Functions become Go functions. Closures become Go closures. Generic functions become Go
  generics, or `value.Value` where Go's generics cannot express them, such as a constraint over a
  union.
- `go`, `Channel`, `select`, `defer` and `panic` map onto their Go counterparts. `select` needs
  `reflect.Select` only when the number of heads isn't fixed.
- An Ant panic becomes a Go panic carrying the same message, and the `ant` exit status and stderr format
  stay the same.

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
