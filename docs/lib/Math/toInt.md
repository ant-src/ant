<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.toInt

```ant signature
func Math.toInt(x: float64): int | null
```

Math.toInt returns x as an int, or null if x has a fraction, is out of range, or is NaN or an
infinity. It is int(x) without the panic.

<!-- end generated -->

## Details

- `Math.toInt` returns `null` exactly where `int(x)` panics: on a fraction, on a value outside the
  `int` range, on `NaN`, and on an infinity. It never panics.
- It never rounds. Round first with [Math.round](round.md), [Math.floor](floor.md),
  [Math.ceil](ceil.md), or [Math.trunc](trunc.md) when a fraction should become an `int`.

## Basic examples

```ant
Math.toInt(2.0)  // 2
Math.toInt(-3.0) // -3
Math.toInt(2.5)  // null
```

Give a default with `??`:

```ant
Math.toInt(7.0) ?? 0 // 7
Math.toInt(7.5) ?? 0 // 0
```

## Scope

`NaN` and the infinities are `null`:

```ant
Math.toInt(Math.NaN)       // null
Math.toInt(Math.Infinity)  // null
Math.toInt(-Math.Infinity) // null
```

A fraction of any size is `null`, however close to a whole number:

```ant
Math.toInt(1e-300)    // null
Math.toInt(0.9999999) // null
```

`-0.0` is whole, and gives `0`:

```ant
Math.toInt(-0.0) // 0
```

A whole number outside the `int` range is `null`. `-2^63` is in range; `2^63` is not:

```ant
Math.toInt(1e18)                   // 1000000000000000000
Math.toInt(1e20)                   // null
Math.toInt(-9223372036854775808.0) // -9223372036854775808
Math.toInt(9223372036854775808.0)  // null
```

The parameter is named `x`, and in a pipe the value is the argument:

```ant
Math.toInt(x: 4.0) // 4
4.0 | Math.toInt() // 4
```

`Math.toInt` is a function value, so it can be passed to `Array.map`:

```ant
[1.0, 1.5, 2.0].map(Math.toInt) // [1, null, 2]
```

## Applications

Read an integer from JSON, whose numbers are all `float64`:

```ant
func readCount(text: string): int | null {
  let parsed = Json.parse(text)
  if parsed is Err {
    return null
  }
  let value = parsed.value
  if value is float64 {
    return Math.toInt(value)
  }
  null
}

readCount("8080")   // 8080
readCount("80.5")   // null
readCount("\"80\"") // null
```

Find the side of a square from a small tile count, or `null` when the tiles do not form a square:

```ant
func squareSide(tiles: int): int | null {
  Math.toInt(Math.sqrt(float64(tiles)))
}

squareSide(49) // 7
squareSide(50) // null
```

Keep only the whole numbers of a list, as `int`s:

```ant
let readings = [3.0, 2.5, 7.0, Math.NaN, 10.0]
let whole: Array[int] = readings.flatMap({
  let n = Math.toInt(x)
  if n == null { [] } else { [n] }
})
whole // [3, 7, 10]
```

Round first when a fraction should still give an `int`:

```ant
Math.toInt(Math.round(2.5))  // 3
Math.toInt(Math.floor(-2.5)) // -3
Math.toInt(Math.trunc(-2.5)) // -2
```

## Properties and relations

Where `int(x)` succeeds, `Math.toInt(x)` gives the same value:

```ant
for x in [0.0, -0.0, 42.0, -7.0, 1e15] {
  assert(Math.toInt(x) == int(x))
}
```

An `int` up to 2^53 in size goes to `float64` and back unchanged:

```ant
for n in [0, 1, -1, 123456789, 9007199254740992] {
  assert(Math.toInt(float64(n)) == n)
}
```

A result that is not `null` is whole, so rounding toward zero does not change it:

```ant
for x in [2.0, 2.5, -3.0, Math.NaN] {
  if Math.toInt(x) != null {
    assert(Math.trunc(x) == x)
  }
}
```

## Possible issues

The result may be `null`, so it does not go where an `int` is expected:

```ant
let n = Math.toInt(2.0)
let next = n + 1 // error: + needs both operands of one type, got int | null and int
```

The argument must be a `float64`. An `int` is a compile error:

```ant
let n = Math.toInt(2) // error: cannot use int where float64 is expected
```

Arithmetic that looks whole may leave a fraction in binary, so the result is `null`:

```ant
(0.1 + 0.2) * 10.0                         // 3.0000000000000004
Math.toInt((0.1 + 0.2) * 10.0)             // null
Math.toInt(Math.round((0.1 + 0.2) * 10.0)) // 3
```

`Math.maxInt` does not fit a `float64` exactly. It rounds up to 2^63, which is out of range:

```ant
Math.toInt(float64(Math.maxInt)) // null
```

Above 2^53 a `float64` cannot hold every integer, so the `int` may differ from the number written:

```ant
Math.toInt(9007199254740993.0) // 9007199254740992
```

`int(x)` panics where `Math.toInt` returns `null`:

```ant
int(2.5)        // panic: 2.5 has a fraction, so it does not convert to int
Math.toInt(2.5) // null
```

## See also

[Math.trunc](trunc.md) · [Math.round](round.md) · [Math.floor](floor.md) · [Math.ceil](ceil.md) ·
[Math.isFinite](isFinite.md) · [String.toInt](../String/toInt.md) · [int](../prelude/int.md)
