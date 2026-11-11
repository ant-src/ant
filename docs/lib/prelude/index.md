<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Prelude

<!-- end generated -->

The prelude declares the names every module sees without an `import`: the output functions, `panic`
and `assert`, the conversions, `len`, a few types, and the namespaces of the standard library, such
as `Array`, `String`, `File`, and `Time`. This page covers the names that stand on their own; each
namespace has its own page.

## Details

- No declaration, in any scope, may take a prelude name, so a local cannot be named `len`, `string`,
  or `print`.
- A namespace such as `Os`, `Time`, or `Reflect` is reached only through `.`. It is not a value.
- `assert` is not a value either: it can only be called by name.
- A conversion to an integer type panics rather than lose a value: out of range, a fraction, `NaN`,
  or an infinity. `float32` and `float64` round and never panic.
- `ant doc` shows the declaration and doc comment of any prelude name, such as `ant doc println`.

## Basic examples

`println` writes a value and a newline:

```ant
let name = "Ant"
println("Hello, ${name}")
println([1, 2, 3])
```

```text
Hello, Ant
[1, 2, 3]
```

`len` counts, and a conversion changes a number's type:

```ant
len([10, 20, 30]) // 3
len("héllo")      // 5
int8(100)         // 100
float64(7) / 2.0  // 3.5
```

`assert` stops the program when a condition is false:

```ant
let total = 2 + 2
assert(total == 4, "arithmetic")
```

## Scope

The functions of a namespace are called through it, through `.` on the value, or down a pipe:

```ant
let words = ["ant", "bee"]
Array.map(words, { String.toUpper(x) }) // ["ANT", "BEE"]
words.map({ x.toUpper() })              // ["ANT", "BEE"]
words | Array.map({ len(x) })           // [3, 3]
```

`Result`, `Ok`, `Err`, and `Error` are the types the library reports failure with:

```ant
let r: Result[int, Error] = Ok{ value: 2 }
r is Ok // true
let e: Result[int, Error] = Err{ error: Error{ message: "no" } }
e.unwrapOr(0) // 0
```

`Integer`, `Number`, and `Ordered` bound a type parameter, so one function takes every number type:

```ant
func half[T: Number](x: T): float64 {
  float64(x) / 2.0
}
half(3)       // 1.5
half(int8(5)) // 2.5
half(0.5)     // 0.25
```

A map yields `Entry` records, and `Array.entries` yields `Indexed` ones:

```ant
let ages = Map{ "Ann": 31, "Bob": 27 }
ages.entries()  // [Entry{ key: "Ann", value: 31 }, Entry{ key: "Bob", value: 27 }]
["a"].entries() // [Indexed{ index: 0, value: "a" }]
```

A conversion also unwraps a named primitive type:

```ant
type Count = int
let c = Count(3)
int(c) + 1 // 4
```

## Names by job

### Output

- [print](print.md): write a value to standard output, with no newline
- [println](println.md): write a value and a newline
- [eprint](eprint.md), [eprintln](eprintln.md): the same, to standard error

### Stopping

- [panic](panic.md): stop the goroutine with a message
- [assert](assert.md): panic when a condition is false

### Size

- [len](len.md): the size of an array, a map, a set, a string, `Bytes`, or a channel

### Conversions

- [int](int.md), [int8](int8.md), [int16](int16.md), [int32](int32.md), [int64](int64.md): to a
  signed integer, exactly
- [byte](byte.md), [uint8](uint8.md), [uint16](uint16.md), [uint32](uint32.md), [uint64](uint64.md):
  to an unsigned integer, exactly
- [float32](float32.md), [float64](float64.md): to a float, rounding to the nearest value
- [string](string.md), [bool](bool.md): the payload of a named primitive type

### Types

- [Result](../Result/index.md), with `Ok` and `Err`: a value or an error
- [Error](../Error/index.md): the error the library returns, a message
- [Panic](../Panic/index.md) and [Panic.Frame](../Panic/types/Frame.md): a panic that
  [Panic.catch](../Panic/catch.md) stopped
- [Entry](../Map/types/Entry.md): a key and its value
- [Indexed](../Array/types/Indexed.md): an element and its index
- [Seq](../Iter/types/Seq.md): a sequence, which a `for` loop iterates
- [Integer](../Math/types/Integer.md), [Number](../Math/types/Number.md),
  [Ordered](../Math/types/Ordered.md): bounds for a type parameter

### Namespaces

The [standard library](../index.md) lists every namespace, from [Array](../Array/index.md) and
[String](../String/index.md) to [File](../File/index.md) and [Http](../Http/index.md).

## Applications

Report a problem on standard error and the result on standard output:

```ant
let scores = [72, 95, 88]
if len(scores) < 5 {
  eprintln("warning: only ${len(scores)} scores")
}
println(scores.sum() ?? 0)
```

```text
255
```

Turn a fraction into a percentage, rounding before the exact conversion:

```ant
let done = 7
let total = 9
int(Math.round(float64(done) / float64(total) * 100.0)) // 78
```

## Properties and relations

`len` agrees with the size function of each namespace:

```ant
assert(len([1, 2]) == Array.length([1, 2]))
assert(len("héllo") == String.length("héllo"))
assert(len(Map{ "a": 1 }) == Map.size(Map{ "a": 1 }))
```

A call through `.` is the call through the namespace:

```ant
let s = "  hi  "
assert(s.trim() == String.trim(s))
```

## Possible issues

A namespace is not a value, so it cannot be stored:

```ant
let t = Time // error: Time is a namespace, not a value
```

Neither can `assert`:

```ant
let check = assert
// error: assert is not a value; it can only be called, as assert(cond) or assert(cond, message)
```

A prelude name cannot be redeclared, even as a local:

```ant
let len = 3 // error: len is a prelude name and cannot be redeclared
```

A conversion to an integer panics on a fraction. Round first:

```ant
int(3.5)             // panic: 3.5 has a fraction, so it does not convert to int
int(Math.round(3.5)) // 4
```

`len` takes only collections, strings, `Bytes`, and channels:

```ant
type Point = { x: int, y: int }
let n = len(Point{ x: 1, y: 2 })
// error: len needs an array, a map, a set, a string, Bytes, or a channel, got Point
```

## See also

[Standard library](../index.md), [Modules](../../guide/modules.md), [Errors](../../guide/errors.md),
[ant doc](../../tools/doc.md), [Calls through .](../../language/operators/dot-calls.md),
[Pipe](../../language/operators/pipe.md)
