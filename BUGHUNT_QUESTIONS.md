# Bug hunt 2026-10-24: decisions that need your input

Found while hunting bugs on branch `bugfix-2026-10-24` (ant and spec). Everything here is either a
spec question, a fix I made that changes the spec's promises and you may want to undo, or a bug
whose fix is a design choice. Probe programs that show each one are in the session scratchpad
(not committed).

## Decisions I took; confirm or revert

1. **NaN in `Array.sort` and `Array.sortBy` sorts after every number** (spec + impl). With NaN in
   the array the sort broke down and left even the other numbers unsorted, so some order had to
   be picked. The other choice is NaN first, as Go's `sort.Float64s` does.
2. **`Math.min` and `Math.max` are NaN when either argument is** (spec + impl). Before, the result
   depended on argument order. This matches `Array.min`. The order of `-0.0` and `0.0` is still
   unspecified: `Math.max(0.0, -0.0)` is `0.0` and `Math.max(-0.0, 0.0)` is `-0.0`.
3. **`Array.filled`, `Array.range` and `String.fromFloat` bound their counts at 2^31 − 1** and are
   now in the "a count" row of the section 12 table. `Array.filled(Math.maxInt, 0)` crashed the
   interpreter with a Go panic. Note that the bound limits the count, not the bytes:
   `String.repeat("abcdefgh", 2147483647)` still tries to allocate 16 GB.
4. **Concurrent `yield` calls run the `for` body one at a time.** A sequence that called `yield`
   from several goroutines ran the body on all of them at once. That wrote the loop function's
   plain (unboxed) slots concurrently, tore them, and segfaulted the interpreter. A per-loop lock
   now serializes body runs; a `yield` inside the body on the same goroutine re-enters it.
   - Trade-off: a body that blocks waiting for another goroutine that is itself blocked in `yield`
     now deadlocks.
   - The alternative is for the resolver to treat a `for` body as a capture boundary, so outer
     locals used there become shared slots. That keeps full concurrency but costs an atomic store
     on every such local in every loop.
   - The spec does not say whether concurrent `yield` is allowed.
5. **`Once.value` panics when its function was left by a `return`** that unwound through it, from
   a `for` body whose `yield` the function called. It has no value to give. This is a new
   documented panic. `Once.do` in the same situation now counts as done instead of hanging every
   later caller forever.
6. **An empty list holding only a block comment is formatted broken** (`f(\n  /* none */\n)`).
   `[ // line` followed by `]` keeps the line comment on the bracket's line, as before. fmt.md 4.4
   says "the comment takes the place of an item", which could also mean that one moves to its own
   line.

## Open questions (no change made)

7. **Regex `\Q...\E` and source normalization.** Normalization turns `\/` into `/` and a line break
   into `\n` everywhere, including inside `\Q...\E`, where RE2 takes every character literally. So
   `Regex.compile("^\\Q\\/\\E$")` stops matching the two characters `\/`, and a raw line feed
   inside `\Q` becomes the two characters `\n`. Section 2 says to "write `/` there", but a `/`
   ends a literal, so in a literal that is impossible.
   - Proposal: copy `\Q...\E` spans verbatim. Close and reopen the span around a line break
     (`\E\n\Q`) and around a `/` when rendering (`\E\/\Q`), so a pattern still renders as a
     literal that reads back to it.
8. **`Time.parseZoned("2024-09-08[America/Santiago]")`**, a date with no time. Today it resolves
   midnight through `disambiguation`, which on a DST gap day gives
   `2024-09-07T23:00-04:00`, the previous day. Should this be `Err`, as "each where its type needs
   it" suggests, or the start of the day, as in Temporal?
9. **Very deep values (fixed, confirm the approach).** `==`, `String.from` and `Json.stringify`
   overflowed the Go stack, a fatal crash, on a value nested about a million deep, such as a
   linked list built in a loop. Section 12 says `==` never panics, and section 9 says the
   interpreter never crashes.
   - Fix: the walks stay recursive, but every 4096 levels they continue on a fresh goroutine and
     wait for it (`value.Deep` and `value.OnFreshStack`), so no single Go stack overflows.
   - Cost: memory proportional to depth, about 1 KB per level, and a goroutine per 4096 levels.
   - Cost on shallow values: a set that may reach another goroutine cannot live on the stack, so
     `==` and `String.from` take theirs from a `sync.Pool`. A loop of 2M `==` on a small record
     takes 0.38 s against 0.33 s before the change (it was 0.53 s with a fresh set per call).
     `String.from` also gives up stack allocation of its builder, about 13% on a small record.
   - Alternative: an iterative walk with an explicit stack. That is harder for `==` because map
     lookups nest comparisons that may fail without affecting the outer one.
10. **The `Iter` functions silently ignore `yield` calls after `false`**, while `for` panics on
    them. The spec names only `for`.
