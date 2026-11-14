<!-- status: reviewed -->

# Numbers

Ant has whole numbers and floating-point numbers of several sizes, and it never mixes two numeric
types on its own. This program splits a bill.

```ant
let total = 4_250
let people = 3

// Integer division drops the fraction; % gives what is left over.
println("each pays ${total / people} cents")
println("${total % people} cents left over")

// Convert to float64 to keep the fraction.
let exact = float64(total) / float64(people)
println("exactly ${exact} cents")

// Convert back to int: Math.round first, since the result must be whole.
let rounded = int(Math.round(exact))
println("rounded: ${rounded} cents")

// Sized types, such as byte, hold smaller ranges.
let tip = byte(0x0F)
println("tip: ${tip + byte(5)} percent")
```

```text
each pays 1416 cents
2 cents left over
exactly 1416.6666666666667 cents
rounded: 1417 cents
tip: 20 percent
```

`total` is an `int`, and `exact` is a `float64`. An integer literal is always an `int`, the 64-bit
default, and a literal with a point or an exponent, such as `2.5` or `1e9`, is always a `float64`.
Underscores group digits, and `0x`, `0o`, and `0b` write hexadecimal, octal, and binary. A number
followed by a time unit, such as `500ms` or `2h30m`, is not a number but a
[duration literal](../language/literals/durations.md) of type `Time.Duration`.

Ant has ten numeric types: `int8` to `int64`, `uint8` to `uint64`, `float32`, and `float64`; `int`
is `int64` and `byte` is `uint8`. To get one, call its name as a function, such as `float64(total)`
or `byte(0x0F)`. Change the program and try a few things.

Mix two numeric types, and `ant check` rejects the program before it runs:

```ant
let total = 4_250
let rate = 0.15
let tax = total * rate
// error: * needs both operands of one type, got int and float64
let tip = byte(15)
let more = tip + 5 // error: + needs both operands of one type, got byte and int
```

Write `float64(total) * rate` instead. A literal is always an `int`, so write `byte(5)`. Numbers
that leave their type's range do not wrap around silently, as they do in many languages. The program
panics:

```ant
let tip = byte(200)
println(tip + byte(100)) // panic: integer overflow
```

A conversion to an integer type panics too when the value has a fraction or does not fit, so
`int(2.5)` and `byte(300)` both stop the program. Round with `Math.round`, truncate with
`int(Math.trunc(x))`, or use `Math.toInt(x)`, which gives `null` instead of panicking. Integer
division by zero panics as well, while float division by zero follows IEEE 754 and gives an
infinity, or NaN for `0.0 / 0.0`.

When you want wrapping, ask for it with [Bits.wrapAdd](../lib/Bits/wrapAdd.md) and its siblings.
[Math](../lib/Math/index.md) has the rounding functions, `Math.sqrt`, `Math.pow`, and constants such
as `Math.PI` and `Math.maxInt`; [Bits](../lib/Bits/index.md) has the bitwise operations, which Ant
writes as functions rather than operators.

Read more in the [types chapter](../guide/types.md) of the guide, and in the reference for
[int](../language/types/int.md), [float](../language/types/float.md),
[number literals](../language/literals/numbers.md), and
[arithmetic operators](../language/operators/arithmetic.md).
