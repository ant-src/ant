# Work still to do

Eight gaps in the language, the library and the tooling. Each one is a sketch, not a decision: the spec
(`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

## `len`

**Today:** each namespace spells the size of a value its own way: `Array.length`, `String.length`,
`Map.size` and `Channel.length`, and the `Set` sketch below adds `Set.size`. A generic helper has
no single name to call, and a reader has to remember which word each type uses.

**Sketch:** a prelude function `len`, built into the checker like the twelve conversion names, that
takes any value with a size. The namespace functions stay.

```
len(x): int    // x is an Array[T], a Map[K, V], a string, or a Channel[T]
```

- An array counts its elements, a map its entries, a string its code points (as `String.length`),
  and a channel the values buffered now (as `Channel.length`).
- Any other argument is a compile error, and a kind panic in the interpreter.

**Open questions:**
- Whether a named primitive over `string` is accepted and unwrapped, as the conversions do, while
  `String.length(id)` needs `string(id)` (section 4).
- Whether `len` is a reserved name or an ordinary prelude name a module can shadow.
- Whether it covers `Set[T]` once that lands, and whether the library adds a matching `isEmpty`
  helper or leaves `len(x) == 0`.
- Whether `len` on a string is linear, as counting UTF-8 code points is, or the runtime caches it.
- Whether to keep `Array.length` and the others at all, now that one name does the job.

## `min` and `max`

**Today:** `Array.min` and `Array.max` take an array and return `T | null`, and `Math` has no
two-argument form. Picking the larger of two values means `Array.max([a, b]) ?? a` or writing the
`max` that section 4 uses as its generics example.

**Sketch:** two prelude functions over `Ordered`, with at least two arguments so the result is never
`null`.

```
min[T: Ordered](a: T, b: T, ...rest: T): T
max[T: Ordered](a: T, b: T, ...rest: T): T
```

- `max(1, 2.0)` is a compile error, as in the section 4 example: both arguments are one `T`.
- NaN behaves as it does in `Array.min` and `Array.max`: any NaN argument gives NaN.
- A tie returns the first argument, as `Array.minBy` returns the first element with the least key.

**Open questions:**
- The language has no variadic parameters. Either they arrive with this, or `min` and `max` take
  exactly two arguments and longer lists go through `Array.min`.
- Section 4 defines its own `max` in an example. Whether a module may shadow a prelude name, and
  the example changes to another name if not.
- Whether named primitives such as `Meters` are accepted and keep their tag, as arithmetic on them
  does (section 4), or must be unwrapped first, as library functions require.
- Whether `-0.0` and `0.0` are ordered, as Go's `min` and `max` order them, or tie.
- Whether a `clamp(x, lo, hi)` joins them.

## Set

**Today:** there is no set type. Programs use `Map[T, bool]` and `Map.has`, which works but says
the wrong thing and renders as a map.

**Sketch:** a built-in `Set[T]`, keyed by `==` and kept in insertion order like `Map[K, V]`
(section 4), so a record or an array can be a member.

```
Set.new[T](): Set[T]
Set.of[T](xs: Array[T]): Set[T]                    // later duplicates are dropped
Set.size[T](s: Set[T]): int
Set.isEmpty[T](s: Set[T]): bool
Set.has[T](s: Set[T], x: T): bool
Set.add[T](s: Set[T], x: T): bool                  // mutates; true when x was new
Set.remove[T](s: Set[T], x: T): bool               // mutates; true when x was there
Set.values[T](s: Set[T]): Array[T]                 // insertion order
Set.union[T](a: Set[T], b: Set[T]): Set[T]
Set.intersect[T](a: Set[T], b: Set[T]): Set[T]
Set.difference[T](a: Set[T], b: Set[T]): Set[T]
Set.isSubset[T](a: Set[T], b: Set[T]): bool
Set.same[T](a: Set[T], b: Set[T]): bool            // identity, as Map.same
```

**Open questions:**
- Is there a literal, such as `Set{ 1, 2 }`? It fits beside `Map{ ... }`, but a bare word inside
  braces already means a name.
- How it renders, how `==` compares two sets (by members, ignoring order?), and whether `for x in s`
  works.
- What `Json.encode` does with one: an array is the obvious answer.
- Whether the mutation-after-insert rule for map keys carries over word for word.

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
