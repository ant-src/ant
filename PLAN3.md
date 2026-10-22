# Plan 3: the world outside

Ant can read and write files, fetch over HTTP, sleep, read its arguments and environment, and
exit. It cannot start a process, read its own standard input, write to standard error, make a
directory, take a path apart, or say what time it is in words. `scripts/coverage.ant` shows the
gap: the gate is an Ant program, but the `go test` that produces its input still runs from the
workflow. This plan adds the missing pieces to `stdlib.md` first and to the implementation second,
taking each from what Go's standard library offers, since every builtin is a thin wrapper over a
Go package and the shape of the Go function is the best evidence of what a program needs.

## 1. Where we start

`stdlib.md` has five namespaces: `File` (whole files, listing, handles with lines and `pipe`),
`Http` (`get` and `post`), `Time` (milliseconds, sleep, a monotonic clock), `Os` (`args`, `env`,
`exit`) and `Json`. Its one rule is that an operation that can fail returns a `Result` with
`Error`, and nothing panics on bad input but `Json.stringify` on a cycle.

In the implementation each namespace is a record of `*interp.Builtin` values built in
`internal/lib`, declared in Ant notation in `internal/prelude/prelude.ant`, and the checker types a
call from that declaration. A builtin has a name and parameter names and takes every argument:
named arguments work, defaults do not, because `Builtin` has no place for them. `print` writes
through one buffered writer that blocking builtins flush. The module has `WithStdout`,
`WithStderr`, `WithArgs`, `WithEnv` and `WithoutNamespaces`, and no way to give a program its
standard input.

The one agreed addition is `Os.exec`, with the shape recorded when the coverage gate was written:
no shell, the program and its arguments, `Err` only when the process cannot start, a nonzero
status as `Ok`, output captured.

## 2. Definition of done

- `stdlib.md` lists every new function with its signature, its comment and its ◆ where it has
  one; `ant-lang.md` names the new namespace in section 11, says that a library signature may
  carry a default, and gives the time values their kind, equality and rendering.
  `SPEC-FEEDBACK.md` has no open entry.
- The prelude declares each function and a builtin implements each; `TestLibraryCoverage` holds
  the implementation to it, so every function is used by a program under `tests/run` tagged
  `stdlib <Namespace>`, and the checker's corpus has a program per new type.
- `scripts/coverage.ant` runs `go test` itself, the `coverage` step of the workflow is one
  command, and the README says so.
- Two new programs under `examples/`: one that starts processes, one that is a filter over
  standard input.
- Temporal's documented examples, gaps and overlaps included, run as corpus programs with the
  documented results.
- The suite is green under `-race` on both platforms, the gate stays at 90% overall and 80% per
  package, and what the corpus cannot reach (the standard-input option, a child killed when the
  run is cancelled) has a Go test.
- `CHANGELOG.md` lists the additions; `doc.go` and the README mention the streams.

## 3. What Go provides and what Ant takes

Each signature below is proposed spec text. The Go function behind it is named so the mapping is
checkable; the semantics follow Go's unless the comment says otherwise.

### 3.1 Os

```
type Output = { status: int, stdout: string, stderr: string }

Os.exec(program: string, args: Array[string] = [], dir: string | null = null,
        env: Map[string, string] | null = null, stdin: string = ""): Result[Os.Output, Error]
Os.environ(): Map[string, string]                // os.Environ
Os.cwd(): Result[string, Error]                  // os.Getwd
Os.stdin(): File.Handle                          // os.Stdin, or what the host gave the run
Os.stdout(): File.Handle                         // os.Stdout; shares the buffer of print
Os.stderr(): File.Handle                         // os.Stderr
```

`Os.exec` is `os/exec` with `CommandContext`: no shell, so `args` are passed as they are and
`program` is found through `PATH` when it has no path separator. The child inherits the parent's
environment and working directory unless `env` or `dir` is given; a map is the child's whole
environment, as `Cmd.Env` is, so a program that wants to add one variable starts from
`Os.environ()`. `stdin` is what the child reads. Both outputs are captured in full and taken as
text the way `File.read` takes a file. The call blocks the calling goroutine until the child has
exited. `Err` is returned when the process cannot start: the program is not found, `dir` does not
exist, the run was cancelled. A nonzero exit is `Ok` with its `status`; a process that a signal
ended has status `-1`, as `ExitCode` reports it. Cancelling the run kills the child.

The three streams are `File.Handle`s, so `File.readLine(Os.stdin())`, `File.readAll(Os.stdin())`,
`File.writeLine(Os.stderr(), msg)` and `File.pipe(Os.stdin(), Os.stdout())` need nothing new.
`File.close` on a stream has no effect. A read of standard input flushes standard output first, so
a prompt is seen before the program waits.

