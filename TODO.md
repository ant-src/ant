# Work still to do

Six gaps in the language, the library and the tooling. Each one is a sketch, not a decision: the
spec (`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

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

## Binary data

**Today:** a `string` holds only scalar values, so text from outside a program is decoded with
U+FFFD (section 11, String), and bytes that are not UTF-8 do not survive it. The only byte-level
I/O is `File.readBytes` and `File.writeBytes`, which take a path. An HTTP body, the output of
`Os.exec`, and what a socket, standard input or an open file gives are strings, in both
directions: `Http.request` sends a `string` body, and a `File.Handle` has `writeLine` and
`writeText` but nothing for bytes. Before outside text was decoded, a PNG download survived in a
string by accident, and `String.toBytes` recovered it; now it is lost.

`Array[byte]` is also no carrier for bulk data. Every element is a slot of its own, so reading a
10 MB file with `File.readBytes` peaks at about 430 MB of memory, against about 90 MB for
`File.read` of the same file.

**Sketch:** a compact immutable `Bytes` type, stored as Go's `[]byte`, with byte forms of the I/O
that has only text forms today.

```
Bytes.length(b: Bytes): int
Bytes.get(b: Bytes, i: int): byte                  // panics outside 0 ..= length - 1
Bytes.slice(b: Bytes, from: int, to: int): Bytes
Bytes.concat(parts: Array[Bytes]): Bytes
Bytes.fromArray(xs: Array[byte]): Bytes
Bytes.toArray(b: Bytes): Array[byte]
String.toUtf8(s: string): Bytes
String.fromUtf8(b: Bytes): string | null           // null on invalid UTF-8; decoding with U+FFFD is another function

File.readChunk(f: File.Handle, max: int, ctx: Context | null = null): Result[Bytes | null, Error]   // null at end of file
File.writeChunk(f: File.Handle, b: Bytes): Result[void, Error]
```

- `Http.Response` and `Http.Request` gain the body as `Bytes`, and `Http.request` takes one.
- `Os.Output` gains `stdout` and `stderr` as `Bytes`, and `Os.exec` takes standard input.
- `Crypto` and `Encoding`, which take `Array[byte]` today, move to `Bytes`.

**Open questions:**
- Whether `Bytes` replaces `Array[byte]` across the library, which is a breaking change to
  `String.toBytes`, `File.readBytes`, `Crypto` and `Encoding`, or lives beside it.
- Whether a body is `string` and `Bytes` side by side, which holds two copies, a union the
  program narrows, or a choice the call makes, such as `Http.request(..., binary: true)`.
- Whether `Bytes` is mutable. Immutable shares safely across goroutines (section 9) and slices
  without copying, while a buffer that grows needs a builder.
- How `String.from`, `==`, map keys, and `Json.stringify` treat `Bytes`: a hex rendering such as
  `Bytes(0x89504e47)`, equality by content, and base64 or an error in JSON.
- What ant.js maps it to: `Uint8Array` fits, with `TextDecoder` for `fromUtf8`.
