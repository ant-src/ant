# Library tables

A file here tests library functions one line at a time: a call, and the value it must return or
the panic it must raise.

```ant
// spec: 11 String
String.count("aaaa", "aa")          // => 2
String.cut("kvw", "=")              // => null
String.lines("a\r\nb\n")            // => ["a", "b"]
String.repeat("x", -1)              // => ◆ String.repeat needs a count of at least 0, got -1

var s = "añbñc"
String.lastIndexOf(s, "ñ")          // => 3
```

The files are for functions that map arguments to a result. A behavior with effects stays a
program under `../run/`: files, sockets, standard input and output, goroutines, `select`, and
anything that differs between runs, such as `Time.now` or `Math.random`.

## Files

A file is `<name>.ant`, one per namespace, named after it in lower case: `string.ant`,
`array.ant`, `time.ant`. A namespace that outgrows one file splits by topic into
`<name>_<topic>.ant`, such as `time_zones.ant`. The file starts with the same tag line as any
program in the corpus (`../README.md`), and may carry a `// panics:` line.

## Inputs

The file is read as the REPL reads what is typed: one input at a time, an input being one line,
or as many lines as it takes to close an open brace, bracket, string, or template. Each input is
a setup or a case.

- **A case** is an expression whose last line ends in a comment that starts with `=>`. The text
  after `=>`, with the spaces around it trimmed, is what the case expects.
- **A setup** is any other input: a declaration, an import, a statement, or a comment line. It
  runs for its effect on the inputs after it, and it fails the file if it panics.

A `=>` comment on a declaration or a statement, or on an expression of type `void`, is an error
in the file. Blank lines and comment lines are skipped.

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

`=> ◆ text` expects the case to raise a ◆ panic whose message equals `text`. A case that
expects a value fails if it panics, and a case that expects a panic fails if it returns.

The inputs after a panic still run, and they see everything the earlier inputs bound.

## How a file runs

The harness gives each file a fresh session (`ant.NewSession`), with empty standard input and no
arguments, and evaluates the inputs in order. A file that writes to standard output or standard
error fails: output is tested under `../run/`. Each file runs under the suite's context, and
files run in parallel.

Before it runs, the file is checked as a program. Declarations and imports stay at the top level,
and every other input goes into `main` in order, each case as `var _ = <expression>`.
`ant check` must report nothing. A ◆ panic is a value panic that a checked program can reach, so
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
`TestPanicCoverage` count them as they count programs. A case expecting `◆` on a call of a ◆
function of chapter 11 covers that function without a `// panics:` line. A row of the chapter 12
table, such as `int-overflow`, still needs one.

## Where a test goes

- A library function's result for given arguments, or its ◆: a case here, in its namespace's file.
- A library function's effect, output, or timing: a program under `../run/`.
- A bug in a library function: the case that shows it, added before the fix.