### 3.2 File

```
type Info = { name: string, size: int, isDir: bool, modified: Time.Instant }

File.stat(path: string): Result[File.Info, Error]                // os.Stat
File.mkdir(path: string): Result[void, Error]                    // os.MkdirAll: parents too
File.removeAll(path: string): Result[void, Error]                // os.RemoveAll
File.rename(from: string, to: string): Result[void, Error]       // os.Rename
File.walk(dir: string): Result[Array[string], Error]             // filepath.WalkDir, recursive
File.tempDir(prefix: string): Result[string, Error]              // os.MkdirTemp: a fresh directory
File.append(path: string): Result[File.Handle, Error]            // os.OpenFile, O_APPEND|O_CREATE
File.readAll(f: File.Handle): Result[string, Error]              // io.ReadAll: the rest of the file
File.writeText(f: File.Handle, text: string): Result[void, Error]   // writeLine without the newline
File.flush(f: File.Handle): Result[void, Error]                  // bufio.Writer.Flush
```

`File.walk` lists files only, as paths relative to `dir`, in lexical order, and does not follow
symbolic links; a directory is walked, not listed. `File.remove` stays for one file or an empty
directory, and `File.removeAll` is the recursive form, named so the dangerous one is the longer
name. `File.list` keeps returning names. `modified` is an `Instant` of section 3.4, which is why
the files phase follows the time phases.

### 3.3 Path

A new namespace over `path/filepath`, pure functions that follow the host's conventions: `/` on
Unix, `\` and drive letters on Windows.

```
Path.join(parts: Array[string]): string          // filepath.Join: cleans the result
Path.base(p: string): string                     // filepath.Base
Path.dir(p: string): string                      // filepath.Dir
Path.ext(p: string): string                      // filepath.Ext: ".ant", or "" when there is none
Path.clean(p: string): string                    // filepath.Clean
Path.isAbs(p: string): bool                      // filepath.IsAbs
Path.abs(p: string): Result[string, Error]       // filepath.Abs: needs the working directory
Path.rel(base: string, target: string): Result[string, Error]   // filepath.Rel
```

It lives in `stdlib.md` with the I/O namespaces because its results depend on the host, and
section 11 of `ant-lang.md` adds it to the list.

### 3.4 Time, modeled on Temporal

The first draft of this section had a UTC record and RFC 3339 strings: what a script needs today
and what every language has regretted, since it gives a date-time with no zone, arithmetic that
ignores daylight saving, and a local time that depends on the machine. JavaScript replaced `Date`
with Temporal, which reached stage 4 in March 2026 and ships in Firefox 139, Chrome 144 and
Node 26. It is the best current design, built from thirty years of `Date`, `java.util.Date` and
`moment` mistakes, and Ant takes its model whole, in Ant's notation, over Go's `time` package.
Only the parts that need data Go does not have (other calendars, locale formatting) are left out.

The model, and what it fixes:

- **Exact time and wall-clock time are different types.** `Instant` is a moment, with no zone
  and no calendar; `PlainDate`, `PlainTime` and `PlainDateTime` are what a wall clock shows, with
  no zone, for a birthday, an alarm, or a meeting before its zone is known; `ZonedDateTime` is an
  instant seen in a zone and is the only type that answers "what time is it here". Converting a
  plain value to an exact one needs a zone, said out loud; nothing assumes UTC or the machine's
  zone.
- **A zone is an IANA name or a fixed offset**, and a `ZonedDateTime` carries the name, not the
  offset, so adding a day to a Saturday in March in New York gives Sunday at the same wall-clock
  time, with the offset changed. Its string carries both, as RFC 9557 has it:
  `2024-11-03T01:00:00-04:00[America/New_York]`, so the exact moment survives a round trip even
  where the wall-clock time is ambiguous.
- **Calendar units and exact units are added differently.** Years, months, weeks and days move the
  wall clock, then the result is placed in the zone; hours and smaller move the instant. A day has
  23, 24 or 25 hours when it has to be, or 23.5 on Lord Howe Island, and `hoursInDay` says which.
- **Gaps and overlaps are resolved by a rule the program chose.** A wall-clock time in a
  daylight-saving gap or overlap has zero or two instants; the `disambiguation` option is
  `compatible` (the later instant in a gap, the earlier in an overlap, what every other library
  does by default), `earlier` or `later`. Ant has no `reject`: `possibleInstants` returns the zero,
  one or two instants, and a program that wants to reject tests the length.
- **A `Duration` keeps what it was given.** Ten fields from years to nanoseconds, all of one
  sign; 90 minutes stays 90 minutes until `roundDuration` balances it. Years, months and weeks
  have no fixed length, so `total`, `roundDuration` and `compareDuration` need a `relativeTo`
  point for them; a day is 24 hours unless `relativeTo` is a `ZonedDateTime`, whose days have
  their daylight-saving hours. Two durations add when neither has years, months or weeks.
- **Out-of-range fields are constrained, not rolled over.** August 31 plus one month is
  September 30, as Temporal's `overflow: "constrain"` default has it, not October 1, which is
  what Go's `AddDate` and JavaScript's `Date` give.
- **Values are immutable, always valid, and compare by value.** There is no setter; `with`
  returns a new value. `Time.compare` orders them. A value renders as its RFC 9557 string and
  `Json.stringify` writes it as a JSON string, which is what Temporal's `toJSON` does.
- **Nanosecond precision**, and Temporal's range of ±100 million days from the epoch, which Go's
  `time.Time` holds. `epochNanos` as an `int` covers 1677 to 2262 only, so it panics outside;
  `epochMillis` always fits.

The six types and `Zone` are opaque prelude types declared through the namespace, like
`File.Handle`, and all functions live in `Time`, since Ant has no methods and a union-typed
parameter is how one function serves the three dated types. Each carries its own tag, which `is`
tests, so the unions below are legal where a union of two sync types is not; section 4.7 and
Phase 21 carry that rule into `ant-lang.md`. The option types are enumerations, unions of empty
records, so `Time.Hour{}` is checked where Temporal's `"hour"` is a string checked at run time.

```
type Zone                                           // an IANA zone or a fixed offset
type Instant                                        // an exact moment: nanoseconds, no zone
type ZonedDateTime                                  // an instant in a zone
type PlainDate                                      // a calendar date, no zone
type PlainTime                                      // a wall-clock time, no date
type PlainDateTime                                  // a date and a time, no zone
type Duration                                       // years to nanoseconds, one sign, as given
type Point = Instant | ZonedDateTime | PlainDate | PlainTime | PlainDateTime
type Wall = ZonedDateTime | PlainDate | PlainTime | PlainDateTime     // has fields: all but Instant
type Dated = ZonedDateTime | PlainDate | PlainDateTime
type Timed = ZonedDateTime | PlainTime | PlainDateTime
type Unit = Year | Month | Week | Day | Hour | Minute | Second | Millisecond | Microsecond
          | Nanosecond
