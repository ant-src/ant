# Library tables

A file here tests library functions one line at a time: a call, and the value it must return or
the panic it must raise.

```ant
// spec: 11 String
String.count("aaaa", "aa")          // => 2
String.cut("kvw", "=")              // => null
String.lines("a\r\nb\n")            // => ["a", "b"]
String.repeat("x", -1)              // => panic: String.repeat needs a count of at least 0, got -1

var s = "añbñc"
String.lastIndexOf(s, "ñ")          // => 3
```

The files are for functions that map arguments to a result. A behavior with effects stays a
program under `../run/`: files, sockets, standard input and output, goroutines, `select`, and
anything that differs between runs, such as `Time.now` or `Math.random`. A file here has no `go`
block, no `select`, and no call that waits: no `Time.sleep`, `Chan.send`, `Chan.receive`,
`Context.wait`, `WaitGroup.wait`, `Mutex.lock` or the like.

## Files

A file is `<name>.ant`, one per namespace, named after it in lower case: `string.ant`,
`array.ant`, `time.ant`. A namespace that outgrows one file splits by topic into
`<name>_<topic>.ant`, such as `time_zones.ant`. The file starts with the same tag line as any
program in the corpus (`../README.md`), and may carry a `// panics:` line.

## Inputs

The file is read as the REPL reads what is typed: one input at a time, an input being one line,
or as many lines as it takes to close an open brace, bracket, string, or template. Each input is
a setup or a case.

- **A case** is an expression whose last line contains `// =>`. The text after the last `// =>`
  on that line, with the spaces around it trimmed, is what the case expects. The harness finds it
  in the text, not in the parse, so an expected value cannot itself contain `// =>`; a `// =>`
  inside a string earlier on the line is harmless, since only the last one counts.
- **A setup** is any other input: a declaration, an import, a statement, or a comment line. It
  runs for its effect on the inputs after it, and it fails the file if it panics.

A `// =>` on a declaration or a statement is an error in the file, and so is one on an
expression of type `void`, which the harness finds when the case returns no value. Blank lines
and comment lines are skipped.

The file is also a program (see below), so it follows a program's rules where the REPL's are
looser: a name is declared once, imports come before everything else, and a function or type is
declared before the first input that uses it. Since setup binds names once, a value that changes
from case to case is a `var` that setup assigns again, not a second declaration.

## Expectations

`=> text` expects the case's value. The value is rendered as it would appear inside an array, so
a string is quoted with escapes: `=> "2"` is the string, `=> 2` the int, and `=> null` the null.
The rendering must equal `text` character for character. Floats, records, tagged unions, maps,
and `Result` values compare by their rendering too:

```ant
Math.sqrt(2.0)                      // => 1.4142135623730951
Url.parse("::")                     // => Err{ error: Error{ message: "parse \"::\": missing protocol scheme" } }
Map.fromEntries([Entry{ key: "a", value: 1 }])   // => Map{ "a": 1 }
```

`=> panic: text` expects the case to raise a panic whose message equals `text`. A case that
expects a value fails if it panics, and a case that expects a panic fails if it returns. A
message that quotes the host's own wording, such as the `parse "::"` above from Go's URL parser,
is matched as it is; a change to it is a change to the library's messages and updates the case.

The rendering is the value's and not its type's: `int8` 2 and `int` 2 both render as `2`. The
type of each case is already fixed by the call's signature, which the checker holds the file to,
so a case does not repeat it. An expression whose type cannot be inferred alone, such as `[]` or
`Map{}`, is bound in a setup `var` with its type and the case names the variable.

The inputs after a panic still run, and they see everything the earlier inputs bound.

## How a file runs

The harness gives each file a fresh session (`ant.NewSession`), with empty standard input and no
arguments, and evaluates the inputs in order. A file that writes to standard output or standard
error fails, since output is tested under `../run/`; the report a session writes for a panic
does not count, as the harness takes the panic from the error `Eval` returns. Each file runs
under the suite's context, and files run in parallel.

Before it runs, the file is checked as a program. Declarations and imports stay at the top level,
in the order they appear, and every other input goes into `main` in order, each case as
`var _ = <expression>`. `ant check` must report nothing. A panic is a value panic that a checked program can reach, so
it belongs here. A kind panic belongs under `../unchecked/`.

## Failures

A failure names the file, the line where the case's comment is, the expression, what it gave,
and what was expected:

```
tests/stdlib/string.ant:4: String.lines("a\r\nb\n")
    got:  ["a", "b", ""]
    want: ["a", "b"]
```

The harness reports every failing case in the file, not only the first. `go test ./tests -run
TestStdlib` runs all the files, and `-run TestStdlib/string` runs one.

## Coverage

The files are part of the corpus, so `TestSpecCoverage`, `TestLibraryCoverage` and
`TestPanicCoverage` count them as they count programs. A case expecting a panic covers the
function of its outermost call, when that function is one of chapter 11 or of `stdlib.md` that
panics on a value, and covers the `library-value` row of the chapter 12 table with it. So
`Array.map(xs, { String.repeat(x, -1) })` covers neither: `Array.map` does not panic on a value,
and `String.repeat` is not the outermost call. A case that means `String.repeat` calls it
outermost. Any other row of the table, such as
`int-overflow`, still needs a `// panics:` line.

## Where a test goes

- A library function's result for given arguments, or its panic: a case here, in its namespace's file.
- A library function's effect, output, or timing: a program under `../run/`.
- A bug in a library function: the case that shows it, added before the fix.
