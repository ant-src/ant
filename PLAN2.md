# Plan 2: a tested language

PLAN.md built the implementation in eight phases. This plan makes it trustworthy. The method is the
same as before: the specification (`spec/ant-lang.md`, `spec/stdlib.md`) is the only source of
truth, every test is derived from a sentence in it, and each phase ends in a commit with every test
green under the race detector.

The one new idea is **traceability**: a test is not only a program that passes, it is a program
that names the rule it proves, and a test enforces that no rule goes without one. Coverage is then
a fact the repository can state, not an impression.

## 1. Where we start

Measured on the current tree (commit 449e2fb).

| | Now |
| --- | --- |
| Programs under `tests/run`, `tests/unchecked`, `tests/check` | 63, 36, 9 |
| Largest test program | 100 lines |
| Spec code blocks parsed by the parser test | 118 of 118 |
| Library functions with no use in any test | not known |
| Checker asserted to reject every `tests/unchecked` program | no |
| Fuzz targets, benchmarks | 0, 0 |
| Go statement coverage, all tests, whole program | 83.8% |
| Weakest packages (approximate, per function) | `cmd/ant` 0%, `internal/repl` 0%, `ast` 25%, `internal/types` 56% |
| Continuous integration | none |

## 2. Definition of done

Ant is "tested" when all of these hold and a test or a CI job keeps each one true:

1. **Every rule-bearing heading of the spec has tests.** Each `###` heading in chapters 2 to 12,
   each chapter without subsections (3, 10), and each namespace of `stdlib.md` is named by at
   least one test program, and a test fails when one is not (section 3.1).
2. **Every spec example is honored.** Each `// compile error` comment in a spec code block has a
   check program that rejects the same construct, and each `// panic:` or `// <value>` comment on
   an expression is asserted by a run test (section 3.2).
3. **Every library function and every panic is exercised.** Each function and constant in
   `prelude.ant` appears in some test program; each of the 33 ◆ functions in chapter 11 and each
   of the 23 rows of the chapter 12 panics table has a program that triggers it (section 3.3).
4. **The checker is sound on everything we have.** It accepts every program under `tests/run`
   and `examples/`, rejects every program under `tests/unchecked`, and agrees with the runtime on
   generated expressions (sections 3.4, 3.6).
5. **No input crashes the interpreter.** Fuzzers for the lexer, parser, checker, interpreter and
   JSON decoder run in CI; an Ant panic is fine, a Go panic or a fatal error is a bug (section 3.5).
6. **Race-clean.** All tests run under `-race` in CI, and chapter 9 has stress programs that race
   on purpose on every container kind and every sync object.
7. **Go coverage at least 90% of statements overall and 80% in every package**, measured whole
   program, enforced by a gate (section 3.7).
8. **Real programs run.** At least eight programs under `examples/` of 100 to 300 lines each,
   run by the harness, linked from the README (section 3.8).
9. **CI is green on every push** on macOS and Linux: gofmt, vet, race tests, coverage gate, fuzz
   smoke, and a build of the command.
10. **Spec feedback is written down**, one file, each item pointing at a spec line, so the spec
    and the implementation never silently disagree (section 3.10).

## 3. Mechanisms

### 3.1 Tags and the spec coverage test

Every test program starts with a tag line naming the heading or headings it proves:

```
// spec: 4 Unions; 4 Tagged unions
// spec: 10
// spec: stdlib Json
```

The form is `<chapter number> <subsection heading>`, the chapter number alone for a chapter
without subsections, or `stdlib <namespace>`. The heading text is copied from the spec, so a
renamed heading fails the test and the tags get updated with it. The existing free-form first-line
comments (`// Section 4: unions need ...`) become tags plus a sentence saying which rule.

`tests/spec_test.go` parses the headings of both spec files, scans every `.ant` file under
`tests/` and `examples/`, and fails when a tag names a heading that does not exist, when a heading
outside the `pending` list has no tag, or when a heading in `pending` has one (a stale entry). The
`pending` list is a literal in the test file; it starts with every heading the current corpus
does not cover and the audit phases empty it. Under `-v` the test prints the matrix, one line per
heading with its file count, which is the coverage report.

Chapter 1 (Principles) is exempt. Chapter 13 is the one program block in the spec and is already
`tests/run/example.ant`.

The harness already requires every `tests/run` program to pass `ant check`; that stays. Programs
under `tests/unchecked` are tagged like any other, usually with `12 Runtime panics` and
`4 Checking and running`, since they prove that `ant run` runs what `ant check` rejects.

### 3.2 Spec comments as oracles

