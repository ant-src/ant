<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromFloat

```ant signature
func String.fromFloat(x: float64, digits: int): string
```

String.fromFloat returns x in fixed point with digits decimals, rounded half to even. NaN and the
infinities render as [`String.from`](from.md) does. It panics if digits \< 0.

<!-- end generated -->

## Details

- Writes `x` in fixed point, never with an exponent, with exactly `digits` decimals. With `digits` 0
  there is no `.`.
- Rounds half to even, and rounds the binary value of `x`, not the decimal it was written as.
- A result that is zero has no sign.
- `NaN`, `Infinity`, and `-Infinity` render as `String.from` renders them.
- Panics if `digits` is negative or above 2^31 − 1.

## Basic examples

Write a number with two decimals:

```ant
String.fromFloat(3.14159, 2) // "3.14"
String.fromFloat(42.0, 2)    // "42.00"
```

With `digits` 0 the result is a whole number, with no `.`:

```ant
String.fromFloat(3.14159, 0) // "3"
```

Put it in a string with interpolation:

```ant
let total = 19.5
"Total: ${String.fromFloat(total, 2)}" // "Total: 19.50"
```

## Scope

A tie rounds to the even digit:

```ant
String.fromFloat(0.125, 2) // "0.12"
String.fromFloat(0.375, 2) // "0.38"
String.fromFloat(2.5, 0)   // "2"
String.fromFloat(3.5, 0)   // "4"
String.fromFloat(-2.5, 0)  // "-2"
```

A negative number keeps its `-`, but a result that rounds to zero has none, and neither does
negative zero:

```ant
String.fromFloat(-1.25, 1)  // "-1.2"
String.fromFloat(-0.001, 2) // "0.00"
String.fromFloat(-0.0, 1)   // "0.0"
```

A large number is written out in full, where `String.from` uses an exponent:

```ant
String.fromFloat(1e21, 2) // "1000000000000000000000.00"
String.from(1e21)         // "1e+21"
```

A tiny number becomes zero at the given precision:

```ant
String.fromFloat(1e-10, 3)  // "0.000"
String.fromFloat(5e-324, 2) // "0.00"
```

More digits show more of the binary value. The `float64` nearest `0.1` is a little above it:

```ant
String.fromFloat(0.1, 1)  // "0.1"
String.fromFloat(0.1, 20) // "0.10000000000000000555"
```

`NaN` and the infinities render the same at every valid `digits`:

```ant
String.fromFloat(Math.NaN, 2)       // "NaN"
String.fromFloat(Math.Infinity, 2)  // "Infinity"
String.fromFloat(-Math.Infinity, 0) // "-Infinity"
```

Call it with named arguments, or in a pipe:

```ant
String.fromFloat(x: 1.5, digits: 3) // "1.500"
1.0 / 3.0 | String.fromFloat(4)     // "0.3333"
```

## Applications

Format prices to cents:

```ant
let prices = [4.5, 12.0, 0.99]
let total = Array.sum(prices) ?? 0.0
for p in prices {
  println(String.fromFloat(p, 2))
}
println("total ${String.fromFloat(total, 2)}")
```

```text
4.50
12.00
0.99
total 17.49
```

Show a ratio of two counts as a percentage. Convert the counts to `float64` first:

```ant
func percent(done: int, all: int): string {
  "${String.fromFloat(100.0 * float64(done) / float64(all), 1)}%"
}

println(percent(1, 3))
println(percent(7, 8))
```

```text
33.3%
87.5%
```

Right-align a column of numbers with [`String.padStart`](padStart.md):

```ant
type Row = { name: string, score: float64 }

let rows = [Row{ name: "ann", score: 9.25 }, Row{ name: "bob", score: 10.0 }, Row{ name: "cy", score: 7.125 }]
for row in rows {
  println("${String.padEnd(row.name, 4)}${String.padStart(String.fromFloat(row.score, 2), 6)}")
}
```

```text
ann   9.25
bob  10.00
cy    7.12
```

Format a whole array with [`Array.map`](../Array/map.md) and join it:

```ant
let temps = [21.04, 19.96, -3.5]
Array.map(temps, { String.fromFloat(x, 1) }) | String.join(", ") // "21.0, 20.0, -3.5"
```

## Properties and relations

A finite result has exactly `digits` decimals after the `.`:

```ant
for digits in [1, 2, 5] {
  let c = String.cut(String.fromFloat(Math.PI, digits), ".")
  assert(c != null)
  assert(String.length(c.after) == digits)
}
assert(!String.contains(String.fromFloat(Math.PI, 0), "."))
```

For `NaN` and the infinities, the result is what `String.from` gives:

```ant
for x in [Math.NaN, Math.Infinity, -Math.Infinity] {
  assert(String.fromFloat(x, 3) == String.from(x))
}
```

A whole number with `digits` 0 is the same text `String.fromInt` writes:

```ant
for x in [0.0, 7.0, -12.0, 1e15] {
  assert(String.fromFloat(x, 0) == String.fromInt(int(x)))
}
```

`String.toFloat` reads the result back:

```ant
String.toFloat(String.fromFloat(2.5, 2)) // 2.5
```

## Possible issues

The rounding works on the binary value. The `float64` nearest `1.005` is a little below it, and the
one nearest `2.675` too, so both round down:

```ant
String.fromFloat(1.005, 2) // "1.00"
String.fromFloat(2.675, 2) // "2.67"
```

A tie rounds to even here, but [`Math.round`](../Math/round.md) rounds a half away from zero:

```ant
String.fromFloat(2.5, 0) // "2"
Math.round(2.5)          // 3.0
```

A negative count of digits panics:

```ant
String.fromFloat(1.5, -1) // panic: String.fromFloat needs a count of digits of at least 0, got -1
```

The range check applies to `NaN` and the infinities too:

```ant
String.fromFloat(Math.NaN, -1) // panic: String.fromFloat needs a count of digits of at least 0, got -1
```

A count above 2^31 − 1 panics as well:

```ant
String.fromFloat(1.5, 2147483648) // panic: String.fromFloat needs a count of digits of at most 2147483647, got 2147483648
```

An `int` is not a `float64`, so passing one is a compile error:

```ant
let n = 3
println(String.fromFloat(n, 2)) // error: cannot use int where float64 is expected
```

Convert it with `float64(n)`, or write the literal as `3.0`.

A number has no namespace, so `x.fromFloat(2)` is a compile error:

```ant
let price = 1.25
println(price.fromFloat(2)) // error: float64 has no field or function fromFloat
```

Use a pipe instead:

```ant
let price = 1.25
price | String.fromFloat(2) // "1.25"
```

## See also

[String.from](from.md) · [String.toFloat](toFloat.md) · [String.fromInt](fromInt.md) ·
[String.padStart](padStart.md) · [Math.round](../Math/round.md) ·
[pipe](../../language/operators/pipe.md)
