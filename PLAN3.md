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
  one; `ant-lang.md` names the new namespace in section 11 and says that a library signature may
  carry a default. `SPEC-FEEDBACK.md` has no open entry.
- The prelude declares each function and a builtin implements each; `TestLibraryCoverage` holds
  the implementation to it, so every function is used by a program under `tests/run` tagged
  `stdlib <Namespace>`, and the checker's corpus has a program per new type.
- `scripts/coverage.ant` runs `go test` itself, the `coverage` step of the workflow is one
  command, and the README says so.
- Two new programs under `examples/`: one that starts processes, one that is a filter over
  standard input.
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
type Info = { name: string, size: int, isDir: bool, modified: int }   // modified counts as Time.now

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
name. `File.list` keeps returning names.

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

### 3.4 Time

```
type Date = { year: int, month: int, day: int, hour: int, minute: int, second: int,
              millisecond: int, weekday: int }  // UTC; month 1 to 12; weekday 0 is Sunday

Time.format(ms: int): string            // RFC 3339, UTC, milliseconds: 2026-10-22T14:03:05.123Z
Time.parse(text: string): int | null    // RFC 3339 with any offset, as milliseconds; null otherwise
Time.date(ms: int): Time.Date           // time.UnixMilli(ms).UTC()
Time.fromDate(d: Time.Date): int        // time.Date: month 13 rolls into the next year
```

Only UTC. Local time and zones need a database and a type for a zone, and no program in the
corpus or the examples wants them yet; they are a later addition over `time.LoadLocation`.

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

## 5. Phases

Each phase is one commit or a few, with the suite green under `-race` and the gate where it was.

- **Phase 21: spec.** `stdlib.md` gains the signatures of section 3 and their paragraphs;
  `ant-lang.md` section 11 lists `Path`, adds `Math.random` and the sentence on defaults. One
  commit in the spec repository. Done when every decision of section 6 is written down and the
  text does not describe the implementation's convenience anywhere.
- **Phase 22: defaults and streams.** `Builtin.Required`, the prelude test, `WithStdin`, the
  `// stdin:` trailer, `Os.stdin`, `Os.stdout`, `Os.stderr`, `File.readAll`, `File.writeText`,
  `File.flush`, with the corpus and the REPL rule. Done when a filter program runs under the
  harness with input from a trailer.
- **Phase 23: processes.** `Os.exec`, `Os.environ`, `Os.cwd`, the cancellation test, the `-run`
  flag of the coverage script, the workflow and the README. Done when CI's coverage step is the
  one command and green on both platforms.
- **Phase 24: files and paths.** The `File` additions and the `Path` namespace with their
  corpus programs. Done when `TestLibraryCoverage` has no pending entry.
- **Phase 25: time and random.** `Time.format`, `Time.parse`, `Time.date`, `Time.fromDate`,
  `Math.random`, with a round-trip test over a range of instants.
- **Phase 26: examples and documentation.** `examples/runner.ant`, a task runner that reads a
  JSON list of commands, runs each with `Os.exec`, reports status and elapsed milliseconds on
  standard error and exits 1 when one failed; `examples/sift.ant`, a filter that keeps the lines of
  standard input containing every argument. `CHANGELOG.md`, `doc.go`, the README, and the status
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
5. **`Time`**: the UTC `Date` record and RFC 3339 formatting now (recommended), or deferred.
6. **`Math.random`** in the core library now (recommended), or deferred with the rest of a random
   namespace.
7. **Deferred**: `Http.request`, regular expressions, encodings, digests, sockets, a streaming
   form of `exec` (recommended, section 3.6).
8. **The REPL** gives `Os.stdin()` an empty stream (recommended), since the REPL owns the terminal.
9. **Embedding**: no new option; hosts drop `Os` and declare their own (recommended).

Taken, carried over from the earlier conversation: `Os.exec` runs no shell, returns `Err` only
when the process cannot start, and captures both outputs.

## 7. Status

Written on 2026-10-22; no phase started. The decisions of section 6 are the user's.