The spec's code blocks carry three kinds of comment the corpus must honor, and the audit
transcribes each by hand into a tagged program:

- `// compile error: ...` (11 in code blocks, more in prose). Each becomes a `tests/check`
  program with an `// ERR:` marker on the same construct. The marker text is the checker's own
  wording; the spec's comment explains the error and does not dictate the message.
- `// panic: ...`. Each becomes a run or unchecked program with an `// err:` trailer line.
- `// <literal>` after an expression, such as `// 3.5`, `// true`, `// "u-1"`. Each becomes an
  assertion in a run test.

No test matches these comments automatically, since the blocks are fragments; the tag on each
program ties it to the heading, and the coverage test of section 3.1 is what keeps the
transcription complete.

Making whole spec blocks executable was considered and rejected: most blocks are fragments with
`...` placeholders, and a wrapper that guesses the missing context would be a second, hidden
corpus. Transcription with tags is explicit and the coverage test keeps it complete.

### 3.3 Library and panic coverage tests

`TestLibraryCoverage` parses `prelude.ant` through `prelude.Decls()`, and for every function and
constant `Ns.name` requires the token to occur in some test program. `TestPanicCoverage` parses
chapter 11 for lines marked ◆ (the function name is the first token) and requires a program tagged
`// panics: Array.slice` for each; for the chapter 12 table it holds a hand-written key per row
(`arith-kind`, `bool-kind`, `int-overflow`, `div-zero`, `shift`, `conversion`, `string-bool`,
`field`, `index`, `index-range`, `call`, `record-literal`, `array-literal`, `is`, `for`,
`for-yield`, `destructuring`, `destructuring-short`, `variable`, `underscore`, `library-kind`,
`library-value`, `var-init`, `panic`, `depth`, `try`, `never`) and requires a `// panics: <key>`
program for each. The test also compares its key count with the table's row count, so an edited
table fails the test until the keys follow.

A ◆ panic in a checked program belongs under `tests/run` with `// exit: 1` and an `// err:` line.
Only programs the checker rejects go under `tests/unchecked`. Both pending lists start full and
shrink through the audit.

### 3.4 Soundness on the corpus

Two assertions that tie the halves together:

- `TestUncheckedRejected`: `Check` reports at least one diagnostic for every program under
  `tests/unchecked`. Missing today.
- `TestCheckExact` (strengthening the existing marker test): a `tests/check` program yields exactly
  the marked diagnostics, on the marked lines, and no others. Cascades are bugs.

### 3.5 Fuzzers

Go native fuzzing, seeded from the corpus, one target per layer. Every target has one rule: no Go
panic, no fatal error, no hang.

| Target | Package | Property |
| --- | --- | --- |
| `FuzzLex` | `internal/lexer` | terminates; positions never go backwards; errors are values |
| `FuzzParse` | `parser` | no Go panic in Program, Session and Spec modes; `Incomplete` only with a clean prefix |
| `FuzzCheck` | `internal/check` | for any input that parses and resolves, the checker returns diagnostics or nothing, never panics |
| `FuzzRun` | root | for any input that checks clean, `Run` with a 50 ms deadline and no I/O namespaces returns normally, an `*PanicError`, or the context error |
| `FuzzJSON` | `internal/lib` | `Json.parse` never panics; parse, stringify, parse again gives an equal value |
| `FuzzStrings` | `internal/lib` | `String` functions on arbitrary bytes respect UTF-8 boundaries and never index outside the string |

`FuzzRun` skips inputs that mention `go `, the sync types, or `Time`, because a goroutine blocked
in a host call outlives the deadline and would leak across iterations; chapter 9 gets
deterministic stress tests instead. Two fixes are expected to come out of this phase before the
fuzzers can run long: a nesting depth limit in the parser (recursive descent on a Go stack cannot
recover from overflow) and the same limit in the checker's and evaluator's recursion on nested
expressions. Cyclic values are a third hazard to test on purpose rather than hope the fuzzer
finds: `==`, `String.from` and map hashing are specified to terminate on a cycle, and
`Json.stringify` panics with `cyclic value` (section 6); each needs a test.

CI runs each fuzzer for 20 seconds; long runs are local. Crashers Go writes under `testdata/fuzz`
are committed as regressions.

### 3.6 Generated expressions: checker versus runtime

`TestGenerated` in `tests/` builds random typed expressions from a small grammar and checks that
the two halves agree. The grammar: literals of `int`, `float64`, `string`, `bool`, `null`;
arithmetic and comparison on same-typed numbers; `&&`, `||`, `!`; `??` and `?.` on nullable
operands; array literals and indexing; record literals, fields, and `is` on declared tagged
unions; calls of small declared functions and lambdas. Each case yields `(expr, T)` and asserts:

1. `var x: T = <expr>` checks clean;
2. evaluating `<expr>` through a `Session` yields a value whose kind matches `T`, or an Ant panic
   whose row in the chapter 12 table is marked ◆;
3. after one random mutation that swaps a literal for one of another kind, either the checker
   reports an error or the run does not raise a kind panic.

The third assertion is the soundness probe. The seed is printed and settable through an
environment variable; `-short` runs a thousand cases, the full run a hundred thousand. A failure
prints the program so it can be copied into the corpus as a regression test.

### 3.7 Coverage gate and CI

`scripts/coverage.sh` runs the whole test suite with `-coverpkg=./... -coverprofile`, computes
statement-weighted coverage overall and per package from the profile, and exits non-zero below the
thresholds it is given. The gate starts at today's 83% overall with no per-package floor, and
each phase that raises coverage raises the gate in the same commit; the end state is 90% overall
and 80% per package. `cmd/ant` reaches its floor by moving everything but `func main` into a
function the test calls with piped standard input; `internal/repl` the same way.

`.github/workflows/ci.yml` on push and pull request, matrix of ubuntu and macOS, Go from `go.mod`
and latest: `gofmt -l` must print nothing, `go vet ./...`, `go build ./cmd/ant`,
`go test -race ./...`, the coverage gate, and one short job per fuzzer. The public spec
repository is checked out beside the module (`ANT_SPEC` points the parser test at it) so the
spec tests run in CI too.

### 3.8 Examples: real programs

Programs of 100 to 300 lines under `examples/`, each with the harness trailer and a tag line,
written to make features meet. Each is also documentation and the README links them.

| Program | What it exercises |
| --- | --- |
| `config.ant` | the decoder `stdlib.md` describes: `Json.parse`, narrowing a `Json.Value` tree with `is`, a function per record type, `try` with a path in the error (`$.servers[2].port`), `Math.toInt` |
| `wordfreq.ant` | files from `Os.args`, `String` splitting, a `Map` of counts, N goroutines with `WaitGroup` and `Mutex`, `Array.sort` with a comparator, top ten |
| `calc.ant` | tokenizer, recursive-descent parser and evaluator for arithmetic with variables; a tagged union for the tree, `Result` for errors, functions on types |
| `queue.ant` | producer and consumer over a `Cond`-guarded queue, `Atomic` counters, `defer` in goroutines, a goroutine that panics and the program's exit status |
| `nbody.ant` | float arithmetic in nested loops over records and arrays, `Math`; doubles as a benchmark |
| `todo.ant` | a command with subcommands over a JSON file: `File`, `Json.stringify`, `Os.args`, `Os.exit` statuses |
| `fetch.ant` | `Http.get` and `Http.post` against the harness echo server, header maps, JSON bodies, error paths that must be `Err` and never panics |
| `bf.ant` | an interpreter for a small stack language: `Array[byte]`, `Bits`, code points, `String.fromBytes` |
| `bank.ant` | chapter 9's sharing patterns: `RwMutex`, `Once`, `Atomic.compareAndSwap`, a thousand goroutines; runs under the race detector |
| `shapes/` | a multi-module program with a nested module directory and `export var`, imported both ways |

Bugs these find go to the corpus as single-rule tests before the fix, so each bug leaves a test
behind.

### 3.9 Benchmarks

`bench/*.ant` with `BenchmarkRun/<name>` in `bench/bench_test.go` running each through the
library: recursion (`fib 27`), a ten-million-iteration loop, array push and index, map set and get
with string keys, string building, record allocation and field access, closure calls, a sequence
function driven by `for`, goroutine fan-out with `WaitGroup` and `Mutex`, and JSON parse and
stringify of a megabyte. `bench/BASELINE.md` records the numbers with machine, Go version and
date. Comparisons use `benchstat` installed as a tool, not a dependency. No optimization lands
without a benchmark showing it, and the first candidates are the ones PLAN.md lists: the cell
allocated per shared store, and frame allocation per call.

### 3.10 Spec feedback

`SPEC-FEEDBACK.md` in this repository: one entry per disagreement or gap, with the spec line, what
the implementation does today, and the suggested wording. The audit appends to it; you resolve
entries in the spec; a resolved entry is deleted and the tests follow the spec. The initial
entries, all found so far:

- Chapter 9, Visibility: the text gives the necessary condition (a write is visible "only when
  something orders them") but never the sufficient one; nothing promises that a write ordered
  before a read is seen. State both halves of the Go rule.
- Chapter 9, Data races: "the ordinary index or key error" for a racy map operation, but a missing
  key yields `null` everywhere else; say "or finds the key absent".
- Chapter 9, Data races: "before or after the concurrent mutation" should be per mutation, since
  a multi-element operation can see one racing store and not another. The pricing sentence should
  say one lock per map operation, shared for reads, and that push and pop exclude each other.
- Chapter 9, the race definition names "one variable, field, or element"; the length of an array
  and the key set of a map are shared state too.
- Chapter 9: the first bullet is what keeps the checker's types true under races, and the
  pricing paragraph is an implementation note; both are worth saying as such.
- `stdlib.md`, `Os.exit`: whether deferred blocks run. The implementation does not run them,
  reading `never` as an exit without unwinding.
- Chapter 4, Function types: `func(): int | Meters` parses greedily as a union return type; a
  reader may expect the union of a function and `Meters`. A sentence settling it.
- Chapter 7, Narrowing: assignment narrows a local to the assigned type, so `v = "plain"`
  followed by `v is UserId` is an always-false error. Intended, but worth an example since it
  surprises.
- Chapter 5, Loops: a map literal in a `for` header needs parentheses, as the block brace is
  ambiguous. Worth stating.
- Chapter 2, Literals: `1.x` is a lexer error rather than `1` then `.x`. Worth stating.
- Chapter 12, Values: what `String.from`, map hashing, and `Json.stringify` do on a cyclic value.
- Chapter 11, `String.from`: rendering of floats of very large or small magnitude is unspecified;
  the implementation prints plain decimal.

### 3.11 Diagnostics

The spec's own `// compile error:` comments set the style: short, concrete, in the spec's
vocabulary ("a Person is not a Named", "T would be Number, a union; narrow a and b first").
The messages are not matched against the spec verbatim, but the diagnostics phase brings them
into that voice, fixes the flattened union names (keep the declared alias beside the
normalized form and print the alias), confirms no cascades on every check program, and gives the
panic report a tested format: `panic: <message>` then one `at` line per frame, innermost first.

## 4. The audit, chapter by chapter

The audit is the bulk of the work: read a chapter, list its normative sentences, write one program
per sentence or tight group, tag it, run it, and when it fails decide whether the bug is in the
implementation, the test, or the spec. Each chapter below lists what it contains so the phases can
be sized; the list is a reminder of scope, not a substitute for reading the chapter.

- **2 Lexical structure.** Comments; identifiers and keywords (reserved words as names are errors);
  every literal form (integer forms and overflow, floats, strings and escapes, templates with
  interpolation, bytes); the line-ending rule and which tokens continue a line. Lexer table tests
  in Go plus check programs for each error named.
- **3 Variables.** Declaration and reassignment, `_` as a sink and a read error, `void` read as a
  compile error, shadowing, top-level `var` order and the ◆ read-before-initialized panic.
- **4 Types.** The largest chapter and the checker's core: the erasure model; the ten numeric
  kinds, defaults and conversions; records nominal by tag, no dynamic access, spread, `==` and
  `Record.same`; named primitive types and their constructors, unwrapping, the tag-mismatch error
  and kind panic; arrays and maps; unions, the discriminator rule and the join error; tagged
  unions; function types; generics and the solving order (expected type, then non-lambda
  arguments, then lambdas), `T would be Number`, `int and float64 are different T`; type
  declarations and alias cycles; the assignability table, row by row, each with an accepting and a
  rejecting program; inference of returns, lambdas, and empty literals.
- **5 Expressions.** Operator table with precedence and kinds; string `+` and comparison; pipes;
  blocks; `if`; `is` with narrowing, exhaustive chains, always-true and always-false errors; loops
  over arrays, maps, strings and sequence functions, `while`, `break` and `continue`;
  destructuring of records and arrays, fixed and rest positions; `return`.
- **6 Functions.** Positional and named arguments, defaults, arity errors and panics; anonymous
  functions; closures capturing by reference; functions on types.
- **7 Nullability.** `T | null`, `??`, `?.`; narrowing on locals only, reset by assignment, never
  through a field path, never for a local assigned in a nested body.
- **8 Errors.** `Result` and its helpers; `try` and the error-type compatibility rule; `panic`;
  `defer` order, running on panic and on return, in goroutines.
- **9 Concurrency.** `go`; each sync object's documented semantics and ◆ panics; data races as
  memory-safe stress programs on every container and the sync objects under `-race`; visibility
  through each happens-before edge; the sharing patterns; a goroutine panic's report, exit status
  and deferred blocks.
- **10 Modules.** Import forms and paths, `export`, the module as a record with live `export var`,
  cycles as errors, initialization order, `main` requirements, the three exit statuses of
  `ant run` and `ant check`.
- **11 Standard library.** Every function in the prelude with one program per namespace at
  minimum and one per ◆ case; the `String.from` rendering rules on every kind; `Iter` sequences
  including early exit; `Math` and `Bits` edges (`minInt`, `NaN`, infinities, shifts at the
  width); `Result` and `Record`; the sync namespace.
- **12 Dynamic semantics.** Values and identity; names and scopes; evaluation order of every
  expression form, observed through side effects; deferred blocks and panics; goroutines; the
  panics table, every row.
- **stdlib.md.** `File` with every operation and its error cases (missing file, `readLine` to
  end, `pipe`, `close` twice); `Http` against the echo server and failure paths; `Time`
  (`sleep` blocks only its goroutine, `monotonic` never decreases); `Os` (`args`, `env` null,
  `exit` status); `Json` (every value form, numbers as `float64`, `Math.toInt` null cases,
  records stringified without tags, round trips, decode errors as `Err`).

## 5. Phases

Each phase is one commit, or a few, with the suite green under `-race` and the coverage gate at
least where it was.

- **Phase 9: harness.** Tag convention; `TestSpecCoverage`, `TestLibraryCoverage`,
  `TestPanicCoverage` with their pending lists filled from the current corpus; `TestUncheckedRejected`; `TestCheckExact`; `tests/README.md` describing corpora,
  trailers and tags; `SPEC-FEEDBACK.md` with the initial entries. Done when the current corpus is
  tagged and every new test passes with honest pending lists.
- **Phase 10: fuzzers and CI.** The six fuzzers with seeds; the nesting limits; the cyclic-value
  decision and tests; `scripts/coverage.sh` gating at 83%; the workflow. Done when each fuzzer
  has run ten minutes locally without a finding and CI is green on both platforms.
- **Phase 11: examples.** The ten programs of section 3.8, each a run test. Bugs found become
  corpus tests. Done when all ten run under the harness and the race detector and the README
  links them.
- **Phase 12: audit of chapters 2, 3, 5, 6.** Lexical, variables, expressions, functions.
- **Phase 13: audit of chapters 4 and 7.** Types and nullability: the checker's corpus grows
  from nine files to one per subsection at least.
- **Phase 14: audit of chapters 8, 9, 12.** Errors, concurrency with the stress programs,
  dynamic semantics with the panics table complete.
- **Phase 15: audit of chapters 10, 11 and stdlib.md.** Modules, the library with every
  function used, the I/O namespaces. Done when all four pending lists are empty.
- **Phase 16: generated expressions.** The differential test of section 3.6.
- **Phase 17: unit coverage.** `ast` through golden S-expression files for the whole corpus;
  `internal/types` tables for assignability, union construction and printing; `cmd/ant` and
  `internal/repl` through piped input; `token` and `prelude` gaps. Gate raised to 90% overall and
  80% per package.
- **Phase 18: diagnostics.** Section 3.11.
- **Phase 19: benchmarks.** Section 3.9, the baseline and at most the two listed optimizations.
- **Phase 20: documentation and release.** `doc.go` with the mapping between Go and Ant values
  that `Func.Call` and host functions rely on; README sections for examples, testing and spec
  conformance; `CHANGELOG.md`; a `v0.1.0` tag. Pushing and tagging are yours.

Phases 12 to 15 can be reordered or split; the order above puts the checker's chapter early
because the checker is the newest code with the thinnest tests.

## 6. Decisions

Taken:

- The spec's `// compile error:` comments explain errors; they do not dictate the checker's
  wording. No verbatim matching.
- The REPL stays as it is. No line editing, no dependency.
- Coverage thresholds are 90% overall and 80% per package, enforced by the gate.
- The repository is `github.com/ant-src/ant`, public, default branch `master`, like the spec.
- **Cyclic values**: the spec already settles `String.from` (a value being rendered renders as
  `...`, line 1478) and map hashing (a bounded prefix, line 1601). `Json.stringify`, where `...`
  is not JSON, panics with the message `cyclic value`; `SPEC-FEEDBACK.md` records it.

## 7. Status

Phases 9 to 17 are done and committed: the harness with tags and coverage tests, the fuzzers
and CI, the ten examples, the audit of every chapter (all three pending lists are empty), the
generated-expression test (300 000 cases across three seeds without a disagreement), and the
coverage gate at 90% overall and 80% per package, which the tree passes. Phases 18 to 20
(diagnostics, benchmarks, documentation and release) remain.