type Disambiguation = Compatible | Earlier | Later
type RoundingMode = Ceil | Floor | Expand | Trunc | HalfCeil | HalfFloor | HalfExpand | HalfTrunc
                  | HalfEven

Time.utc: Time.Zone
Time.zone(id: string): Result[Time.Zone, Error]     // "Europe/Stockholm", "UTC", "+05:30"; any case
Time.localZone(): Time.Zone                         // TZ, else the system's zone, else its offset
Time.zoneId(z: Time.Zone): string                   // the name given, in the database's case

Time.now(): Time.Instant                            // was an int of milliseconds
Time.nowIn(z: Time.Zone): Time.ZonedDateTime
Time.sleep(ms: int): void                           // unchanged
Time.monotonic(): int                               // unchanged

Time.fromEpochMillis(ms: int): Time.Instant         // ◆ beyond the range
Time.fromEpochNanos(ns: int): Time.Instant
Time.epochMillis(t: Time.Instant | Time.ZonedDateTime): int      // floored
Time.epochNanos(t: Time.Instant | Time.ZonedDateTime): int       // ◆ outside 1677 to 2262
Time.plainDate(year: int, month: int, day: int): Time.PlainDate  // ◆ no such date
Time.plainTime(hour: int = 0, minute: int = 0, second: int = 0, millisecond: int = 0,
               microsecond: int = 0, nanosecond: int = 0): Time.PlainTime      // ◆ out of range
Time.plainDateTime(date: Time.PlainDate, time: Time.PlainTime): Time.PlainDateTime
Time.duration(years: int = 0, months: int = 0, weeks: int = 0, days: int = 0, hours: int = 0,
              minutes: int = 0, seconds: int = 0, milliseconds: int = 0, microseconds: int = 0,
              nanoseconds: int = 0): Time.Duration           // ◆ mixed signs, or beyond the range

Time.parseInstant(s: string): Result[Time.Instant, Error]     // 2026-10-22T14:03:05Z, or an offset
Time.parseZoned(s: string, disambiguation: Time.Disambiguation = Time.Compatible{}):
    Result[Time.ZonedDateTime, Error]               // needs [zone]; with no offset, disambiguation
                                                    // decides; an offset the zone cannot have: Err
