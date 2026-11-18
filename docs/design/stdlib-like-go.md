<!-- status: reviewed -->

# Why the standard library mirrors Go

Ant's standard library follows Go's. Where Go has settled how a path is cleaned, how a duration is
written, or what a write guarantees, Ant does the same, and the namespaces are smaller than the Go
packages behind them. This page explains why Ant borrows a library instead of designing one, and
where it does not.

## What mirrors Go

```ant
Path.clean("/../a")          // "/a"
Path.clean("a/../..")        // ".."
Time.parseDuration("1m2h")   // Ok{ value: PT2H1M }
Path.match("*.TXT", "a.txt") // Ok{ value: false }
"Inf".toFloat()              // Infinity
Encoding.fromBase64("aGl=")  // Ok{ value: Bytes(68 69) }
let e = Error{ message: "unexpected end of JSON input" }
Error.wrap(e, "parse app.json").message // "parse app.json: unexpected end of JSON input"
```

Each of these answers a question that has more than one reasonable answer: whether `..` can climb
above the root, whether units may come out of order, whether a pattern ignores case on a file system
that does, whether `Inf` reads as a number, and how context joins an error's message. Ant gives Go's
answer each time. The match is on behavior, not on names: a Go programmer recognizes what a function
does, under a name in Ant's style.

| Go                              | Ant                  |
| ------------------------------- | -------------------- |
| `filepath.Clean`                | `Path.clean`         |
| `filepath.Match`                | `Path.match`         |
| `time.ParseDuration`            | `Time.parseDuration` |
| `strconv.ParseFloat`            | `String.toFloat`     |
| `fmt.Errorf("%s: %w", ctx, e)`  | `Error.wrap(e, ctx)` |
| `bufio.Writer`                  | `File.buffered`      |
| `os.CreateTemp`                 | `File.tempFile`      |
| `os.Rename`                     | `File.rename`        |
| `os.WriteFile`                  | `File.write`         |
| `regexp`, with RE2 syntax       | `Regex`              |
| `encoding/json`, `encoding/csv` | `Json`, `Csv`        |
| `os.UserHomeDir`, `os.Exit`     | `Os.home`, `Os.exit` |

## Why Go

### The hard cases are already decided

A library is mostly edge cases. Whether `..` climbs above the root, whether a file that exists keeps
its mode when it is written, what `File.tempFile("report-*.csv")` names its file, how a duration
with a fraction of a nanosecond reads: each needs a decision, and each decision is one that
someone's code comes to rely on. Go has made these decisions, documented them, and kept them for
years. Ant takes them, so the library has one consistent answer to each question instead of a new
one invented for the occasion.

### Formats that other programs share

<!-- test: env TIMEOUT=1m30s -->

```ant
let text = Os.env("TIMEOUT") ?? "30s"
let timeout = try Time.parseDuration(text)
println(timeout)
```

```text
PT1M30S
```

`1m30s` is the notation of configuration files, flags, and environment variables, and of
`ant test -timeout`. `Time.parseDuration` reads it as Go does, and every string Go's
`Duration.String` prints parses. A duration setting written for a Go program works for an Ant
program unchanged. `Regex` takes RE2 syntax, the syntax of Go's `regexp`, so a Go programmer already
knows what a pattern can and cannot say.

### No promise beyond Go's

The file guarantees are Go's, and so are their limits. `File.rename` replaces its target atomically
on Unix, so a reader sees the old file or the new one; elsewhere it is not promised to be atomic, as
Go's `os.Rename` is not. A promise Go does not make would have to be built by Ant alone, on every
platform, and kept forever.

The same rule decides what is left out. `File.write` takes `replace: false`, which creates the file
only when nothing is at the path, in one step the OS makes atomic:

```ant
try File.write("report.txt", "first")
println(File.write("report.txt", "second", replace: false))
```

```text
Err{ error: Error{ message: "open report.txt: file exists" } }
```

`File.rename` has no such parameter, since Go has no rename that refuses to replace. The `Err`
carries the OS's message and changes nothing.

### Writes behave as in Go

```ant
let out = File.buffered(Os.stdout())
try out.writeLine("buffered")
println("direct")
try out.close()
```

```text
direct
buffered
```

