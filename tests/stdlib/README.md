# Library tables

A file here tests library functions one line at a time: a call, and the value it must return or the
panic it must raise.

```ant
// spec: 11 String
String.count("aaaa", "aa")          // => 2
String.cut("kvw", "=")              // => null
String.lines("a\r\nb\n")            // => ["a", "b"]
String.repeat("x", -1)              // => panic: String.repeat needs a count of at least 0, got -1

var s = "añbñc"
String.lastIndexOf(s, "ñ")          // => 3
```

The files are for functions that map arguments to a result. A behavior with effects stays a program
under `../run/`: files, sockets, standard input and output, goroutines, `select`, and any case whose
value differs between runs. A call that does differ may still be a case when what the case computes
from it does not, as in `Crypto.uuid() == Crypto.uuid() // => false`.

The harness holds a file to this. It rejects a `go` block, a `select`, any name in the namespaces
`Channel`, `File`, `Http`, `Net` and `Os`, and the functions that wait: `Time.sleep`, `Time.ticker`,
`Context.wait`, `WaitGroup.wait`, `Mutex.lock`, `Mutex.withLock`, `Cond.wait`, `RwMutex.readLock`,
`RwMutex.writeLock`, `RwMutex.withRead` and `RwMutex.withWrite`. It looks at the parsed input, so
the words inside a string or a comment do not count.

## Files

A file is `<name>.ant`, one per namespace, named after it in lower case: `string.ant`, `array.ant`,
`time.ant`. A namespace that outgrows one file splits by topic into `<name>_<topic>.ant`, such as
`time_zones.ant`; `<name>_edges.ant` holds each function at the edges its documentation names. The
file starts with the same tag line as any program in the corpus (`../README.md`), and may carry a
`// panics:` line.

## Inputs

The file is read as the REPL reads what is typed: one input at a time, an input being one line, or
as many lines as it takes to close an open brace, bracket, string, or raw string. Each input is a
setup or a case.

- **A case** is an expression whose last line ends in `// =>` and what the case expects: the text
  after it, with the spaces around it trimmed. The spaces inside the marker are optional, so that
  `//=> 2` is a case that fails rather than a line nobody checks.
- **A setup** is any other input: a declaration, a `var`, or a statement. It runs for its effect on
  the inputs after it. It fails the file if it panics, and if it has a value, since a value nothing
  compares is a case whose `// =>` is missing; a call whose value is not wanted is bound with `var`.

The harness finds the marker in the text, not in the parse, so it may appear only where it ends a
case: not on the earlier lines of a case that spans several, not twice on one line, and not inside a
string. A string that needs those characters builds them, as `"/" + "/ =>"`. A `// =>` on a
declaration or a statement is an error, and so is one on a case that returns `void`. A void call may
still be a case that expects a panic. Blank lines and comment lines are skipped.

A file cannot import: the library needs no import, and a relative path would mean one thing to the
check and another to the session.

The file is also a program (see below), so it follows a program's rules where the REPL's are looser:
a name is declared once, and a function or type is declared before the first input that uses it.
Since setup binds names once, a value that changes from case to case is a `var` that setup assigns
again, not a second declaration. A function declared in a file is a top-level function of that
program, and the setup's `var`s are locals of `main`, so a function cannot use them; it takes what
it needs as a parameter.

## Expectations

`=> text` expects the case's value. The value is rendered as it would appear inside an array, so a
string is quoted with escapes: `=> "2"` is the string, `=> 2` the int, and `=> null` the null. The
rendering must equal `text` character for character. Floats, records, tagged unions, maps, and
`Result` values compare by their rendering too:

```ant
Math.sqrt(2.0)                      // => 1.4142135623730951
Url.parse("::")                     // => Err{ error: Error{ message: "parse \"::\": missing protocol scheme" } }
Map.fromEntries([Entry{ key: "a", value: 1 }])   // => Map{ "a": 1 }
```

`=> panic: text` expects the case to raise a panic whose message equals `text`. A case that expects
a value fails if it panics, and a case that expects a panic fails if it returns. A message that
quotes the host's own wording, such as the `parse "::"` above from Go's URL parser, is matched as it
is; a change to it is a change to the library's messages and updates the case.

The rendering is the value's and not its type's: `int8` 2 and `int` 2 both render as `2`. The type
of each case is already fixed by the call's signature, which the checker holds the file to, so a
case does not repeat it. An expression whose type cannot be inferred alone, such as `[]` or `Map{}`,
is bound in a setup `var` with its type and the case names the variable.

The inputs after a panic still run, and they see everything the earlier inputs bound.

## How a file runs

The harness gives each file a fresh session (`ant.NewSession`), with empty standard input and no
arguments, and evaluates the inputs in order. A file that writes to standard output or standard
error fails, since output is tested under `../run/`; the report a session writes for a panic does
not count, as the harness takes the panic from the error `Eval` returns. Any other error from `Eval`
fails the input: it is never something a case can expect. Each file has ten seconds under the
suite's context, so a file that never finishes fails alone; files run in parallel.

Before it runs, the file is checked as a program. Declarations stay at the top level, in the order
they appear, and every other input goes into `main` in order, each case as `var _ = <expression>`.
`ant check` must report nothing. A panic is a value panic that a checked program can reach, so it
belongs here. A kind panic belongs under `../unchecked/`.

## Failures

A failure names the file, the line where the case's comment is, the expression, what it gave, and
what was expected:

```
stdlib/string.ant:4: String.lines("a\r\nb\n")
    got:  ["a", "b", ""]
    want: ["a", "b"]
```

Paths are relative to `tests/`, where `go test` runs. The harness reports every failing case in the
file, not only the first; a file it cannot read or check stops at that problem.
`go test ./tests -run TestStdlib` runs all the files, and `-run TestStdlib/string` runs one.
`TestStdlibHarness` feeds the harness tables that are wrong in one way each, and checks that it says
so.

## Coverage

The files are part of the corpus, so `TestSpecCoverage`, `TestLibraryCoverage` and
`TestPanicCoverage` count them as they count programs. A case expecting a panic covers the function
of its outermost call, when that function is one of chapter 11 or of `stdlib.md` that panics on a
value and the message starts with its name, as every library panic's does; it covers the
`library-value` row of the chapter 12 table with it. So `Array.map(xs, { String.repeat(x, -1) })`
covers neither: `Array.map` does not panic on a value, and `String.repeat` is not the outermost
call. Nor does `Array.chunk(Array.filled(-1, 0), 1)` cover `Array.chunk`, since the panic is
`Array.filled`'s. A case that means a function calls it outermost. Any other row of the table, such
as `int-overflow`, still needs a `// panics:` line.

## Where a test goes

- A library function's result for given arguments, or its panic: a case here, in its namespace's
  file.
- A library function's effect, output, or timing: a program under `../run/`.
- A bug in a library function: the case that shows it, added before the fix.
