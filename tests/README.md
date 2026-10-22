# The test corpus

Every program here is derived from a sentence of the specification, and says which one. The Go
tests in this directory run the programs and keep the corpus traceable; `go test ./tests -v -run
TestSpecCoverage` prints the matrix of headings against the number of programs naming each.

## Kinds of program

| Directory | What the harness proves |
| --- | --- |
| `run/` | the program's output, exit status and standard error match its trailer, and `ant check` reports nothing |
| `unchecked/` | the program runs as its trailer says, and `ant check` rejects it: these are the kind panics the checker rules out |
| `check/` | `ant check` reports exactly the errors the inline markers name, and nothing else |
| `../examples/` | real programs, run like `run/` |

Files under a directory named `lib` are modules the programs import, not tests.

## Tags

A program starts with a tag line naming the headings of the specification it proves:

```
// spec: 4 Unions; 4 Tagged unions
// spec: 10
// spec: stdlib Json
```

A tag is a chapter number followed by a subsection heading copied from the spec (backticks
removed), a chapter number alone for the chapter's own text, or `stdlib` and a namespace of
`stdlib.md`. Several tags are separated by `;`. A program that triggers a runtime panic also
carries a `// panics:` line naming a row of the chapter 12 table by its key (see `panicRows` in
`spec_test.go`) or a ◆ function of chapter 11:

```
// panics: index-range
// panics: Array.slice; library-value
```

`TestSpecCoverage`, `TestLibraryCoverage` and `TestPanicCoverage` fail when a tag names nothing
in the spec, when a heading, library name or panic has no program, or when a pending list in
`spec_test.go` still names something that is now covered.

## Trailers

A program under `run/`, `unchecked/` or `examples/` ends with a trailer that starts with a line
reading `// expect`:

```
// expect
// out: one expected line of standard output, in order; "// out:" alone is an empty line
// exit: 1          the exit status as the ant command reports it; default 0
// err: text        standard error must contain this text; may repeat
// args: a b c      command-line arguments; $TMP becomes a fresh temporary directory and $HTTP
//                  the URL of a test server (see echoServer in conformance_test.go)
// stdin: line      one line of standard input, in order, for Os.stdin(); may repeat
```

A program under `check/` marks each expected error on its line:

```
var n: int = "x"      // ERR: string is not int
f(1, 2)               // ERR: first text ;; second text
```

The marker text must be contained in the message. A line `// error: <line>: <text>` at the top
names an error on another line, for errors the checker reports on a declaration rather than a
use.

## Where to add a test

- A rule the checker enforces: `check/`, in the file for its chapter.
- A behavior with observable output: `run/`.
- A kind panic the checker rejects: `unchecked/`, one program per panic.
- A ◆ panic that a checked program can reach: `run/`, with `// exit: 1` and an `// err:` line.
- A bug found by an example or a fuzzer: the smallest program under `run/` or `check/` that shows
  it, added before the fix.
