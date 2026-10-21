# Spec feedback

Points where the implementation found the specification silent, ambiguous or, in one place,
inconsistent with itself. Each entry says what the implementation does today. An entry is deleted
once the spec settles it, and the tests follow the spec. Line numbers refer to `spec/ant-lang.md`
at the time of writing.

## Chapter 9, Visibility (line 1203)

The text gives the necessary condition, a write is visible to a read "only when something
orders them", but never the sufficient one: nothing promises that a write ordered before a read
is seen by it. Every example under Sharing results relies on that promise. Go's rule has two
halves and the spec could state both: a read observes the write that happens before it when no
other write to that location is ordered between them, and a racy read observes some write that is
not ordered after it and not followed by another write ordered before the read.

The implementation gives the strong guarantee through Go's own memory model.

## Chapter 9, Data races (line 1185)

- "with the ordinary index or key error" for a racy map operation: a missing key yields `null`
  everywhere else in the spec and the panics table has no key error. Suggested: "or finds the key
  absent".
- "before or after the concurrent mutation" should be per mutation: a multi-element operation such
  as `Array.map` racing two index stores can see one and not the other.
- The pricing sentence: a hash table cannot be read safely during a resize with one lock per
  mutation only; reads need a shared lock. Suggested: one lock per map operation, shared for
  reads. Push and pop on an array also need mutual exclusion against each other, not only an
  atomic header.
- The race definition names "one variable, field, or element"; the length of an array and the
  key set of a map are shared state too, and the second bullet already treats them so.
- Worth saying: the first bullet is what keeps the checker's types true under races, since every
  store into a slot has the slot's declared type. And the pricing paragraph describes the runtime,
  so it could be marked as an implementation note the way chapter 12 marks slot resolution.

## stdlib.md, `Os.exit`

Whether deferred blocks run. The implementation does not run them, reading `never` as an exit
without unwinding; inside the library it surfaces as an `*ExitError` from `Run`.

## Chapter 4, Function types (line 465)

`func(): int | Meters` parses greedily as a function returning the union. A reader may expect the
union of a function type and `Meters`. One sentence settling the precedence would help; the
implementation follows the greedy reading.

## Chapter 7, Narrowing (line 994)

Assignment narrows a local to the assigned type, so `v = "plain"` followed by `v is UserId` is an
always-false compile error. The rule is right and the implementation follows it, but an example
would prevent the surprise.

## Chapter 5, Loops (line 747)

A map literal in a `for` header needs parentheses, since the `{` would otherwise start the body.
Worth stating.

## Chapter 2, Literals (line 69)

`1.x` is a lexer error rather than `1` followed by `.x`. Worth stating.

## stdlib.md, `Json.stringify` on a cyclic value

`String.from` renders a value already being rendered as `...` (line 1478) and map hashing uses a
bounded prefix (line 1601), so both terminate. `Json.stringify` is described as rendering "as
`String.from` would", but `...` is not JSON. Decided for the implementation: `Json.stringify`
panics with the message `cyclic value`. The chapter 12 table could gain it as a ◆ case of the
library function row.

## Chapter 5, Destructuring (line 793)

"The checker types each binding from the source, `T` for an element of `Array[T]`", so
`var [n, s] = [1, "t"]` gives both `n` and `s` the type `int | string`, and the assignment form
`[n, s] = [1, "t"]` with `n: int` and `s: string` is a compile error, although the runtime would
store 1 and "t". The implementation follows the sentence as written; the spec may want to say so
explicitly, or to give a literal on the right of a destructuring element-wise types.

## Chapter 11, `String.from` on floats (line 1430)

The rendering of floats of very large or very small magnitude is not stated. The implementation
prints plain decimal digits without an exponent.
