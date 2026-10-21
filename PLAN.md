# Ant in Go: implementation plan

A tree-walking interpreter for Ant, written in Go against `../spec/ant-lang.md` (draft 0.1) and
`../spec/stdlib.md`. Priorities, in order: `ant run`, then `ant check`, then the REPL that a bare
`ant` starts. No third-party dependencies. The repository is a Go module that other Go programs
import; the command is a thin layer over the library.

The spec is the only source of truth. Its ```` ```ant ```` blocks, its section 12 panics table,
its section 11 rendering rules, and its section 13 example seed the tests. Nothing is taken from
the spec repo's history.

## 1. Decisions

These are the choices that shape everything else. Each one is the simplest option that meets
the spec; the upgrade path is noted where one exists.

**Values are Go interface values.** `type Value = any`, holding Go's own `int8` … `uint64`,
`float32`, `float64`, `bool`, `string`, two sentinel types for `null` and `void`, and pointers for
arrays, maps, records, tagged primitives, functions, tags, and sync objects. The ten numeric kinds
map one to one onto Go's types, so the kind of a value is its Go dynamic type and a type switch is
the kind test. No custom tagged-union struct. A Go interface value is two words, which matters
only where two goroutines can reach one slot; those slots box each stored value, as the next
decision describes.

**Real goroutines, no interpreter lock.** `go { }` starts a Go goroutine and nothing serializes
interpreter code, so Ant goroutines run in parallel and the Go scheduler preempts them. The
memory-safety guarantee of section 9 comes from the design the spec's own implementation note
prices: every slot that two goroutines can reach holds one machine word. A shared slot is an
atomic pointer to an immutable cell holding the value; a store allocates a cell and publishes the
pointer, a read loads the pointer and copies the value out. The Go memory model guarantees that a
racy word-sized read observes some written value, while an unprotected interface, string, or slice
can tear into arbitrary memory corruption, and a racy Go map is an unrecoverable fatal error.
Shared slots are record fields, array elements, map entries, module variables, and locals that a
nested body captures; every other local is private to its goroutine and stays plain. Arrays
publish an immutable header behind an atomic pointer and take a small mutex only for push and
pop; maps take a read-write mutex per map. Microbenchmarks on this machine put a shared store at
about 6 ns over a plain one, a locked map read under 1 ns over a bare one, and an array read
through its header under 0.2 ns over a bare one, while eight busy goroutines ran 3 to 11 times
faster than under a global lock. All of it lives inside the `value` package behind unexported
fields and accessor methods, so the compiler enforces the discipline everywhere else, and
`go test -race` over the concurrent conformance programs reports only interpreter bugs, since
atomic operations are invisible to the race detector.

**Names resolve to slots at load time.** The parser already tracks scopes to enforce the
redeclaration, shadowed-import, and prelude-shadowing rules, so it also assigns each `var`,
parameter, and loop variable a slot index in its block and annotates every identifier with
(depth, slot), module field, or prelude value, and whether the slot is plain or shared: a local
that a lambda, anonymous function, `go` block, or `defer` block refers to is shared, every other
local is plain. Environments hold both kinds with a parent pointer.
An unknown name is annotated as unbound and panics when evaluated, as section 4 requires. Blocks
that declare nothing share their parent's environment.

**Module scope is a record.** Each module's top-level names are the fields of one record value,
in declaration order, so `Geometry.counter` is an ordinary field read and `import { counter }`
binds the local name to (module record, field index), which makes imports live. Module fields are
shared slots. A field holds an `uninitialized` sentinel until its initializer runs; reading it
panics (section 10).

**Control flow returns, panics unwind.** `eval` returns `(Value, control)` where `control` is
none, return, break, or continue, with the carried value. Every statement list and every operand
evaluation checks it, which is verbose but idiomatic Go and cheap. Ant panics are Go panics
carrying an `*AntPanic`, recovered only at blocks that registered a `defer`, at `for` loops over
sequences, and at goroutine roots. The parser marks which blocks contain a `defer`, so the Go
`defer`/`recover` cost is paid only there. A `return`, `break`, or `continue` inside a `for` body
driven by a sequence function must leave through the user's sequence function; that is the one
place control travels as a Go panic (a private unwind value), caught by the `for`.

**The prelude is written in Ant.** `internal/prelude/prelude.ant` declares every prelude type and
every library signature in Ant notation (`func Array.map[T, U](xs: Array[T], f: func(T): U):
Array[U]` with no body), parsed by the same parser in a declarations-only mode. The checker reads
its types from it; the runtime binds Go implementations to the same names and a test asserts the
two tables match exactly. Parameter names come from the signatures, so named arguments work on
library functions as the spec says they do for any declared function.

**Diagnostics** are `[]Diagnostic` values with path, line, column, and message; the command
prints them as `path:line:col: message` on standard error. Exit status is 0, 1 (panic in main or
check errors), or 2 (could not load: syntax error, import cycle, missing module, missing `main`),
per section 10, and only the command ever exits.

**A library first, a command second.** The module is `github.com/ant-src/ant`, to match the spec
repository, and its root package is the public API; `cmd/ant` is a thin command over it,
installed with `go install github.com/ant-src/ant/cmd/ant@latest`. The library never touches
process state: it reads sources through an `fs.FS` or from strings, writes to the `io.Writer`s it
is given, takes its arguments and environment from options, returns exit status and panics as
values, and never calls `os.Exit` or sets process-wide limits. A loaded `Program` is immutable
and can be run many times, concurrently; each `Run` is an instance with its own module state.
Host programs extend Ant the way the prelude does: a namespace is a set of Go functions plus
their signatures in Ant notation, so `ant check` types host calls like library calls. Ant
functions are callable from Go, and a `Session` evaluates one input at a time on a persistent
instance, which is what the REPL is built on. The conformance runner drives this API in process,
so every program in the suite also tests the API.

## 2. Layout

```
ant/
  go.mod                      module github.com/ant-src/ant
  ant.go                      public API: Load, Program, Run, Check, Func, Session, options
  errors.go                   LoadError, Diagnostic, PanicError, ExitError
  example_test.go             runnable examples: embed a script, call a function, add a namespace
  cmd/ant/main.go             ant run | ant check | ant   (bare: REPL)
  ast/        nodes with positions; resolver annotations                  (public, for tooling)
  parser/     recursive descent, one token of lookahead; scope rules; slot assignment (public)
  value/      Value kinds, tags, records, ordered map, equality, hashing, String.from  (public)
  internal/
    token/      token kinds, positions
    lexer/      newline rule, template parts, literal validation
    loader/     source provider, import graph, cycle check, depth-first module order
    interp/     evaluator, environments, calls, defer and panic, goroutines, cancellation
    lib/        builtins: core library (section 11) and I/O namespaces (stdlib.md)
    prelude/    prelude.ant and the loader that binds names to lib implementations
    types/      the checker's type language
    check/      the checker
    repl/       the read-eval-print loop over a Session
  tests/
    run/        programs written from the spec, one per topic, with lib/ helper modules for the
                import tests; each ends in a comment trailer of expected stdout lines, exit
                status, stderr substrings, and command-line arguments
    check/      checker programs: clean ones, and ones annotated with the compile errors the
                spec names
    conformance_test.go   drives Load and Run in process with injected writers and arguments
                and compares stdout, exit status, stderr; a few cases also run the built
                command to check real exit codes