Time.parsePlainDate(s: string): Result[Time.PlainDate, Error]     // 2026-10-22
Time.parsePlainTime(s: string): Result[Time.PlainTime, Error]     // 14:03:05.123
Time.parsePlainDateTime(s: string): Result[Time.PlainDateTime, Error]
Time.parseDuration(s: string): Result[Time.Duration, Error]       // P1Y2M3DT4H5M6.5S, -PT1H

Time.inZone(t: Time.Instant, z: Time.Zone): Time.ZonedDateTime
Time.withZone(t: Time.ZonedDateTime, z: Time.Zone): Time.ZonedDateTime     // the same instant
Time.toZoned(t: Time.PlainDateTime, z: Time.Zone,
             disambiguation: Time.Disambiguation = Time.Compatible{}): Time.ZonedDateTime
Time.possibleInstants(t: Time.PlainDateTime, z: Time.Zone): Array[Time.Instant]
                                                    // 0 in a gap, 2 in an overlap
Time.toInstant(t: Time.ZonedDateTime): Time.Instant
Time.toPlainDate(t: Time.ZonedDateTime | Time.PlainDateTime): Time.PlainDate
Time.toPlainTime(t: Time.ZonedDateTime | Time.PlainDateTime): Time.PlainTime
Time.toPlainDateTime(t: Time.ZonedDateTime): Time.PlainDateTime
Time.startOfDay(t: Time.ZonedDateTime): Time.ZonedDateTime
Time.zoneOf(t: Time.ZonedDateTime): Time.Zone
Time.offset(t: Time.ZonedDateTime): string          // "+02:00"
Time.offsetSeconds(t: Time.ZonedDateTime): int
Time.hoursInDay(t: Time.ZonedDateTime): float64     // 23.0, 24.0 or 25.0; 23.5 on Lord Howe Island
Time.nextTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null   // the next offset change
Time.previousTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null

Time.year(t: Time.Dated): int
Time.month(t: Time.Dated): int                      // 1 to 12
Time.day(t: Time.Dated): int
Time.dayOfWeek(t: Time.Dated): int                  // 1 Monday to 7 Sunday, as ISO 8601
Time.dayOfYear(t: Time.Dated): int
Time.weekOfYear(t: Time.Dated): int                 // the ISO week, 1 to 53
Time.yearOfWeek(t: Time.Dated): int                 // differs from year around New Year
Time.daysInMonth(t: Time.Dated): int
Time.daysInYear(t: Time.Dated): int
Time.inLeapYear(t: Time.Dated): bool
Time.hour(t: Time.Timed): int
Time.minute(t: Time.Timed): int
Time.second(t: Time.Timed): int
Time.millisecond(t: Time.Timed): int
Time.microsecond(t: Time.Timed): int
Time.nanosecond(t: Time.Timed): int

Time.with[T: Time.Wall](t: T, year: int | null = null, month: int | null = null,
                        day: int | null = null, hour: int | null = null, minute: int | null = null,
                        second: int | null = null, millisecond: int | null = null,
                        microsecond: int | null = null, nanosecond: int | null = null,
                        disambiguation: Time.Disambiguation = Time.Compatible{}): T
                                                    // constrained; ◆ a field the type lacks
Time.add[T: Time.Point](t: T, d: Time.Duration): T  // ◆ years, months, weeks or days on an
                                                    // Instant; a PlainTime ignores them
Time.subtract[T: Time.Point](t: T, d: Time.Duration): T
Time.until[T: Time.Point](from: T, to: T, largestUnit: Time.Unit | null = null): Time.Duration
                                                    // null: seconds for an Instant, hours for a
                                                    // ZonedDateTime or PlainTime, days otherwise;
                                                    // ◆ a unit the type lacks, or Day or larger
                                                    // between ZonedDateTimes in different zones
Time.since[T: Time.Point](t: T, other: T, largestUnit: Time.Unit | null = null): Time.Duration
                                                    // negate(until(t, other)): months count from t
Time.round[T: Time.Instant | Time.Timed](t: T, smallestUnit: Time.Unit, increment: int = 1,
                                         mode: Time.RoundingMode = Time.HalfExpand{}): T
                                                    // ◆ Year, Month or Week; Day on an Instant or
                                                    // a PlainTime; an increment that does not
                                                    // divide the next unit, or is not 1 for Day
