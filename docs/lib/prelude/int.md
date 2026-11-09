<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int

```ant signature
func int[T: Number](x: T): int
```

int converts x to int, which is int64, exactly. It panics if x is out of range, has a fraction, or
is NaN or infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- To convert a float with a fraction, round it first with [Math.trunc](../Math/trunc.md),
  [Math.round](../Math/round.md), or [Math.floor](../Math/floor.md). To get `null` instead of a
  panic, use [Math.toInt](../Math/toInt.md).
- `int` converts numbers only. To read an `int` from text, use [String.toInt](../String/toInt.md).

## Basic examples

Convert a smaller integer kind to `int`:

```ant
int(byte(200)) // 200
int(int8(-5))  // -5
```

Convert a float with no fraction:

```ant
int(3.0)  // 3
int(-0.0) // 0
```

Truncate a float toward zero first, then convert:

```ant
int(Math.trunc(2.7))  // 2
int(Math.trunc(-2.7)) // -2
```

## Scope

Every numeric kind converts, as long as the value fits:

```ant
int(uint64(9223372036854775807)) // 9223372036854775807
int(float32(2.5) * float32(2.0)) // 5
int(1e18)                        // 1000000000000000000
int(7)                           // 7
```

A named primitive over a number is unwrapped:

```ant
type Count = int
type Celsius = float64
int(Count(4)) + 1  // 5
int(Celsius(21.0)) // 21
```

Pass `int` itself as a function:

```ant
Array.map([1.0, 2.0, 3.0], int) // [1, 2, 3]
```

## Applications

Add bytes without overflowing `byte` by summing them as `int`:

```ant
let data = [byte(200), byte(100), byte(50)]
Array.sum(Array.map(data, int)) // 350
```

Turn a price in dollars into whole cents, rounding first:

```ant
func cents(price: float64): int {
  int(Math.round(price * 100.0))
}

cents(19.99) // 1999
cents(0.1)   // 10
```

Split a duration in seconds into minutes and seconds:

```ant
let elapsed = 754.6
let total = int(Math.floor(elapsed))
"${total / 60}m ${total % 60}s" // "12m 34s"
```

## Properties and relations

An `int` survives a round trip through `float64` when `float64` represents it exactly:

```ant
for n in [0, 1, -42, 4503599627370496] {
  assert(int(float64(n)) == n)
}
```

`int(x)` agrees with [Math.toInt](../Math/toInt.md) wherever it does not panic:

```ant
for x in [3.0, -8.0, 1e15] {
  assert(Math.toInt(x) == int(x))
}
Math.toInt(2.5) // null
```

## Possible issues

A float with a fraction panics instead of rounding:

```ant
int(2.5) // panic: 2.5 has a fraction, so it does not convert to int
```

Float arithmetic often leaves a fraction you do not see in the source:

```ant
int(0.1 + 0.2) // panic: 0.30000000000000004 has a fraction, so it does not convert to int
```

Values out of range panic, and so do `NaN` and the infinities:

```ant
int(Math.maxUint64) // panic: 18446744073709551615 out of range for int
int(9.3e18)         // panic: 9300000000000000000.0 out of range for int
int(Math.NaN)       // panic: NaN out of range for int
int(Math.Infinity)  // panic: Infinity out of range for int
```

`int` does not parse text; that is a compile error:

```ant
let n = int("3") // error: int converts a number or a named primitive type over a number, got string
```

A nullable number must be narrowed first:

```ant
let found: int | null = Array.indexOf([5, 6], 6)
let n = int(found)
// error: int converts a number or a named primitive type over a number, got int | null
```

## See also

[int64](int64.md), [float64](float64.md), [byte](byte.md), [Math.toInt](../Math/toInt.md),
[Math.trunc](../Math/trunc.md), [Math.round](../Math/round.md), [String.toInt](../String/toInt.md),
[String.from](../String/from.md)
