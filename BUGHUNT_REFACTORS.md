# Bug hunt 2026-10-24: simplifications and refactors for later

These are behavior-preserving cleanups noticed while hunting bugs. None are done. Several would
have prevented a bug fixed on this branch, noted as "(bug: ...)".

## Checker (internal/check, internal/types)

- **Two discriminator functions.** `discriminatorOf` in `call.go` duplicates
  `types.discriminator`. It keyed named types by name and field count (bug: two modules' `Point`
  were confused); it now uses the declaration pointer. Export one and use it in both places. They
  still differ on opaque sync types, so check that difference first.
- **`types.Expand` swallows substitution errors** and returns `Invalid`, which is accepted
  everywhere. Since `subst` now validates alias bodies (bug: inference built an invalid union),
  `Expand` could return the error, or assert that it cannot happen.
- **Two loop helpers.** `whileLoop` and `forLoop` handle the flow differently (`wf` vs `before`,
  break states). One helper that collects break flows would serve both (bug: the state at a
  `break` was ignored).
- **`resetAssigned`/`walkAssigned`** re-walks the AST by name (~90 lines). The resolver already
  knows which slots are assigned; facts scoped per loop would replace the walk.
- **One-line wrappers.** `synthBlock` and `checkBlock` only call `block`.
- **Phase 3 of the earlier refactor is still open** (see memory): stdlib parameter lists from
  `prelude.ant`; the checker using resolver bindings; top-level `var` in both `Decls` and `Stmts`.

## Parser and lexer

- **End of input.** The lexer's `peek` returns 0 at end of input, and the string scanners tested
  `c == 0` (bug: a NUL byte ended the string). `atEnd` now exists; using it at every
  end-of-input check would drop the sentinel entirely.
- **Overlapping keyword sets.** `endingKeywords` and `operandKeywords` in the lexer overlap, and
  the parser's `kindNames` repeats prelude knowledge.
- **`_` handling** is special-cased in `resolve.declare`, `check.assign`/`assignPlace`/`ident` and
  `interp.bind`/`setParam`. One "discard target" concept would cover them (bugs: a `_` parameter
  crashed or overwrote another parameter; `_ += 1` passed the checker).
- **`spreadEnd` map.** The parser keeps a side map from an array literal to its `[` token and a
  trailing-comma flag, for when the literal turns out to be a destructuring assignment. Fields on
  `ast.ArrayLit` would be simpler, but that is an AST change.

## Formatter

- **Duplicated spacing logic.** `format.inline` reimplements the gap and spacing logic of `seq`.
  One shared "space between elements" helper would keep templates and code in sync.

## Interpreter (internal/interp, value)

- **Arity checks after `push`.** `callFunc` pushes the frame before validating arity, so arity
  panics show a spurious `at f (caller position)` frame. A call-depth overflow also prints all
  10,000 frames; truncating the trace would help.
- **`forLoop` repetition.** It repeats the same `g.check(); if rv, rc, stop := run(v); stop {…}`
  block for five iterable kinds. One `each(func(v) bool)` per kind would remove it.
- **`value/numeric.go`.** `Binary` and `Compare` repeat the same 10-way type switch. A generic
  helper per kind would roughly halve them.
- **Recursive walks.** `value/equal.go`, `value/render.go` and `lib/json.go` stringify all
  recurse once per level of nesting and hop to a fresh goroutine every 4096 levels
  (`value.Nested`; see question 9 in BUGHUNT_QUESTIONS.md). An explicit stack would make that
  unnecessary. The `ancestors` set in `json.go` and the map in `render.go` could share one type.
- **`runOnce` and `g.Catch`.** Now that `Once.Do` completes in a `defer`, `runOnce` could use a
  plain `defer`/`recover` that stores any `*AntPanic` instead of `g.Catch`.

## Standard library (internal/lib)

- **`pad` (`string.go`).** It builds a `[]rune` of the full width; `strings.Repeat` plus a cut by
  code point would be simpler and use about a quarter of the memory.
- **`String.toFloat`.** It rejects input with a character blacklist (`"_xXpP"`, any `i` or `n`)
  before `strconv.ParseFloat`. An explicit grammar check, like `digitsOf`, would be clearer.
- **`String.indexOf`** repeats the null and rune-count logic of the `runeIndex` helper that
  `lastIndexOf` uses.
- **`String.fromInt`.** Its type switch duplicates `intBits` in `bits.go`.
- **`Array.removeAt`.** It checks `i <= Len` and relies on `Remove` failing at `i == Len`; check
  `0 <= i < Len` directly.
- **`Result.map`/`mapErr`** call `g.Call(a[1], …)` without `funcArg`, unlike every other
  callback-taking function.
- **`Map.merge`/`Set.union`** copy entry by entry. A clone-then-insert helper on the value types
  would remove the duplication.
- **`json.go`:**
  - `writeJSONString` builds a `json.Encoder` for every string.
  - `holdsBytes` re-walks every map key.
  - Parsing goes through the `json.Decoder` Token API, which is where the 10000-depth cap and the
    Go-worded errors come from. A small recursive-descent parser would control both.
- **`time.go`.**
  - The `dated` and `timed` maps are duplicated by separate ordered name lists; a slice of
    `{name, fn}` pairs would do.
  - `addPoint`, `untilPoint`, `compareTime` and the `round`/`with` switches repeat the same
    five-way type switch, which a small interface in `temporal` could replace.
- **`temporal` `NextTransition`/`PreviousTransition`** rely on quirks of Go's `ZoneBounds` (24 h
  stepping). They are correct in every zone tried, but fragile.