```

What is public is the surface a tool or a host needs: the AST and parser for editors and
formatters, the values for passing data in and out, and the root package for everything else.
The internal packages can be merged later if the boundaries prove heavier than they are worth;
the split exists so the checker and the interpreter share the parser and nothing else.

## 3. Lexer and parser

The grammar is sections 2, 5, 6, and 10 of the spec: the layout rules, the operator table, the
brace rule, the function forms, and the module form. The parser is written from those sections.
There is no separate grammar file and none is needed, since the spec promises one token of
lookahead and no feedback from the checker.

Lexer:

- Emits NEWLINE unless the previous token cannot end an expression (binary operator, `|`, `??`,
  `,`, `:`, `(`, `[`, `{`, `=`, prefix operator, or a keyword other than `return`, `break`,
  `continue`, `true`, `false`, `null`, `void`) or the next non-blank, non-comment line starts
  with `.`, `?.`, `|`, or `else`, as section 2 says.
- A word after `.` or `?.` is always a name, never a keyword. Reserved words are their own token
  kind so the parser can say "reserved word" wherever one appears outside a field position.
- Integer literals that overflow `int64` are a syntax error. Floats need digits on both sides of
  the point. `_` only between digits.
- Escapes are validated here, including `\u{...}` range and surrogate checks. A template string
  token carries its literal pieces and the source and position of each `${...}`; the parser
  sub-parses those with the enclosing boundary and loop context, so `${x ?? return y}` works.

Parser state: a newline-suppression depth for `( [ {` lists, a `header`
flag under which `Path {` does not start a literal, the innermost boundary (function, lambda,
`go`, `defer`) with its loop depth, and a scope stack. Rules it enforces as syntax errors:

- `import` only before declarations; `type`, named `func`, and `export` only at the top level.
- `return` outside a boundary; `return` with a value directly inside `go` or `defer`; `break` or
  `continue` with no loop in the current boundary.
- Redeclaration in one scope; two imports binding one name; a top-level declaration shadowing an
  import; any declaration shadowing a prelude name (fixed list in `internal/prelude`).
- `??` mixed with `&&` or `||` without parentheses; chained assignment; a pipe right side that
  is not a call; positional after named argument; spread in a map literal; a keyword as a
  shorthand key or destructuring binding.
- An array literal left of `=` is reinterpreted as a destructuring target; `=` only.

Brace disambiguation is the one place the grammar needs care: after an identifier path a `{` is
a literal (map when the path is `Map`), except in a header; after `func(...)`, `else`, `go`,
`defer`, or a header it is a block; anywhere else it is a lambda. A statement starting with `{`
is therefore always a lambda.

The parser also produces the resolver annotations described in section 1, and marks blocks that
contain a `defer` and blocks that declare nothing.

Tests, all derived from the spec: every ```` ```ant ```` block in `ant-lang.md` parses in a spec
mode that allows top-level statements and `...` placeholders, which keeps the parser honest
against every example the spec gives; a table of syntax-error snippets, one per sentence in the
spec that names a syntax error (`.5` and `1.`, an integer literal that does not fit, `break`
without a loop, `return` with a value in a `go` block, redeclaration in one scope, two imports
binding one name, a declaration shadowing a prelude name, `??` mixed with `&&`, a chained
assignment, a pipe whose right side is not a call, a keyword as a shorthand key, and the rest),
each asserting a message substring; and a table of AST assertions for the precedence and
desugaring rules of section 5 (pipe to call, `try` before pipe, `is` as a comparison, greedy
function return types, leading `|` and `else` continuation lines, `return` at end of line,
shorthand and keyword keys, nested destructuring, named arguments).

## 4. Values

```go
type Value = any
// int8 int16 int32 int64 uint8 uint16 uint32 uint64 float32 float64 bool string
// Null, Void              singleton structs
// slot    atomic.Pointer[cell]; cell { v Value } is immutable once published
// *Array  atomic pointer to an immutable header { elems []slot; len int }; mutex for push and pop
// *Map    ordered hash map keyed by ==, with a read-write mutex
// *Record { tag *Tag; fields []slot }           also module records
// *Tagged { tag *Tag; payload Value }           named primitive values
// *Func   declared, anonymous, lambda, builtin, or constructor
// *Tag    record tag (field names in order), alias tag, or sync-type tag; funcs declared on it
// *Mutex *RwMutex *WaitGroup *Once *Cond *Atomic
```

- **Tag.** `name`, `kind` (record, alias, primitive-constructor, sync), `fields []string` with a
  name-to-index map, and `funcs` for the functions declared on the type. A constructor is a
  `*Func` that points at its tag and base kind; `print(UserId)` renders the name.
- **Shared slots and arrays.** A `slot` is an `atomic.Pointer` to an immutable `cell { v Value }`:
  a store allocates a cell and publishes it, a read loads and copies. An array holds an atomic
  pointer to an immutable header `{ elems []slot; len int }`. A read or an index store loads the
  header, bounds-checks against its length, and touches one slot. Push and pop take the array's
  mutex, reuse spare capacity or copy into a larger slot array, and publish a new header. A
  builtin that walks an array works from one header snapshot, so a concurrent push or pop leaves
  it the state before or after, which is what section 9 promises.
- **Ordered map.** An entries slice (`key`, `value`, `dead` flag) plus an index from hash to
  entry positions. Removal marks the entry dead; the slice is compacted when dead entries pass
  half. `for` iterates by position, skipping dead entries, so entries appended during the loop
  are visited and removed ones are not, exactly as section 12 describes for arrays. A removed and
  reinserted key is a new entry at the end. Reads hold the map's read-write mutex shared and
  mutations hold it exclusively; `for` reads one entry under the lock and runs the body outside
  it, so the body may mutate the map.
- **Equality.** `==` is a type switch; records by tag identity then fields; arrays by length and
  elements; maps by key set and values; functions, tags, and sync objects by pointer; floats by
  IEEE rules. A set of pairs under comparison makes cycles terminate.
- **Hashing** agrees with `==`: kinds hash distinctly, `-0.0` hashes as `0.0`, and the hash of a
  container walks a bounded prefix (depth 3, first 8 elements) so cyclic values terminate.
- **Rendering** (`String.from`) per section 11: shortest float repr (`strconv.FormatFloat` with
  precision -1, `.0` appended when integral; `NaN`, `Infinity`, `-Infinity`), strings quoted
  with escapes inside containers, `Map{ "a": 1 }`, `Point{ x: 1.0 }` in declared field order,
  `UserId("u-1")`, `func`, a tag by name, a sync object by its type name, `...` for a value
  already being rendered. Open point: whether very large floats render in exponent form; the
  spec does not say, so start with plain decimal and record the choice.
- **Numerics.** Arithmetic dispatches on the pair of kinds; mismatched kinds or tags panic.
  Overflow checks for every integer width (compute in the wide type and range-check, or
  `math/bits` for 64-bit), unsigned subtraction below zero panics, integer `/` and `%` truncate
  toward zero and panic on zero, floats follow IEEE 754. Conversions follow the section 12 table,
  including the float-to-integer exactness and range checks and unwrapping a tagged argument.
- **Strings** are Go strings; lengths, indices, slices, and iteration count code points via
  `utf8` and `[]rune` conversions where an index is needed.

## 5. Interpreter

**Loading.** `ant.Load` reads the root file through the source provider, an `fs.FS` or the OS by
default, canonicalizes its path, parses it, then resolves imports depth first in source order,
parsing each module once by canonical path. A syntax error, a missing file, a cycle (detected
with an in-progress set), or a missing `main` returns a `*LoadError` with a position, which the
command reports and maps to exit status 2. The result is an immutable `Program`. `Program.Run`
builds each module record with tags, functions (including functions on types, attached as
fields of their tag or constructor), and `uninitialized` var fields; then modules initialize in
dependency order, each once, running top-level `var` initializers in source order; then `main`
is called with no arguments. `Run` returns nil when `main` returns, a `*PanicError` carrying the
message and the Ant stack trace when the main goroutine panics, and an `*ExitError` for
`Os.exit`; the command maps these to exit status 0, 1, and the given status. `Os.args()` returns
the arguments given in the run options; the command passes those after the file path.

**Environments.** `Env { plain []Value; shared []slot; parent *Env }`, the resolver having decided
at parse time which kind each local uses. A function call allocates one `Env` for
its parameters and body; a block that declares something allocates one; each `for` iteration
allocates one for the loop variable so closures capture the iteration's value. Slots start as
`uninitialized`, which is what makes a local function's self-reference work: the `var` slot exists
while its initializer runs, a closure created there sees it once it is bound, and a direct read
before then panics.

**Calls.** The callee evaluates first. A constructor takes exactly one positional argument of its
base kind. A builtin receives the running instance, `[]Value`, and its signature's parameter names
for named arguments, so one builtin value serves every run. A declared or anonymous function fills
parameters positionally, then by name, then defaults evaluate in the callee's environment in
parameter order. A lambda binds `x` to the one argument or to `void`. A per-goroutine depth counter
enforces a call limit (10 000 frames) with a clean panic; the limit is chosen so that the Go frames
behind 10 000 Ant calls stay well inside Go's default maximum stack, since the library never calls
`debug.SetMaxStack`. A per-goroutine stack of (function name, call position) feeds the panic trace.

**Expressions** follow section 12 one bullet at a time: record literal (head must be a record
tag; spread; every listed field exactly once; result in tag order), map literal (bare word keys
are strings, decided by the parser), arrays with spread, field and index reads, `?.` skipping the
rest of the postfix chain, `is` (kind names resolved by the parser to kind tests; any other path
evaluated and required to be a record tag or constructor), `&&`/`||`/`??` short-circuit, pipe as
call with the left operand first, assignment with target sub-expressions before the right side
and array-destructuring stores left to right, compound assignment, `try` (Ok yields the value;
Err returns a fresh Err from the enclosing boundary), templates via `String.from`.

**Blocks, defer, panic.** A block runs its statements and yields the last value, or `void` after
a declaration or assignment. A block the parser marked as containing `defer` keeps a list of
deferred closures and installs one Go `defer` that runs them in reverse, both on normal exit and
during an Ant panic; a panic inside a deferred block replaces the one in progress. `return`,
`break`, and `continue` leave through the same path, so every exited block runs its defers.
When a panic reaches a goroutine root the runtime prints `panic: message` and the trace to stderr;
in the main goroutine `Run` then returns a `*PanicError` and the command exits 1, elsewhere
the goroutine just ends.

**Loops.** `while` and `for` over arrays (index below current length each step), maps (position
scan, as above), strings (code points), and sequence functions. For a sequence, the loop builds a
`yield` builtin that runs the body and returns `true`, or `false` after `break`; calling it after
`false` or after the sequence returned panics. A `return` or an outer `continue`/`break` from the
body unwinds through the sequence function as a Go panic that the `for` recovers, after which the
loop reports the control to its own enclosing block.

**Goroutines and sync.** `go { }` spawns a Go goroutine that runs the block with its own call stack
and depth counter and reports any panic; the Go scheduler preempts and parallelizes freely. `Mutex`,
`RwMutex`, `WaitGroup`, `Once`, and `Cond` are small state machines, each guarded by its own Go
mutex and condition variable, so that `Mutex.unlock` when not locked, `WaitGroup` going negative,
and `Cond.wait` without the lock can panic as section 11 lists, and so that a `Mutex` may be
unlocked by a goroutine other than the one that locked it. `Atomic` is a shared slot, which gives
the total order section 9 asks for, since Go's atomic operations are sequentially consistent.
`compareAndSwap` and `add` are lock-free loops: load the cell pointer, compare its value with `==`
or compute the sum, then compare-and-swap the pointer against the one loaded and retry if another
store won the race. Every store allocates a fresh cell, so the pointer identifies a version and the
loop needs no lock. `Once.do` and `Once.value` record a panic so later callers re-panic with the
same message, which Go's `sync.Once` does not do. `main` returning makes `Run` return after flushing
stdout, whatever other goroutines are doing, and the command then exits. Interpreter state outside
the `value` package is immutable after loading, like the AST and its annotations; per run, like
module records and tags, whose functions close over one run's modules; or per goroutine, like call
stacks, deferred lists, and the cycle sets of `==` and `String.from`. The stdout buffer is the one
exception and has a mutex.


**Output.** `print` writes `String.from(v)` and a newline to the run's stdout writer, buffered
behind a mutex and flushed when `Run` returns, before any stderr report, and before any blocking
call, so interleaving with stderr and with `Time.sleep` reads naturally. Panic reports go to the
run's stderr writer. The command passes `os.Stdout` and `os.Stderr`; a host passes what it likes.

**Cancellation.** `Run` takes a `context.Context`. The instance keeps one atomic `done` flag, set
when the context is cancelled and when `Run` returns; every call and loop back-edge checks it,
which costs one atomic load, and a goroutine that finds it set unwinds like a panic, running
deferred blocks under the same check, and ends without a report. Blocking library calls select on
the context too. This is how a host stops a runaway script, and how goroutines left behind by a
returned `main` are reaped, since Go offers no way to kill a goroutine.

## 6. Library

Every namespace is a record value whose fields are builtins; `Result`, `Mutex`, and the other types
with functions are tags whose `funcs` hold them, so `Result.ok` and `Array.map` are both field
reads. Builtins check argument kinds and panic with the library row of the panics table. A host
namespace is registered the same way, through `ant.WithNamespace(name, signatures, funcs)`, where
`signatures` is Ant declaration text the checker reads and `funcs` maps each name to a Go function.
The I/O namespaces of `stdlib.md` are registered through this path too, so a host can leave them out
or replace `File` with a sandboxed one.


Order of implementation, driven by what section 13 and the spec's other examples use:

1. `print`, `panic`, the twelve conversion functions, `Result` (`ok`, `err`, `isOk`, `map`,
   `mapErr`, `unwrap`, `unwrapOr`, `all`), `Error`, `Ok`, `Err`, `Entry`, `Indexed` tags.
2. `Array`, `Map`, and `String` in full, including `String.from`, and `Record.same`.
3. `Iter`, `Math`, `Bits` (with width-aware shifts and wrapping arithmetic per kind).
4. Sync types, `Time`, `Os`.
5. `File`, `Http` (`net/http`; a request blocks only its own goroutine), `Json` (`encoding/json`
   into `Json.Value` shapes: `float64`, `string`, `bool`, `null`, arrays, ordered maps;
   `stringify` renders records as objects without their tag).

Callbacks into Ant from a builtin (`Array.map`, `Mutex.withLock`) go through `interp.Call`, which
is re-entrant and is also what `Program.Func(name).Call` exposes to a host, with Ant panics
converted to `*PanicError` at that boundary. Inside the interpreter, panics propagate as Go panics
through the builtin, and `withLock` releases its lock in a Go `defer`.

## 7. Checker (`ant check`)

The checker is the largest piece and is scheduled after `ant run` passes the run corpus.
It shares the parser and the resolver annotations and never runs code.

**Type language** (`internal/types`): primitives by kind, `null`, `void`, `never`, `Array[T]`,
`Map[K, V]`, function types (parameter types, result, number of defaulted parameters), anonymous
record types, named types (a declaration plus type arguments, compared by declaration identity
and argument equality), unions (flattened, deduplicated, `never` dropped, discriminator rule
checked wherever a union is formed), and type parameters with bounds. Aliases expand to their
bodies with substitution when a type expression is resolved; a cycle of aliases, or a self
reference not through a record, array, map, or function, is an error.

**Declarations.** Each module gets an export table: types, functions with their signatures,
functions on types, and `var`s with their declared or inferred types. Exported functions must
state a return type. Return types of unexported functions are inferred on demand with an
in-progress marker so direct or mutual recursion without an annotation is an error.

**Expressions** are checked bidirectionally: `check(expr, expected)` and `synth(expr)`. A generic
call solves type parameters by unification in the spec's fixed order (expected type, non-lambda
arguments left to right, lambdas last); an unsolved parameter is an error. Lambdas take their
arity and parameter type from the expected function type and are an error without one. `if`
branches check against the expected type or join to a union. Assignability follows section 4
exactly, including invariance of type arguments, width subtyping into anonymous record types
only, and the function-type rule with defaults.

**Narrowing** is a flow analysis over locals and parameters only: a map from variable to current
type threaded through statements, split by conditions (`is`, `== null`, `!= null`, through `&&`,
`||`, `!`), joined at merges, reset by assignment to the assigned type, and cancelled from the
loop head when a loop body assigns. A pre-pass per function records which locals are assigned
inside a nested body (never narrow in the enclosing function) and which are never reassigned
(keep their narrowing inside nested bodies). A chain of `is` tests that narrows the subject to
`never` makes the `if` complete, which is the exhaustiveness check.

**Everything else** the spec calls a compile error gets a rule: reading `_`, a type name in
expression position, `is` against an alias or an always-true test, `?.` and `??` on a
non-nullable left side, `try` outside a `Result`-returning function, assignment through a union,
assignment to an imported name, a top-level initializer reading a later `var`, a literal head on a
non-record type, mixing tagged and untagged operands, and so on. The checker walks every module
in the import graph, including unused code, and returns every error it finds as
diagnostics sorted by position; the command prints them and exits 1 if there are any.

Tests: a `tests/check/` corpus in the same trailer format, each file either clean or annotated
with `// error: line: substring` lines. The spec marks its compile errors inline, so every such
comment in `ant-lang.md` becomes one case, and every program under `tests/run` must check clean.

## 8. REPL

A bare `ant` runs the loop in `cmd/ant` over the library's `Session`: a persistent instance
whose module record and environment grow with each input. It reads from stdin with a `> ` prompt
and a `. ` continuation prompt, accumulating input while brackets are unbalanced or the last
token cannot end an expression. Each input is parsed in a session mode that also allows
top-level expressions and redeclaring a name (which replaces the binding), then evaluated on the
instance; `Session.Eval` returns the value, which the loop prints with `String.from` unless it
is `void`. An Ant panic comes back as a `*PanicError`, is printed, and the session continues;
goroutines started by earlier inputs keep running. Ctrl-D ends it. The spec does not describe a
REPL, so this behaviour is a project decision; line editing and history can come later through
`golang.org/x/term`, in `cmd/ant` only, so the library stays dependency-free.

## 9. Phases

Each phase ends with its tests green and `go vet` clean. Rough sizes are lines of Go.

| Phase | Work | Done when | Size |
| ----- | ---- | --------- | ---- |
| 0 | `go.mod`, the root package's types and options as stubs, `cmd/ant` with the three commands and exit codes, the trailer format for `tests/run`, `conformance_test.go` that drives `Load` and `Run` in process, and the first programs: section 13 verbatim, and one per rule of `String.from` | the harness runs and every program fails for want of an interpreter | 400 |
| 1 | token, lexer, AST, parser, resolver, scope rules | every spec block parses; the syntax-error and AST tables pass; a program with a syntax error exits 2 | 2 500 |
| 2 | values: kinds, tags, shared slots, atomic-header arrays, locked ordered map, equality, hashing, rendering, numeric ops and conversions | unit tests for equality, hashing, rendering, overflow, and conversions: one case per row of the section 12 conversion table and per rule of `String.from`, including cyclic values and map insertion order | 1 500 |
| 3 | loader over a source provider, environments, evaluator, calls, closures, control flow, defer and panic, destructuring, `try`, `is`, templates, modules, `PanicError`, `ExitError`, cancellation | programs written for sections 3, 5, 6, 7, 8, 10, and 12 pass: scopes and closures, evaluation order, defer and panic order, destructuring, nullability, `try`, modules and initialization order, and one program per row of the panics table | 2 500 |
| 4 | core library (section 11) in the order of section 6 above | one program per namespace of section 11 exercising every function passes; section 13 passes | 2 500 |
| 5 | goroutines, sync types, `Time`, `Os` | programs for section 9 pass: results through containers, `WaitGroup`, `Mutex`, `Once`, `Atomic`, a panic in a goroutine; a stress program with 1 000 goroutines racing on one array, one map, and one record runs clean under `go test -race` and never crashes | 800 |
| 6 | host namespaces (`WithNamespace`, `Program.Func`), then `File`, `Http`, `Json` through that path, and `example_test.go` | one program per namespace of `stdlib.md` passes; the examples pass under `go test` and a host can run a script from an `fs.FS` with injected writers | 1 000 |
| 7 | checker, in sub-steps: types and declarations; expressions and generics; narrowing and exhaustiveness; modules and exports; diagnostics | `ant check` passes every program under `tests/run` and reports every compile error the spec shows inline | 5 000 |
| 8 | `Session` and the loop in `cmd/ant` | a `Session` test covers declarations, expressions, redeclaration, panics, imports; a manual session matches | 400 |

Phase 8 is small and independent of phase 7, so it can be slotted in whenever a demo of `ant`
is wanted. Phases 1 through 4 are strictly sequential; phase 5 and 6 can proceed in parallel with
phase 7.

## 10. Open points

Decisions taken here that the spec leaves open, recorded so they can be revisited:

- Float rendering of very large or small magnitudes (plain decimal, no exponent).
- Panic report format: `panic: <message>` followed by `  at <function> (<file>:<line>:<col>)`
  lines, innermost first. Programs check only that the message appears.
- Call depth limit of 10 000.
- Shared slots allocate a cell per store. If profiling shows it matters, a later optimization is
  a tagged one-word value representation; nothing outside the `value` package would change.
- `Os.exit` does not run deferred blocks, as the spec's `never` suggests; in the library it
  unwinds to `Run` as an `*ExitError` rather than ending the process.
- Module path `github.com/ant-src/ant`, inferred from the spec repository's remote; confirm before
  the first commit, since changing it later breaks importers.
- The `go` directive in `go.mod` is the lowest version the code needs, not the installed 1.27, so
  importers are not forced to upgrade.
- Goroutines still running when `Run` returns are stopped at their next check rather than killed,
  since Go cannot kill a goroutine; a goroutine blocked in a host call it does not cancel lingers.
- Buffered stdout with the flush points listed in section 5.
- The checker's handling of the I/O namespaces depends on `prelude.ant` carrying their
  signatures, which `stdlib.md` calls a sketch; they will be transcribed as written.