Time.compare[T: Time.Point](a: T, b: T): int        // -1, 0, 1; zoned values by instant alone
Time.part(d: Time.Duration, unit: Time.Unit): int   // the field as given, unbalanced
Time.sign(d: Time.Duration): int
Time.negate(d: Time.Duration): Time.Duration
Time.abs(d: Time.Duration): Time.Duration
Time.addDuration(a: Time.Duration, b: Time.Duration): Time.Duration
                                                    // days of 24 hours, balanced to the larger
                                                    // largest unit; ◆ years, months or weeks
Time.compareDuration(a: Time.Duration, b: Time.Duration, relativeTo: Time.Dated | null = null): int
                                                    // ◆ years, months or weeks without relativeTo
Time.total(d: Time.Duration, unit: Time.Unit, relativeTo: Time.Dated | null = null): float64
                                                    // ◆ years, months or weeks, in d or as unit,
                                                    // without relativeTo
Time.roundDuration(d: Time.Duration, smallestUnit: Time.Unit | null = null,
                   largestUnit: Time.Unit | null = null, increment: int = 1,
                   mode: Time.RoundingMode = Time.HalfExpand{},
                   relativeTo: Time.Dated | null = null): Time.Duration
                                                    // balances up to largestUnit; ◆ no unit, or
                                                    // years, months or weeks without relativeTo
