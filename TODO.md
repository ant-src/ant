# Work still to do

Five gaps in the language and library. Each one is a sketch, not a decision: the spec
(`spec/ant-lang.md`, `spec/stdlib.md`) is written first, and the code and tests follow it.

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
Random.int(r: Random, from: int, to: int): int     // from inclusive, to exclusive; ◆ to <= from
Random.shuffle[T](r: Random, xs: Array[T]): void   // mutates, in place
Random.pick[T](r: Random, xs: Array[T]): T | null  // null for an empty array
```

**Open questions:**
- The spec has to name the generator (PCG, SplitMix64, xoshiro) and how `int` and `shuffle` draw
  from it. Otherwise the same seed gives different sequences in the interpreter and in `antc`.
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
  and a misspelled name would be a ◆. A record of defaults the parser fills in would be checked,
  but needs a way to iterate a record's fields.
- The syntax accepted: `--name=value`, `--name value`, `-n`, grouped short flags, `--` to stop, and
  `--no-verbose`.
- Subcommands such as `tool serve --addr :80`.
- What `--help` prints, and whether a parse error exits with status 2 as Go's `flag` does or is
  left to the program.