11. **An unchecked top-level `var x = try f()` that meets an `Err`.** The checker rejects it. At
    run time it now panics with a real message instead of "internal error". The case is not a row
    of the section 12 table. The options are to add a row, or to have the parser reject `try` in a
    top-level initializer as it rejects `return`.
12. **Inside a generic function, `x is int` with `[T: Ordered]` narrows `x` to plain `int`.** So
    `if x is int { x } else { x }` returning `T` is rejected. Should the narrowed value still be
    a `T`?
13. **Expected type versus a union bound.** `var r: int | string = bigger(1, 2)`, with
    `[T: Ordered]`, works because the implementation skips expected-type solutions that break a
    bound. Section 4 says inference uses "the expected type first" and does not mention this
    exception.
14. **Does a nullable result count as "a Result" for `try`?** In a function returning
    `Result[int, E] | null`, `try` reports "try would return Err[E], but the function returns
    Result[int, E] | null", although `Err[E]` is assignable to that type.
15. **Case of identifiers.** Section 2 says "Case carries no meaning", but identifiers are
    case-sensitive: `value`, `Value` and `VALUE` are three variables, and the corpus tests this.
    doc.md matches names "without regard to case". One of the two texts needs rewording.
16. **`Time.compareDuration` of two identical durations with calendar units and no
    `relativeTo`.** Temporal's `Duration.compare` returns 0 for identical durations before it
    looks at `relativeTo`. The branch now panics first, which the spec's "panics without one"
    allows but which departs from Temporal, which Ant follows elsewhere. Keep the panic, or return
    0 as Temporal does and say so in the spec?

## Smaller spec gaps the hunters noted

- Lexing:
  - Number spellings `0X`, `0B` and `E` are accepted; leading zeros (`007`) are unspecified.
  - `1e400` is a syntax error, but `1e-400` silently becomes `0.0`.
  - Section 2 never defines a lone `\r`, which fmt.md 4.1 refers to.
- Parsing:
  - A `return` at the end of a line, followed by a line starting with `.` or `|`.
  - `select` clauses separated by `;` are accepted, though the spec says "one clause per line".
  - `var [] = xs` and `var {} = r` are accepted.
  - An `if` header inside parentheses may put its `{` on the next line.
- fmt: tabs inside a multi-line `${...}`. 4.1 and 4.9 disagree, and the repository's own
  `FuzzFormat` fails on ``func A(){(`${\n\t0}`)}``.
- Rendering: `-0.0` renders as `-0.0`, where JavaScript gives `0`. The spec otherwise follows JS
  for the exponent form.
- `String.fromFloat`:
  - Is a tie decided on the decimal literal or on the binary value? It uses the binary value, so
    `fromFloat(1.005, 2)` is `"1.00"`.
  - Small negatives render as `"-0.00"`.
- Compound assignment reads its target before the right side runs (`x += { x = 10; 5 }()` gives
  6). Section 12 does not say so.
- `Random` compares by identity (section 4), but the section 12 equality list names only
  functions, tags and sync objects.
- `select` on a closed channel that still holds buffered values counts its receive twice, so it is
  favoured over the other ready clauses.
- `Time`:
  - `Time.with(day: 0)` and negative fields clamp, where Temporal's "constrain" throws.
  - Valid durations whose balanced fields do not fit an `int` panic, and no panic is documented.
  - The parser accepts more than the stated grammar: lowercase `t`/`z`, a comma as the decimal
    mark, `±HH`, `±HHMM`, `+YYYYYY` and `[!zone]`. It rejects leap second `:60`.
- `Json`:
  - The 10000-level depth limit is unspecified.
  - Errors leak Go wording, e.g. "json: cannot unmarshal number 1e400 into Go value of type
    float64".
  - With duplicate keys the last one wins.
  - Functions, channels and `Random` stringify silently as `null`.
- `Csv.stringify([[""]])` gives `"\n"`, which `Csv.parse` drops, so the row does not round-trip.
- `Url` cannot tell an empty authority from none (`x:/a` comes back as `x:///a`), and `%2F` in a
  path becomes `/` after a round trip.
- `Path.join([])` is `""`, though "cleans the result" suggests `"."`.
- After `File.readExactly` returns `Err("unexpected end of file")`, the partial bytes are kept for
  the next read.
- `Http.request` headers that differ only in case collapse silently.
- When a later caller of `Once.do` or `Once.value` repanics with the first run's panic, the stack
  trace shows the first goroutine's call site.
- A generic named primitive's unfixed phantom parameter leaks: `Array.map(["a"], Id)` with
  `type Id[T] = string` gives `Array[Id[T]]`.
- `type C[T] = C[T]` reports "would mint a second tag" instead of the alias-cycle error.
- `Time.round` on an `Instant` with `Hour` and increment 24 panics, as the spec's "and be smaller"
  requires. Temporal allows it.
