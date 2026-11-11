<!-- status: reviewed -->

# Language reference

The language reference has one page for each keyword, operator, literal, and built-in type, and for
each rule of how source text is read. This page lists them all, one line each. The functions of the
standard library are in the [library reference](../lib/prelude/index.md).

A few principles run through every page:

- Every construct is an expression with a type. `if` and blocks give values; a statement is an
  expression whose value is `void`.
- No value changes type without a visible conversion. A condition is a `bool`, and numbers of two
  types never mix.
- Errors are values: a `Result` carries an expected failure and `null` carries absence. A panic is
  for bugs.
- When two constructs would do the same job, the language keeps one.

```ant
let n = 7
let parity = if n % 2 == 0 { "even" } else { "odd" }
parity           // "odd"
float64(n) / 2.0 // 3.5
let name: string | null = null
name ?? "anonymous" // "anonymous"
```

## Lexical structure

Source files are UTF-8 text with the extension `.ant`.

- [Comments and doc comments](lexical/comments.md): `//` and `/* */` comments, and the doc comments
  that `ant doc` shows.
- [Identifiers](lexical/identifiers.md): the names of variables, functions, types, parameters, and
  fields.
- [Keywords and reserved words](lexical/reserved.md): the words that cannot name a variable,
  function, type, or parameter.
- [Line endings](lexical/line-endings.md): a newline ends a statement, and a few signs continue it
  onto the next line.

## Literals

- [Number literals](literals/numbers.md): `42`, `0xFF`, and `1_000` are `int`; `3.14` and `1e9` are
  `float64`.
- [String literals and interpolation](literals/strings.md): `"Hello, ${name}"`, with escapes.
- [Raw strings](literals/raw-strings.md): `` `C:\temp` ``, with no escapes and no interpolation.
- [true and false](literals/booleans.md): the two values of `bool`.
- [null](literals/null.md): marks an absent value.
- [Array literals](literals/arrays.md): `[1, 2, 3]`, and `...` to spread another array.
- [Map literals](literals/maps.md): `Map{ key: value }`.
- [Set literals](literals/sets.md): there are none; `Set.of([1, 2, 3])` builds a set.
- [Record literals](literals/records.md): `Point{ x: 1, y: 2 }`, with field shorthand and `...`.
- [Lambdas](literals/lambdas.md): `{ x * 2 }` as a function value, and anonymous `func`.
- [Regular expression literals](literals/regex.md): `/^\d+$/` and `/ant/i`.
- [HTML expressions](literals/html.md): `<p>Hello, {name}</p>` builds an `Html` value.

## Keywords

### Declarations

- [let](keywords/let.md): a binding that keeps its value, and record and array destructuring.
- [var](keywords/var.md): a binding that can be assigned again, with the same destructuring as
  `let`.
- [func](keywords/func.md): a named function, an anonymous function, or a function on a type.
- [type](keywords/type.md): a record type, a named primitive type, or an alias, and `type(T)`, a
  type as a value.
- [interface](keywords/interface.md): a named set of function signatures that a type satisfies.
- [import](keywords/import.md): the exports of another module, as names or as a namespace.
- [export](keywords/export.md): a top-level declaration of a library that other modules can import.
- [test](keywords/test.md): a test or subtest in a test file, which `ant test` runs.

### Control flow

- [if and else](keywords/if.md): one of two branches by a `bool`, as an expression.
- [for and in](keywords/for.md): a block for each element of an array, map, set, string, `Bytes`, or
  sequence.
- [while](keywords/while.md): a block as long as a condition is `true`.
- [break](keywords/break.md): ends the innermost loop.
- [continue](keywords/continue.md): starts the next iteration of the innermost loop.
- [return](keywords/return.md): leaves the innermost function, lambda, script body, or `go` or
  `defer` block early, with a value or none.

### Errors and concurrency

- [try](keywords/try.md): unwraps a `Result`, or returns its error from the function.
- [defer](keywords/defer.md): a block that runs when the enclosing block exits.
- [go](keywords/go.md): runs a block on a new goroutine.
- [select](keywords/select.md): waits for the first of several blocking calls that can proceed.

`true`, `false`, and `null` are under [Literals](#literals), `is` under [Operators](#operators), and
`void` under [Built-in types](#built-in-types). `panic`, `assert`, `len`, `println`, and the
conversions such as `float64(n)` are prelude functions, in the
[library reference](../lib/prelude/index.md).

## Operators

- [Arithmetic](operators/arithmetic.md): `+`, `-`, `*`, `/`, `%`, and unary `-`; `+` also joins
  strings.
- [Comparison](operators/comparison.md): `<`, `<=`, `>`, and `>=` on numbers and strings.
- [Equality](operators/equality.md): `==` and `!=`, which compare structure.
- [Logical](operators/logical.md): `&&`, `||`, and `!` on `bool`.
- [is](operators/is.md): tests the member of a union and narrows to it.
- [Null coalescing](operators/null-coalescing.md): `a ?? b` replaces `null`.
- [Optional chaining](operators/optional-chaining.md): `a?.b`, `a?.[i]`, and `f?.(x)` through a
  value that may be `null`.
- [Indexing](operators/indexing.md): `a[i]` on an array, `Bytes`, or a map.
- [Calls through .](operators/dot-calls.md): `x.f(a)` calls a field, or `f` of the namespace of
  `x`'s type.
- [The pipe](operators/pipe.md): `a | f(b)` calls `f(a, b)`.
- [Assignment](operators/assignment.md): `=`, `+=`, `-=`, `*=`, `/=`, and `%=`.
- [Operator precedence](precedence.md): which operator takes an operand first.

Bitwise operations are functions of the [Bits](../lib/Bits/index.md) namespace, not operators.

## Built-in types

- [bool](types/bool.md): `true` and `false`.
- [int and the sized integers](types/int.md): `int`, and `int8` to `int64` and `uint8` to `uint64`.
- [byte](types/byte.md): another name for `uint8`, the element of `Bytes`.
- [float32 and float64](types/float.md): IEEE 754 floating-point numbers.
- [string](types/string.md): immutable Unicode text.
- [void](types/void.md): the type and value of code that runs for its effect.
- [never](types/never.md): the type of an expression that does not finish.
- [Array](types/Array.md): `Array[T]`, an ordered, growable list.
- [Map](types/Map.md): `Map[K, V]`, a hash map in insertion order.
- [Set](types/Set.md): `Set[T]`, distinct values in insertion order.
- [Function types](types/functions.md): `func(A, B): R`.
- [Self](types/Self.md): the type that satisfies an interface, in its signatures.

`null` is the type of [null](literals/null.md). [Bytes](../lib/Bytes/index.md) is built in for
binary data. Records, unions, and generics have chapters of their own in the guide:
[Records](../guide/records.md), [Unions](../guide/unions.md), and [Generics](../guide/generics.md).

## At run time

- [Runtime panics](panics.md): every way a checked program can panic.