```

`Time` is computation, not I/O, so it panics where the core library would: a field that cannot
be, `Time.plainDate(2026, 2, 30)`, is a ◆ like `int8(300)`, while parsing text from the world
returns `Result`. The constructors reject what `with` and `add` constrain, as Temporal's
constructors do beside its `from`: `Time.plainDate(2026, 13, 1)` panics and
`Time.with(d, month: 13)` is December. A program holding a year, a month and a day from the world
checks the month against 1 to 12 and the day against `Time.daysInMonth(Time.plainDate(year,
month, 1))` before it constructs, or parses the text it was given; `stdlib.md` says so.

`until` and `since` are both there because they differ: months are counted from the first
argument, so with months as the largest unit `Time.since(mar31, feb29)` is `P1M` and
`Time.until(feb29, mar31)` is `P1M2D`, as in Temporal. Two zoned values at the same instant in
different zones are `0` under `compare` and unequal under `==`, as Temporal's `compare` and
`equals` have it. `<` on time values is a compile error, since the comparison operators take
numbers and strings; `Time.compare` is the ordering. A bound needs a single member, so a value
typed `Time.Dated` passes to `Time.year` and not to `Time.add` until an `is` test narrows it, and
`stdlib.md` shows that narrowing once.

Where a result's wall-clock time falls in a gap or an overlap, `add` and `subtract` on a
`ZonedDateTime` resolve it as `Compatible`; `with` and `round` keep the value's offset when the
new wall-clock time can have it and resolve as `Compatible` otherwise, so `Time.with(t, day: 3)`
from November 2 keeps `-04:00` and from November 4 keeps `-05:00`, each a valid offset for 01:05
on November 3 in New York. That is Temporal's `offset: "prefer"`, the one place its `offset`
option is taken.

A `Zone` compares by the zone it names: `Time.zone("Asia/Calcutta")` and
`Time.zone("Asia/Kolkata")` are `==`, as Temporal canonicalizes them, and `Time.zoneId` returns
the name each was given, in the database's case. An offset zone never equals a named one. A
`Duration` is bounded as Temporal's: years, months, weeks and days each below 2^32, and the time
part below 2^53 seconds in all, so `Time.duration` and a balancing that leaves those bounds are ◆.

Not taken, each a later addition that changes nothing above: calendars other than ISO 8601 (Go
has no others; `[u-ca=iso8601]` is accepted when parsing and never written), `PlainYearMonth` and
`PlainMonthDay`, locale-aware output (`toLocaleString` needs CLDR data Go does not ship), a
format with a pattern (Temporal has none: compose from the fields), the `reject` disambiguation
(`possibleInstants` serves it), and Temporal's `offset` option beyond the `prefer` of `with` and
`round`: a parsed offset that disagrees with its zone is an error, Temporal's own default.

### 3.5 Math.random

`Math.random(): float64`, a value in `[0, 1)` from `math/rand/v2`, goes into section 11 of
`ant-lang.md` beside `Math.floor`, as JavaScript places it: a random source is not I/O, and an
integer in a range is one line on top of it. No seeding, so a program cannot ask for a repeatable
sequence; a generator with a seed can come later as a sync-like object.

### 3.6 Considered and left out

- `os.Chdir` and `os.Setenv`: process-wide state that every goroutine sees change. `dir` and
  `env` on `Os.exec` cover the uses.
- Streaming the child's output, or a child that inherits the terminal: a second form of `exec`
  with the outputs as handles. Captured output serves the coverage script and the examples; the
  streaming form can be added without changing `Os.exec`.
- Signals, `os.Getpid`, `os.Hostname`: no program asks for them.
- Sockets and servers (`net`, `net/http.Server`): a plan of their own, with a type for a
  connection.
- Regular expressions over `regexp`, base64 and hex over `encoding`, digests over `crypto`: pure
  namespaces, each small, each a plan of its own once a program wants one.
- An `Http.request` for other methods, headers on `get`, and a timeout: deferred until a program
  needs more than `get` and `post`.

## 4. Mechanisms

### 4.1 Defaults on builtins

`Os.exec` is the first library function with optional parameters, and the natural way to write
the call is `Os.exec("go", ["test", "./..."], dir: root)`. Section 6 already allows a default on
a parameter and a named argument in a direct call through a namespace, so the spec needs one
sentence in section 11: a signature in the library may carry a default.

In the implementation the prelude line declares the default, `args: Array[string] = []`, and the
checker already reads `Param.Default` from the declaration, so a call that omits the argument
checks today once the parser accepts a default on a bodiless declaration (to verify; the
parameter parser is shared). `Builtin` gains `Required int`; `callBuiltin` passes `nil` for an
omitted trailing parameter instead of panicking, and the builtin substitutes its own default, a
fresh array or map each call. `internal/lib/prelude_test.go` checks that `Required` agrees with
the declaration's defaults, as it checks names today.

### 4.2 Standard streams

`ant.WithStdin(io.Reader)` joins the options, default `os.Stdin`, and the `ant` command passes
its own. `Os.stdin()` returns one `Handle` per instance over that reader; `Os.stdout()` a handle
whose writer is the instance's print buffer under the same lock, so `print` and `writeText`
interleave in program order; `Os.stderr()` a handle over the run's stderr writer. The handle gets
a `std` flag that makes `close` a no-op. A read on standard input runs in a goroutine and the
builtin selects on it and the run's context, so a cancelled run is not held by a terminal; the
goroutine ends when input arrives or the process exits. In the REPL, standard input belongs to
the REPL's own line reader, and `Os.stdin()` is at end of file.

### 4.3 Child processes

`exec.CommandContext(in.Context(), program, args...)` with `Dir`, `Env`, `Stdin` from a string
reader, and two buffers; `WaitDelay` of a second so a child that ignores the kill does not hold
the run forever. The instance flushes its stdout before starting the child, as it does before
sleeping, so what the program printed precedes anything the program prints from the child's
output afterwards. The Go test for cancellation starts `sleep 30` through a program, cancels the
context, and checks the call returns promptly with `Err`.

### 4.4 The coverage script and CI

`scripts/coverage.ant` takes a `-run` flag before the profile path. With it the script runs
`go test -coverpkg=./... -coverprofile=<profile> ./...` through `Os.exec`, writes the child's
stdout to standard output and its stderr to standard error, and exits with the child's status when
it is not zero; then it gates as it does today. `TestCoverageScript` keeps running the script
without the flag on its synthetic profile, and gains a case that the flag is parsed. The workflow
step becomes `go run ./cmd/ant run scripts/coverage.ant -run cover.out 90 80`; the README and
`PLAN2.md` section 3.7 follow.

### 4.5 Embedding

A host that runs scripts it does not trust must not hand them `Os.exec`. `WithoutNamespaces("Os")`
already removes the whole namespace, and a host can give back `args` and `env` under the same name
through `WithNamespace`, since a host namespace is declared in Ant notation like the prelude's;
`doc.go` gets the recipe. No finer option until a host asks for one.

### 4.6 Tests

Corpus programs under `tests/run`, one per namespace, tagged `stdlib Os`, `stdlib File`,
`stdlib Path`, `stdlib Time`, plus `11 Math and Bits` for `Math.random`. The `Os` program runs
`go version` and `go nonsense`, which the suite can rely on since it runs under `go test`, and
checks `Ok` with status 0 and 2 and `Err` for a program that does not exist. `dir`, `env` and
`stdin` are checked through a small Go program written to a `File.tempDir` and started with
`go run`, which echoes its working directory, one variable and its input; that is portable where
`pwd` and `cat` are not. The harness needs a way to
feed standard input to a program: a `// stdin:` trailer in `tests/README.md`, read by the runner
into `WithStdin`. The checker's corpus gets a program per new record type and one for the defaults
on `Os.exec`: an omitted `args`, a named `dir`, a wrong type for `env`.

### 4.7 Time values

