<!-- status: reviewed -->

# Types

Ant is statically typed, and you write few types. The type checker infers the type of a `let`, each
lambda parameter, and most return types; you annotate function parameters and the few places where
nothing else says what a value is.

This chapter shows what the two commands do with types, how the primitive types behave, why numbers
never mix, how named primitive types keep identifiers apart, and when one type can stand in for
another.

## Checking and running

Two commands read your program, and only one of them looks at types:

- `ant check` runs the type checker over the file and every module it imports, including code that
  nothing calls. It reports every compile error it finds, and exits with a non-zero status when
  there is one.
- `ant run` parses the program and runs it. It discards the annotations and checks nothing in
  advance.

```ant
func greet(name: string): string {
  "hello, " + name
}

println("start")
println(greet(42)) // error: cannot use int where string is expected
```

`ant check` rejects this program at the call. `ant run` runs it anyway: it prints `start`, then
panics inside `greet`, at the `+` that gets a `string` and an `int`. Without the checker, a type
mistake surfaces only when the program reaches it, as a panic, after the work before it has run.

Run `ant check` in your editor and in continuous integration, and `ant run` in the terminal. No flag
makes `ant run` check first.

## Primitive types

| Type                               | Values                                            |
| ---------------------------------- | ------------------------------------------------- |
| `bool`                             | `true`, `false`                                   |
| `int8` `int16` `int32` `int64`     | signed integers of 8, 16, 32, and 64 bits         |
| `uint8` `uint16` `uint32` `uint64` | unsigned integers of 8, 16, 32, and 64 bits       |
| `float32` `float64`                | IEEE 754 floats                                   |
| `string`                           | immutable text, a sequence of Unicode code points |
| `null`                             | the single value `null`                           |
| `void`                             | the single value `void`                           |
| `never`                            | the type of `panic(...)`, `return`, `break`       |

`int` is another name for `int64`, and `byte` for `uint8`. Use `int` and `float64` unless the size
matters, such as for binary data. Each pair is one type, shown by the short name whichever you
write:

```ant
type(int64)        // int
type(Array[uint8]) // Array[byte]
```

## Numbers never mix

Ant has ten numeric kinds, and an operator takes two values of the same one. There is no implicit
conversion between them, not even from `int` to `float64`:

```ant
let items = 7
let average = items / 2.0 // error: / needs both operands of one type, got int and float64
```

A literal gets no type from its context: an integer literal is always an `int`, and a float literal
always a `float64`. So a literal where another kind is expected is an error too. Convert it:

```ant
let level: int8 = 5 // error: cannot use int where int8 is expected
```

```ant
let level = int8(5)
let next = level + int8(1)
println(next)
```

```text
6
```

Integer `/` truncates toward zero, and `%` keeps the dividend's sign. Integer overflow and integer
division by zero panic instead of wrapping:

```ant
7 / 2                            // 3
-7 / 2                           // -3
-7 % 2                           // -1
7.0 / 2.0                        // 3.5
Math.maxInt + 1                  // panic: integer overflow
1 / 0                            // panic: division by zero
1.0 / 0.0                        // Infinity
Bits.wrapAdd(int8(127), int8(1)) // -128
```

Use [Bits.wrapAdd](../lib/Bits/wrapAdd.md) and its siblings when you want arithmetic that wraps.

## Conversions

To convert a number, call the target type as a function. Each of the ten numeric names takes any
numeric value:

```ant
let items = 7
float64(items) / 2.0  // 3.5
int(3.0)              // 3
byte(255)             // 255
float64(float32(0.1)) // 0.10000000149011612
```

A conversion to an integer kind must be exact. A fraction, a value out of range, `NaN`, or an
infinity panics, so the conversion never loses data without telling you. Round or truncate first
when you mean to:

```ant
int(3.5)             // panic: 3.5 has a fraction, so it does not convert to int
int8(300)            // panic: 300 out of range for int8
byte(-1)             // panic: -1 out of range for byte
int(1.0 / 0.0)       // panic: Infinity does not convert to int
int(Math.trunc(3.7)) // 3
int(Math.round(3.5)) // 4
```

A conversion to a float kind rounds to the nearest value the kind holds, and never panics.

Text is not a number, and calling `int` on a string does not parse it. Parse with
[String.toInt](../lib/String/toInt.md) and [String.toFloat](../lib/String/toFloat.md), which return
`null` for text that is not a number, and render with [String.from](../lib/String/from.md) or
interpolation:

```ant
"42".toInt()    // 42
"4x2".toInt()   // null
"2.5".toFloat() // 2.5
String.from(42) // "42"
"total: ${42}"  // "total: 42"
```

The reverse holds too: `string` converts only a string or a named primitive type over `string`, so
it does not render a number. Use `String.from` or interpolation:

```ant
let s = string(42) // error: string converts only a string or a named primitive type over string, got int
```

The conversion functions are values, so you can pass them to another function:

```ant
[1, 2, 3].map(float64) // [1.0, 2.0, 3.0]
```

## Named primitive types

A user id and an order id are both strings, but passing one where the other belongs is a bug. A
`type` declaration over a primitive type makes a new type that the checker keeps apart from its base
and from every other type over the same base:

```ant
type UserId = string
type OrderId = string

func findOrder(id: OrderId): string {
  "order ${string(id)}"
}

let user = UserId("u-1")
let order = OrderId("o-7")
println(findOrder(order))
println(findOrder(user)) // error: cannot use UserId where OrderId is expected
```

Call the type name to tag a value, `UserId("u-1")`, and the base name to get the plain value back,
`string(user)`. A literal never carries a tag, so the call is the only way to make one:

