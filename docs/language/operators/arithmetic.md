<!-- status: reviewed -->

# Arithmetic operators

`a + b`, `a - b`, `a * b`, `a / b`, and `a % b` add, subtract, multiply, divide, and take the
remainder. `-a` negates a number. `+` also joins two strings.

## Details

- Both operands have one type: a single numeric kind such as `int` or `float64`, or a named
  primitive type over one. Ant never converts between kinds, so `1 + 1.0` and `"n=" + 1` are compile
  errors. The result has the operands' type.
- Integer `/` truncates toward zero, and `%` takes the sign of the dividend, for floats too.
- Integer arithmetic panics when the result does not fit the kind, including unsigned subtraction
  below zero, and `/` and `%` panic on a zero divisor. Float arithmetic follows IEEE 754 and never
  panics: dividing by zero gives an infinity or `NaN`.
- `*`, `/`, and `%` bind tighter than `+` and `-`; unary `-` binds tighter than both.

## Basic examples

The five binary operators on integers:

```ant
7 + 2 // 9
7 - 2 // 5
7 * 2 // 14
7 / 2 // 3
7 % 2 // 1
```

On floats, `/` keeps the fraction:

```ant
7.0 / 2.0 // 3.5
1.5 * 4.0 // 6.0
```

`+` joins two strings:

```ant
"ant" + "hill" // "anthill"
```

## Scope

### Negative operands

Integer `/` truncates toward zero, and `%` keeps the sign of the dividend:

```ant
-7 / 2 // -3
-7 % 2 // -1
7 % -2 // 1
```

`%` on floats follows the same rule:

```ant
5.5 % 2.0  // 1.5
-7.5 % 2.0 // -1.5
```

### Unary minus

`-` before an operand negates it, and binds tighter than `*`:

```ant
let n = 4
-n * 3   // -12
-(n - 6) // 2
```

### Precedence and grouping

`*`, `/`, and `%` bind tighter than `+` and `-`. Parentheses group first:

```ant
2 + 3 * 4   // 14
(2 + 3) * 4 // 20
```

### Other numeric kinds

The operators work on every numeric kind, as long as both operands have it. Convert a literal to the
kind of the other operand:

```ant
let level = byte(200)
level + byte(50)            // 250
uint8(255) / uint8(2)       // 127
float32(1.5) * float32(2.0) // 3.0
```

### Float infinities and `NaN`

Float division by zero follows IEEE 754 and does not panic:

```ant
1.0 / 0.0    // Infinity
-1.0 / 0.0   // -Infinity
0.0 / 0.0    // NaN
1.0 % 0.0    // NaN
1e308 * 10.0 // Infinity
```

### Named primitive types

Two values of one named primitive type combine as their base kind, and the result keeps the type:

```ant
type Cents = int
type Name = string
Cents(150) + Cents(25)     // Cents(175)
Cents(150) * Cents(2)      // Cents(300)
Name("Ann") + Name(" Lee") // Name("Ann Lee")
```

### Compound assignment

Each compound assignment reads the target, applies the operator, and stores the result:

```ant
var n = 10
n += 5
n *= 2
n -= 1
n /= 3
n %= 4
println(n)
```

```text
1
```

### Generic functions

Inside a function with `[T: Number]`, the operators apply to `T`, because they apply to every member
of `Number`:

```ant
func double[T: Number](x: T): T { x + x }

println(double(21))
println(double(1.25))
println(double(byte(100)))
```

```text
42
2.5
200
```

## Applications

### Average of integers

Convert to `float64` before dividing to keep the fraction:

```ant
let scores = [90, 85, 72]
let total = scores.sum() ?? 0
println(total / scores.length())
println(float64(total) / float64(scores.length()))
```

```text
82
82.33333333333333
```

### Hours and minutes

`/` and `%` by the same divisor split a quantity into units:

```ant
func clock(minutes: int): string {
  let h = minutes / 60
  let m = minutes % 60
  let mm = String.fromInt(m).padStart(2, "0")
  "${h}:${mm}"
}

println(clock(135))
println(clock(61))
```

```text
2:15
1:01
```

### Index that wraps around

`%` takes the sign of the dividend, so step back from `0` by adding the length first:

```ant
let days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
func shift(i: int, by: int): int {
  let n = days.length()
  ((i + by) % n + n) % n
}
println(days[shift(5, 3)])
println(days[shift(0, -1)])
```

```text
Tue
Sun
```

### Money in cents

A named type keeps cents from mixing with other integers, and integer arithmetic stays exact.
`Array.sum` takes only the members of `Number`, so add the `Cents` values with `+=`:

```ant
type Cents = int

func total(prices: Array[Cents]): Cents {
  var sum = Cents(0)
  for p in prices {
    sum += p
  }
  sum
}

println(total([Cents(199), Cents(250)]))
```