A Go package `internal/temporal`, with no dependency on the interpreter, holds the six types,
RFC 9557 parsing and formatting, zone resolution, calendar arithmetic with constraint, the
difference algorithm with a largest unit, rounding, and duration balancing and totals. An
instant is epoch seconds and nanoseconds, and the arithmetic runs on those, since `time.Duration`
spans 292 years and `Sub` saturates beyond it while Temporal's range is 270,000 years each way;
`time.Time` serves for a zone's offset at an instant, `In` and `ZoneBounds`. A zone is a
`*time.Location`; a plain value is its fields. Go's `time.Date` does not promise which instant it
picks in a gap or overlap, so the package finds the candidates itself: the offsets in force just
before and after the wall-clock time, from `ZoneBounds`, each tried as `wall - offset` and kept
when the zone agrees, which gives zero, one or two instants and is Temporal's
`GetPossibleInstantsFor`. `ZoneBounds` also stops where only the abbreviation changes, so
`nextTransition` skips a bound that keeps the offset. The package has its own tests, with
Temporal's documented examples as cases: New York on 2024-03-10 and 2024-11-03, a day added
across each, an hour added across the overlap, `with` on each side of the overlap, August 31 plus
a month, the combined duration that lands on `2023-01-26T17:41:03.008-05:00`, 90 minutes balanced
to `PT1H30M`, the month-end `since` and `until`, Lord Howe's half-hour day.

In the runtime the values are one new kind, `time`, beside `sync`: immutable, compared by value
by `==` and as map keys, rendered as the RFC 9557 string, a `Zone` as its identifier, written by
`Json.stringify` as a JSON string, tested by `is` through their tags. The checker sees seven
opaque prelude types, as it sees `File.Handle`, each with a tag that is its discriminator in a
union. `ant-lang.md` gets one row in the values table of section 12, a clause in the equality
rule of section 5, `is` accepting a time type in section 5, the discriminator in section 4's
union rule, and a bullet under `String.from`; the rest is `stdlib.md`.

The zone database is Go's: `LoadLocation` reads the system's, and `cmd/ant` imports `time/tzdata`
so the command carries its own copy where there is none, on Windows and in minimal containers, at
450 KB. A host that embeds the module does the same import if it wants that. Go has no list of
zone names, so `LoadLocation` takes `europe/stockholm` on macOS, whose file system ignores case,
and rejects it on Linux, and it takes `Local` and the empty string; and Go has no link table, so
`Asia/Calcutta` and `Asia/Kolkata` are two locations. `internal/temporal` therefore carries a
generated table, `zones_gen.go`, of every zone name with its primary name, from the IANA
database's `Zone` and `Link` lines, and of CLDR's `windowsZones.xml`, which maps Windows zone
names to IANA names; a `go generate` program under `internal/temporal` rebuilds it from the two
published files and the output is committed. `Time.zone` resolves a name through the table, in
any case, and loads the primary; `==` on zones compares primaries. `Time.localZone()` reads `TZ`,
then the target of `/etc/localtime` on Unix, and on Windows the registry's `TimeZoneKeyName`
through the table, and falls back to a fixed-offset zone of the moment's local offset, which is
right now and wrong across the next change.

`Time.now()` returning an `Instant` instead of an `int` changes three corpus programs and the
examples that time themselves, which move to `Time.monotonic()`.

## 5. Phases

Each phase is one commit or a few, with the suite green under `-race` and the gate where it was.

- **Phase 21: spec.** `stdlib.md` gains the signatures of section 3 and their paragraphs, the
  `Time` chapter rewritten; `ant-lang.md` section 11 lists `Path`, adds `Math.random` and the
  sentence on defaults; section 4 gives a time type its tag as discriminator, section 5 lets `is`
  test one and adds the equality clause, section 12 the values row, and `Json.stringify` the
  string form. One commit in the spec repository.
  Done when every decision of section 6 is written down and the text does not describe the
  implementation's convenience anywhere.
- **Phase 22: defaults and streams.** `Builtin.Required`, the prelude test, `WithStdin`, the
  `// stdin:` trailer, `Os.stdin`, `Os.stdout`, `Os.stderr`, `File.readAll`, `File.writeText`,
  `File.flush`, with the corpus and the REPL rule. Done when a filter program runs under the
  harness with input from a trailer.
- **Phase 23: processes.** `Os.exec`, `Os.environ`, `Os.cwd`, the cancellation test, the `-run`
  flag of the coverage script, the workflow and the README. Done when CI's coverage step is the
  one command and green on both platforms.
- **Phase 24: time values.** `internal/temporal` with the types, the zone table and its
  generator, parsing, rendering, fields, conversions and `compare`; the `time` kind in the
  runtime and the checker; `Time.now` changed and the corpus and examples moved; `cmd/ant`
  imports `time/tzdata`. Done when the Temporal examples for construction, parsing and gaps run
  as corpus programs with the documented results.