A write is handed to the OS before it returns, as in Go, so nothing needs a flush. `File.buffered`
is the opt-in for many small writes, Go's `bufio.Writer` as a handle, with Go's trade-off: its
output waits in the buffer until it is full, flushed, or closed, so a write that goes around it,
such as `println` here, comes out first. A buffered handle neither flushed nor closed when the run
ends loses what it holds, as a Go program that returns from `main` without `Flush` does.

## Where Ant is not Go

Ant mirrors what Go's library does, not how a Go program calls it.

**A `Result`, not a pair.** A Go function that can fail returns a value and an `error`, and the
caller checks the second:

```go
data, err := os.ReadFile(path)
if err != nil {
	return err
}
```

The same call in Ant returns one `Result`, and `try` passes the `Err` on:

```ant
func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.lines().length())
}
```

**An error is a message.** `Error` holds only `message`, and `Error.wrap` matches `%w` in the text
it writes, not in a chain: there is no `errors.Is` or `os.IsExist`. A program that treats an
existing file differently from other failures checks `File.exists` after the `Err`.

**The subject first, through `.` and `|`.** A library function takes its subject as the first
argument, so it is called through `.` on the value or chained with a pipe: `x.f(a)` is `T.f(x, a)`
for a value of type `T`.

```ant
"a,b,c".split(",")                            // ["a", "b", "c"]
" Ant ".trim() | String.toUpper()             // "ANT"
["b", "a"] | Array.sort() | String.join(", ") // "a, b"
```

**Strings count code points.** Go's `len` counts the bytes of a string. Ant's lengths, indices, and
slices count code points, and the bytes are a conversion away:

```ant
"héllo".length()           // 5
"héllo".toBytes().length() // 6
```

**Globs go further.** `Path.match` and `File.glob` take Go's `filepath.Match` syntax with `**` and
`[!...]` added, and a wildcard skips a name that begins with `.` unless `hidden: true`:

```ant
Path.match("*", ".env")               // Ok{ value: false }
Path.match("*", ".env", hidden: true) // Ok{ value: true }
Path.match("**/*.ant", "x/y/a.ant")   // Ok{ value: true }
```

**Time follows Temporal.** `Time` is modeled on JavaScript's Temporal, not on Go's `time`: exact
time and wall-clock time are different types, a zone is named and carried, and a duration keeps the
units it was given. `Time.parseDuration` reads Go's notation, but the value is a Temporal duration,
printed in ISO 8601:

```ant
let a = 90m
let b = 1h30m
Time.parseDuration("90m")  // Ok{ value: PT90M }
a == b                     // false
Time.compareDuration(a, b) // 0
```

`==` compares the fields as given, so 90 minutes and an hour and a half are two values of one
length, and `Time.compareDuration` is the comparison of lengths. In Go both are the same
`time.Duration`.

**Smaller.** The namespaces cover fewer cases than the Go packages behind them. A function is added
when a program needs it, and what is there is complete as written.

## What it costs

- **Go's quirks come along.** `Path.match` compares code points exactly on every system, so `*.TXT`
  does not match `a.txt` even on a file system that ignores case. `Time.parseDuration("0.1ns")` is
  zero, since Go drops digits below a nanosecond. `Encoding.fromBase64` ignores the bits past the
  last byte, so `aGl=` and `aGk=` decode to the same bytes. Each is defensible, and each is Go's
  choice rather than one made for Ant.
- **Unix first.** The file guarantees are Unix's, as Go's are. On Windows a file mode is Go's
  reading of the read-only attribute, and who may read a file is decided by an access control list
  Ant does not set.
- **Gaps.** A smaller library has no function for some jobs a Go package covers. You write it in Ant
  from the pieces that are there.
- **Two traditions.** A Go programmer finds Go in `File`, `Path`, and `Os`, and JavaScript's
  Temporal in `Time`, which has to be learned as its own model.

## See also

[Coming from Go](../coming-from/go.md) maps Go's syntax and idioms to Ant's. Reference:
[`Path.clean`](../lib/Path/clean.md), [`Path.match`](../lib/Path/match.md),
[`Time.parseDuration`](../lib/Time/parseDuration.md), [`Time`](../lib/Time/index.md),
[`File.write`](../lib/File/write.md), [`File.rename`](../lib/File/rename.md),
[`File.buffered`](../lib/File/buffered.md), [`Error.wrap`](../lib/Error/wrap.md),
[`Regex`](../lib/Regex/index.md). The design notes [Why no exceptions](errors-are-values.md) and
[One obvious way](one-way.md) cover related decisions.