```text
Cents(449)
```

## Properties and relations

For integers with a nonzero divisor, `/` and `%` fit together:

```ant
for a in [7, -7, 13, -13] {
  for b in [2, -2, 5, -5] {
    assert(a / b * b + a % b == a)
  }
}
```

`+` and `*` are commutative on integers, and `-a` is `0 - a`:

```ant
let a = 12
let b = -5
assert(a + b == b + a)
assert(a * b == b * a)
assert(-a == 0 - a)
```

Wrapping arithmetic is a function, not an operator:

```ant
Bits.wrapAdd(Math.maxInt, 1) == Math.minInt // true
```

## Possible issues

### No mixing of `int` and `float64`

An integer and a float do not combine. Convert one of them:

```ant
let half = 1 + 0.5 // error: + needs both operands of one type, got int and float64
```

```ant
float64(1) + 0.5 // 1.5
```

### `+` does not convert to a string

Use interpolation to build a string from a number:

```ant
let count = 3
let label = "n=" + count // error: + needs both operands of one type, got string and int
```

```ant
let count = 3
"n=${count}" // "n=3"
```

### Only `+` works on strings

`+` joins strings, but `*` and `-` do not apply to them:

```ant
let line = "-" * 3 // error: * needs both operands of one type, got string and int
```

Use the library function for the job:

```ant
"-".repeat(3) // "---"
```

### Integer division drops the fraction

`1 / 2` is `0`, not `0.5`. Divide floats when you need the fraction:

```ant
1 / 2     // 0
1.0 / 2.0 // 0.5
```

### Integer overflow panics

There is no silent wraparound. A result that does not fit the kind panics:

```ant
Math.maxInt + 1     // panic: integer overflow
-Math.minInt        // panic: integer overflow
int8(100) * int8(2) // panic: integer overflow
```

Unsigned subtraction below zero overflows too:

```ant
byte(0) - byte(1) // panic: integer overflow
```

Compound assignment overflows the same way:

```ant
var n = Math.maxInt
n += 1 // panic: integer overflow
```

Use `Bits.wrapAdd`, `Bits.wrapSub`, or `Bits.wrapMul` when you want wraparound.

### Integer division by zero panics

```ant
10 / 0 // panic: division by zero
10 % 0 // panic: division by zero
```

Guard the divisor before dividing:

```ant
func perItem(total: int, count: int): int {
  if count == 0 { return 0 }
  total / count
}

println(perItem(10, 0))
println(perItem(10, 4))
```

```text
0
2
```

### Literals are always `int` or `float64`

A literal does not take the kind of the other operand, so convert it:

```ant
let level = byte(200)
let next = level + 1 // error: + needs both operands of one type, got byte and int
```

Write `level + byte(1)`.

### A named type does not mix with its base

`Cents(5) + 3` is a compile error. Wrap the plain value:

```ant
type Cents = int
let price = Cents(5) + 3 // error: + needs both operands of one type, got Cents and int
```

Write `Cents(5) + Cents(3)`.

### Narrow nullable and union values first

A value that may be `null`, or a union such as `Number`, is not a single kind:

```ant
let stock: int | null = 3
let next = stock + 1 // error: + needs both operands of one type, got int | null and int
```

Use `??` or a guard: `(stock ?? 0) + 1`.

```ant
func add(a: Number, b: Number): Number {
  a + b // error: + applies to a single numeric kind, string, or a named primitive type over one, not to Number; narrow first
}
```

Use a type parameter, `func add[T: Number](a: T, b: T): T`, as in the generic example above.

### Float results are rounded

Floats are binary, so decimal fractions are not exact:

```ant
0.1 + 0.2        // 0.30000000000000004
0.1 + 0.2 == 0.3 // false
```

Count money in integer cents, or compare floats within a tolerance.

### A negative remainder

`%` is not a modulo: a negative dividend gives a negative remainder, which is not a valid index. Add
the divisor and take `%` again, as in the wrap-around example above:

```ant
-1 % 7           // -1
(-1 % 7 + 7) % 7 // 6
```

## See also

[Comparison operators](comparison.md) · [Assignment](assignment.md) · [int](../types/int.md) ·
[float64 and float32](../types/float.md) · [byte](../types/byte.md) ·
[Number literals](../literals/numbers.md) · [Bits.wrapAdd](../../lib/Bits/wrapAdd.md) ·
[Array.sum](../../lib/Array/sum.md) · [String.repeat](../../lib/String/repeat.md) ·
[String.padStart](../../lib/String/padStart.md) · [Math.abs](../../lib/Math/abs.md) ·
[float64(x)](../../lib/prelude/float64.md)