- **Phase 25: time arithmetic.** `duration`, `add`, `subtract`, `until`, `since`, `round`,
  `with`, `total`, `addDuration`, `compareDuration`, `roundDuration`, the transitions;
  `Math.random` beside it. Done when the arithmetic examples run with the documented results and
  the package's own tests cover each rounding mode.
- **Phase 26: files and paths.** The `File` additions and the `Path` namespace with their
  corpus programs. Done when `TestLibraryCoverage` has no pending entry.
- **Phase 27: examples and documentation.** `examples/runner.ant`, a task runner that reads a
  JSON list of commands, runs each with `Os.exec`, reports status and elapsed milliseconds on
  standard error and exits 1 when one failed; `examples/sift.ant`, a filter that keeps the lines of
  standard input containing every argument; `examples/agenda.ant`, which reads meetings as
  wall-clock times with zones from JSON and prints each in the reader's zone with the days until
  it, across a daylight-saving change. `CHANGELOG.md`, `doc.go`, the README, and the status
  section here.

## 6. Decisions

Open, each with the recommendation the plan is written to:

1. **Optional arguments on `Os.exec`** through defaults on builtins (recommended), rather than a
   `Command` record whose every field a literal must list, or two functions.
2. **The child's environment**: a map given as `env` is the whole environment, as Go, Python and
   Node have it (recommended), rather than a set of additions to the parent's.
3. **The standard streams** as `File.Handle`s returned by `Os.stdin`, `Os.stdout` and
   `Os.stderr` (recommended), rather than new functions such as `Os.readLine` and `Os.eprint`.
4. **`Path`** as its own namespace (recommended), rather than functions under `File`.
5. **`Time`** modeled on Temporal, as section 3.4 (recommended), rather than the UTC record and
   RFC 3339 strings of the first draft, or nothing until later.
6. **`Math.random`** in the core library now (recommended), or deferred with the rest of a random
   namespace.
7. **Deferred**: `Http.request`, regular expressions, encodings, digests, sockets, a streaming
   form of `exec` (recommended, section 3.6).
8. **The REPL** gives `Os.stdin()` an empty stream (recommended), since the REPL owns the terminal.
9. **Embedding**: no new option; hosts drop `Os` and declare their own (recommended).
10. **The shape of `Time`**: one namespace of functions with union-typed parameters,
    `Time.year(t: Time.Dated)` (recommended), rather than functions on each type reached as
    `Time.ZonedDateTime.year(t)`, or top-level prelude types like the sync types, which would
    take `Instant`, `Duration` and five more names in every module.
11. **Time values** as a new immutable kind compared by value and rendered as RFC 9557 strings
    (recommended), rather than records, which can be mutated into an invalid state and cannot
    carry a zone and its fields without redundancy.
12. **Options as enumerations**, `Time.Hour{}` and `Time.Later{}` (recommended), rather than
    strings checked at run time as Temporal has them.
13. **Panics for impossible fields** in constructors, `Result` for parsing, constraint in `with`
    and `add` (recommended), rather than `Result` from every constructor; the spec gives the
    `daysInMonth` check for fields from the world.
14. **Nanoseconds and Temporal's range** (recommended), rather than milliseconds, which is what
    `Time.now()` returns today.
15. **Zones as `Time.Zone` values**, validated once by `Time.zone` (recommended), rather than a
    string on every call that can fail.
16. **`time/tzdata` embedded** in the `ant` command (recommended), rather than relying on the
    system's database.
17. **Left for later** in `Time`: other calendars, `PlainYearMonth` and `PlainMonthDay`, locale
    output, a pattern format (recommended, section 3.4).
18. **A generated zone table** in `internal/temporal`, from the IANA database and CLDR's Windows
    mapping, so names resolve the same on every platform, aliases compare equal and Windows has a
    local zone (recommended, section 4.7), rather than Go's lookup as it is, which is
    case-sensitive on Linux only, tells aliases apart, and names the Windows zone `Local`.

Taken, carried over from the earlier conversation: `Os.exec` runs no shell, returns `Err` only
when the process cannot start, and captures both outputs.

## 7. Status

Written on 2026-10-22, the Time section rewritten around Temporal the same day and corrected
after a second agent's review: the unions needed a discriminator, `since` is not `until`
reversed, `hoursInDay` is a float, days need no `relativeTo`, `modified` is an `Instant`, and the
zone table of decision 18 was added. No phase started. The decisions of section 6 are the user's.
