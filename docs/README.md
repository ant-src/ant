# Ant documentation

The documentation of Ant, rendered at [ant-lang.org/docs](https://ant-lang.org/docs). It lives next
to the code so that a change to `ant` and the pages it affects land together, and so that `ant`'s
own tests run every example on every change.

This file is the contract for writing a page: layout, status, style, page format, examples, and how
to test. `tests/docs_test.go` enforces the parts a test can check.

## Layout

```
docs/
  index.md        Introduction
  start/          Get started              tutorial and how-to
  tour/           Tour                     tutorial: one concept per page
  guide/          Language guide           explanation: one chapter per area
  howto/          How-to guides            how-to: one task per page
  coming-from/    Coming from              explanation: JavaScript, Go, Python
  tools/          Tools                    reference: one page per ant command
  language/       Language reference       reference: keywords, operators, literals, types
  lib/            Standard library         reference: a page per function and type
  design/         Design notes             explanation: one decision per page
```

Every page is one of four kinds: a **tutorial** teaches by doing, with no choices or digressions; a
**how-to** solves one task for a reader who already knows the language; a **reference** page
describes completely and austerely; an **explanation** gives the why. A page never mixes them.

The standard library reference has one page per prelude declaration:

- `lib/<Namespace>/<function>.md`, such as `lib/Array/map.md`; prelude functions such as `println`
  are under `lib/prelude/`.
- `lib/<Namespace>/types/<Type>.md`, such as `lib/Time/types/Zone.md`. Types have their own
  directory so that `Time.Zone` and `Time.zone` stay two files on a file system that ignores case.
  The members of a few unions share the union's page: `Result`, `Html.Node`, `Reflect.Shape`,
  `Time.Unit`, `Time.Disambiguation`, `Time.RoundingMode`.
- `lib/<Namespace>/index.md`: the namespace, with its functions grouped by job. A type named like
  its namespace, such as `Bytes` or `Mutex`, is documented here.
- `lib/guides/`: topic guides across namespaces.

## Status

The first line of every page is its status:

```md
<!-- status: skeleton -->
```

- `skeleton`: generated or stubbed, not yet written.
- `draft`: written, and its test passes.
- `reviewed`: an independent review checked it against the spec and the style below.

A skeleton that is not generated has a brief on its second line, which says what the page covers:

```md
<!-- brief: Tour step 2: var, let, inference, _. ... -->
```

Keep the brief while the page is a draft, and remove it once the page is reviewed.

## Source of truth

The spec (`github.com/ant-src/spec`: `ant-lang.md`, `stdlib.md`, and the tool documents) defines
Ant. The docs explain and show; they never state a rule the spec does not state. When a page and the
spec disagree, the page is wrong. Corner cases stay in the spec: a page covers what a reader needs
and leaves the rest out. Pages never link the spec; they are written for people using Ant, not for
people implementing it.

If an example shows that `ant` does not do what the spec says, or that the spec does not say what a
page needs, that is a bug in `ant` or a gap in the spec. Report it; do not write around it.

## Style

Active voice, present tense, second person.

| Instead of                                                   | Write                                              |
| ------------------------------------------------------------ | -------------------------------------------------- |
| The file is read by `File.read` and a `Result` is returned.  | `File.read` reads the file and returns a `Result`. |
| The program will exit when the script ends.                  | The program exits when the script ends.            |
| We can use `try` to return the error early.                  | Use `try` to return the error early.               |
| It should be noted that `let` bindings cannot be reassigned. | You cannot reassign a `let`.                       |
| Simply add `defer` and the handle is closed.                 | Add `defer` to close the handle.                   |

- **Name the actor.** The function, the checker, the runtime, or you does the verb. Passive voice
  only when the actor does not matter.
- **Present tense.** No "will" for something that always happens.
- **"You", never "we".** Tutorials and guides say "you"; how-to steps are imperative ("Open the
  file. Read each line."); reference pages are third-person statements, like the spec ("`Array.push`
  appends a value").
- **No minimizers:** simply, just, easy, obviously, of course, basically.
- **No hedging.** "Usually", "should", "might" only when behavior really varies, and then say when.
- **Short sentences, one idea each.** The point first, conditions after.
- **Code first.** Show the example, then explain it.
- **One term per concept**, from the word list. Plain American English, sentence-case headings, no
  emoji, no exclamation marks.

| Use                       | Not                                       |
| ------------------------- | ----------------------------------------- |
| Ant (the language)        | ANT, ant                                  |
| `ant` (the command)       | Ant CLI, the ant tool                     |
| record                    | struct, object, class                     |
| function on a type        | method                                    |
| union, tagged union       | sum type (once, on the coming-from pages) |
| enumeration               | enum                                      |
| panic                     | exception, throw                          |
| error, `Result`           | exception                                 |
| goroutine                 | thread, task, async function              |
| script, library           | main module, package                      |
| type checker, `ant check` | compiler                                  |
| compile error             | type error, check error                   |
| interpreter, `ant run`    | VM, runtime engine                        |

## Page format

### Reference pages

Function, type, keyword, operator, literal, and built-in type pages share one format, after the
Wolfram Language documentation: mostly examples, from the first call a reader writes to the corner
cases and the tasks the function serves.

A `lib/` page starts with a generated part, between `<!-- generated from ... -->` and
`<!-- end generated -->`: the title, the signature, and the doc comment from
`internal/prelude/prelude.d.ant`. Never edit it; change the doc comment and regenerate. Everything
after it is written by hand.

The second-level headings are these, in this order, each only when it has content:

| Section                  | Holds                                                                                       |
| ------------------------ | ------------------------------------------------------------------------------------------- |
| Details                  | A few short bullets a user needs and the examples do not make obvious. No spec links.       |
| Basic examples           | The two or three calls a newcomer writes first. Required once a page is a draft.            |
| Scope                    | Every shape of input: empty, Unicode, `null`, records, unions, function values, `.`, pipes. |
| Named arguments          | Each named parameter with a default, such as `limit:` or `ctx:`.                            |
| Applications             | Small real tasks the function solves, a few lines each.                                     |
| Properties and relations | Equivalences and invariants, written with `assert`.                                         |
| Possible issues          | Mistakes and surprises, with the panic or compile error they produce.                       |
| Neat examples            | Optional: something surprising or pretty.                                                   |
| See also                 | Related functions, as links.                                                                |

Details is short: usually two to five bullets, often fewer. Keep what changes how someone uses the
function: what an empty or edge input gives, when it panics, whether it changes its argument or
returns a new value, whether it is lazy, the order it calls a callback in when that matters. Leave
out what the signature already says, what every function shares (inference, which function values
can be passed, how `.` resolves, assignability rules), and corner cases few readers hit. Each bullet
must still be true by the spec. When nothing is worth saying, leave Details out.

Third-level headings inside a section are free. Each example teaches something the others do not,
and the examples use the variety of the language on purpose: lambdas with `x`, named functions,
records, unions, `null`, calls through `.`, pipes. A reader browsing `Array.map` also learns how Ant
code looks. Declare with `let`, and with `var` only a name the example reassigns, so the keyword
tells the reader which names change.

### Other pages

A tour page is one concept and one editable program, about three minutes of reading. A guide chapter
explains how an area fits together and when to pick what. A how-to is one complete program for one
task, with numbered steps when there are several. A coming-from page is a table of equivalents and
then the differences that bite. A design note argues one decision.

### Links

Link pages with relative paths: `[Array.filter](filter.md)`,
`[for](../../language/keywords/for.md)`. The test checks that every relative link names a file under
`docs/`, spelled with its exact case, which macOS would forgive and Linux would not. Never link the
spec: link the docs page that explains the topic instead. A link that starts with `/` names a page
of the site outside the docs, such as `/vscode`.

## Examples

Every ` ```ant ` block is tested; there is no way to leave one out. A block is either a block of
cases or a program.

### Cases

An expression followed by a `//` comment is a case: the comment is the value the expression renders
as, as the REPL shows it. A block with a case is a block of cases, like a notebook of inputs and
outputs. The block is read as the REPL reads input, so a case may span lines; its comment goes at
the end of its last line.

```ant
Array.map([1, 2, 3], { x * 2 })              // [2, 4, 6]
Array.map(["a", "bb"], { String.length(x) }) // [1, 2]
let xs = [3, 1, 2]
Array.sort(xs) // [1, 2, 3]
```

- Lines without a comment are setup, such as a `let` or a `type`, and run in order.
- A comment after a declaration, an assignment, a loop, or a line inside braces is an ordinary
  comment, not a case. A comment after any other expression is a case, so `println(x) // prints 3`
  is a case whose value is void, and fails. Put explanations on their own line.
- A case may not wait, start a goroutine, or touch `File`, `Http`, `Net`, `Os`, or `Channel`, and
  its value must be the same on every run. Write those examples as programs.
- A case that panics is written `// panic: <message>`, with the whole message:
  `[1][3] // panic: index 3 out of range for an array of length 1`.
- `ant fmt` aligns the comments of consecutive lines; keep its alignment.

### Programs

Any other block is a complete program, run as `main.ant` with `ant check` and then `ant run`. A
` ```text ` block right after it is its standard output, compared exactly except for one final line
break, so output that does not end a line can be shown. Without one, the program must still exit
with status 0. Problems are reported at the lines of the page.

````md
```ant
let name = "Ant"
println("Hello, ${name}")
```

```text
Hello, Ant
```
````

A block with a file name, ` ```ant [lib/geo.ant] `, is written into the page's directory instead of
run, so other blocks of the page can import it. Every named block and `file` directive of a page is
written before any block runs, in one directory per page. A named `.ant` file, a test file included,
is checked with `ant check`; it runs only with `test: cmd`, such as `test: cmd ant test .` on a test
file. A named block of another kind, such as `data.json`, is only written.

### Failing examples

An example that does not compile, or that panics, says so where the reader looks: in a comment on
the line that fails, with the message `ant` prints. Never hide it in a directive.

```ant
let xs = [1, 2, 3]
xs[3] = 1 // panic: index 3 out of range for an array of length 3
```

````md
```ant
let xs = ["a", "b"]
let labeled = Array.map(xs, func(s: string, i: int): string { "${i}${s}" })
// error: cannot use func(string, int): string where func(string): U is expected
```
````

- `// error: <message>` marks a line `ant check` rejects. The block is checked, not run. Every error
  `ant check` reports must be marked, at its line, so follow-on errors are shown too. A block with
  `// error:` has no cases.
- `// panic: <message>` marks the line a program panics at. After a bare expression it makes that
  expression a case, and its block a block of cases, where each case runs on its own. In a program,
  as one that uses `Channel`, which cases cannot, put it on a statement instead:
  `let sent = Channel.send(done, true) // panic: …`.
- At the end of a line, the comment marks that line. When the line would grow past 100 columns, put
  the comment on its own line right below; it then marks the code line above it.
- Give the whole message. The test checks that the reported message contains the comment's text.

### Preparation

What the reader does not need to see goes in an HTML comment that starts with `test:`. It applies to
the next ` ```ant ` block. A misspelled prefix (`tests:`, `Test:`), an unknown directive, a second
directive of the same name on one block (`file` aside), and a directive that does not apply to its
block all fail the test.

| Directive                   | Meaning                                                               | Applies to                       |
| --------------------------- | --------------------------------------------------------------------- | -------------------------------- |
| `test: file <path>` + body  | Write the body to `<path>` in the page's directory.                   | any block                        |
| `test: before` + body       | Ant code put in front of the block, hidden from the reader.           | cases, an unnamed program        |
| `test: stdin` + body        | Feed the body to standard input.                                      | a block that runs                |
| `test: args <a> <b>`        | Pass arguments, which `Os.args()` returns.                            | an unnamed program without `cmd` |
| `test: cmd ant <args>`      | Run this `ant` command in the page's directory, such as `ant test .`. | a program                        |
| `test: exit <n>`            | Expect exit status `<n>` instead of 0.                                | a block that runs                |
| `test: check-only <reason>` | Check and format-check only. For output that changes between runs.    | an ant program, not with `cmd`   |

A block runs when it is an unnamed program, or a named one with `cmd`, has no `check-only`, and
marks no line with `// error:`. `exit` and `// panic:` exclude each other.

A directive with a body puts it on the lines after the first, up to `-->`; nothing else goes on the
first line of `stdin` or `before`:

````md
<!-- test: file people.csv
name,age
Ann,31
Bob,27
-->

```ant
let rows = Result.unwrap(Csv.parse(Result.unwrap(File.read("people.csv"))))
for row in rows {
  println(row[0])
}
```

```text
name
Ann
Bob
```
````

`check-only` is for output that depends on goroutine order, `Random`, the clock, the network, or a
process. Prefer an example that runs: a fixed seed, a fixed instant, a local listener.

### Rules for every block

- It is formatted as `ant fmt` formats it. Run `ant fmt` on it before pasting.
- It is complete: no `...`, no names it does not define (put setup in `test: before`).
- ` ```ant signature ` blocks belong to the generated part only, which holds nothing else.
- A block may be indented inside a list item, up to three spaces, as Markdown allows.
- Other languages (` ```sh `, ` ```text ` not after an ant block, ` ````md `) are not run.

## Testing

From the root of this repository:

```sh
# one page, or every page under a directory
ANT_DOCS_ONLY=lib/Array/map.md go test ./tests -run 'TestDocs$'
ANT_DOCS_ONLY=tour/ go test ./tests -run 'TestDocs$'

# every page, and the reference against the prelude
go test ./tests -run 'TestDocs'

# after changing prelude.d.ant: new skeletons, refreshed generated parts
ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference
```

`go test ./...` runs them too, so a change to `ant` that breaks an example fails CI. Format the
Markdown with oxfmt (print width 100, prose wrapped), as every Markdown file of the project is. From
the workspace that holds this repository and its `.oxfmtrc.json`:

```sh
npx oxfmt -c ../.oxfmtrc.json 'docs/**/*.md'
```