```ant
type UserId = string
let id: UserId = "u-1" // error: cannot use string where UserId is expected
```

Arithmetic works on two values of one named type and keeps the type. A plain number on the other
side is an error, because `2.0` is a `float64`, not a `Meters`:

```ant
type Meters = float64
let side = Meters(3.0)
side * Meters(2.0)  // Meters(6.0)
side < Meters(4.0)  // true
float64(side) * 2.0 // 6.0
```

```ant
type Meters = float64
let side = Meters(3.0)
let area = side * 2.0 // error: * needs both operands of one type, got Meters and float64
```

Library functions take the base type, so unwrap before calling them. The conversions accept the
tagged value directly, and so does `len` when the base is `string`:

```ant
type UserId = string
let id = UserId("u-1")
string(id).toUpper() // "U-1"
len(id)              // 3
```

```ant
type UserId = string
let id = UserId("u-1")
let shout = id.toUpper() // error: UserId has no field or function toUpper
```

Equality and map keys work as for the base, so a `Map[UserId, Account]` is the natural way to index
by id. Use a named primitive type for identifiers and units, where mixing two values of the same
base is a mistake; use a [record](records.md) when the value has more than one part.

## Assignability

A value can stand where a type is expected when the types are the same, or the expected type is
wider in one of a few ways:

- a union accepts each of its members, so a `string` goes where `string | null` is expected;
- an anonymous record type accepts any record that has at least its fields, with the same types;
- an interface accepts any type that has its functions; see [Interfaces](interfaces.md);
- `never` goes anywhere, which is why `panic(...)` fits in any branch.

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
type User = { name: string, age: int }

func area(s: Shape): float64 {
  if s is Circle { Math.PI * s.radius * s.radius } else { s.width * s.height }
}

func nameOf(p: { name: string }): string {
  p.name
}

println(area(Circle{ radius: 1.0 }))
println(nameOf(User{ name: "Ann", age: 31 }))
```

```text
3.141592653589793
Ann
```

Nothing widens on its own. An `int8` is not an `int`, and a named type is not its base. Containers
do not widen either: an `Array[Circle]` is not an `Array[Shape]`, because code holding it as an
`Array[Shape]` could push a `Rect` into an array of circles. Build a new array when you need the
wider type:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = circles // error: cannot use Array[Circle] where Array[Shape] is expected
```

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = [...circles]
shapes.push(Rect{ width: 2.0, height: 3.0 })
println(shapes.length())
println(circles.length())
```

```text
2
1
```

Ant has no `any`. A function that takes every type is [generic](generics.md), and data whose shape
you learn at run time is a union you narrow with `is`, such as `Json.Value`.

## Inference

The type checker infers types locally, from the expression in front of it and the type expected
there. Annotate a function's parameters; the rest usually follows:

```ant
let names = ["Ann", "Bob"]
names.map({ x.length() }) // [3, 3]
```

The checker knows `names` is an `Array[string]` from its elements, so `x` in the lambda is a
`string`, and the result is an `Array[int]`.

An annotation gives an expression its expected type, which is how an empty literal gets an element
type. Without one, the checker has nothing to go on:

```ant
let tags: Array[string] = []
tags.push("new")
println(tags)
```

```text
["new"]
```

```ant
let tags = [] // error: cannot infer the element type of an empty array; annotate it
```

Nothing is inferred across statements. A `var` has exactly the type of its first value, so annotate
it when later assignments hold other members of a union:

```ant
var state = "idle"
state = null // error: cannot use null where string is expected
```

```ant
var state: string | null = "idle"
state = null
println(state ?? "none")
```

```text
none
```

A function's return type is inferred from its body, unless the function is exported or recursive. An
exported function states it so that the module's interface reads without the bodies. A recursive one
needs it because the checker meets the call before it knows the result:

```ant
func double(n: int) { n * 2 }

func fact(n: int) {
  if n <= 1 { 1 } else { n * fact(n - 1) } // error: fact is recursive; state its return type
}
```

```ant
func fact(n: int): int {
  if n <= 1 { 1 } else { n * fact(n - 1) }
}

println(fact(5))
```

```text
120
```

## Mistakes to avoid

Coming from JavaScript or Python, these catch people first.

- **Adding a number to a string.** `+` takes two strings or two numbers of one kind. Interpolate
  instead:

  ```ant
  let count = 3
  let label = "items: " + count // error: + needs both operands of one type, got string and int
  ```

  ```ant
  let count = 3
  "items: ${count}" // "items: 3"
  ```

- **Truthiness.** A condition must be a `bool`. Ask the question you mean:

  ```ant
  let items = [1, 2]
  if items.length() { println("some") } // error: cannot use int where bool is expected
  ```

  ```ant
  let items = [1, 2]
  !items.isEmpty() // true
  ```

- **Comparing an `int` with a `float64`.** `1 == 1.0` is a compile error, not `true` or `false`.
  Convert one side.

- **Expecting a float from integer division.** `7 / 2` is `3`. Convert an operand when you want
  `3.5`: `float64(7) / 2.0`.

## See also

- [ant check](../tools/check.md) and [ant run](../tools/run.md)
- [int](../language/types/int.md), [float](../language/types/float.md),
  [byte](../language/types/byte.md), [string](../language/types/string.md),
  [bool](../language/types/bool.md), [never](../language/types/never.md)
- [Arithmetic operators](../language/operators/arithmetic.md) and
  [equality](../language/operators/equality.md)
- [type](../language/keywords/type.md), for declaring types
- [Null](null.md), for nullable types and narrowing
- [Unions](unions.md), [Interfaces](interfaces.md), and [Generics](generics.md)
- [Runtime types](runtime-types.md), for `type(T)` and what types mean at run time
