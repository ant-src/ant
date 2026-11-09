<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int16

```ant signature
func int16[T: Number](x: T): int16
```

int16 converts x to int16 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `int16` holds the whole numbers from `-32768` to `32767`.
- An integer literal is always an `int`; convert it with `int16(5)`.
- Arithmetic on two `int16` values gives an `int16`, and panics when the result leaves the range.

## Basic examples

Convert an `int` or a whole `float64`:

```ant
int16(1200)   // 1200
int16(-7.0)   // -7
int16(32767)  // 32767
int16(-32768) // -32768
```

Two `int16` values add to an `int16`:

```ant
let a = int16(1000)
let b = int16(234)
a + b // 1234
```

## Scope

Every numeric type converts, the sized ones included:

```ant
int16(byte(200))     // 200
int16(uint32(65))    // 65
int16(float32(-3.0)) // -3
int16(int8(-128))    // -128
```

`int16` is a function value, so it converts every element of an array:

```ant
Array.map([10, -20, 30], int16) // [10, -20, 30]
```

It unwraps a named primitive type over a number:

```ant
type Celsius = int16
let freezing = Celsius(int16(0))
freezing        // Celsius(0)
int16(freezing) // 0
```

In a pipe, the value is the argument:

```ant
(String.toInt("-512") ?? 0) | int16() // -512
```

## Applications

Store an audio sample in 16 bits. Scale a level in `-1.0` to `1.0`, round it, and convert:

```ant
func toSample(level: float64): int16 {
  int16(Math.round(Math.clamp(level, -1.0, 1.0) * 32767.0))
}

[0.0, 0.5, -1.0, 1.5].map(toSample) // [0, 16384, -32767, 32767]
```

Widen to `int` before a sum that may pass `32767`:

```ant
let samples = [int16(30000), int16(20000), int16(-1000)]
samples.map(int).sum() // 49000
```

## Properties and relations

Every `int` in range survives the round trip through `int16`:

```ant
for n in [-32768, -1, 0, 1, 32767] {
  assert(int(int16(n)) == n)
}
```

Converting a float with a fraction needs an explicit rounding first; `Math.trunc` rounds toward
zero:

```ant
assert(int16(Math.trunc(2.7)) == int16(2))
assert(int16(Math.trunc(-2.7)) == int16(-2))
assert(int16(Math.round(2.5)) == int16(3))
```

`Bits.wrapAdd` wraps around where `+` panics:

```ant
assert(Bits.wrapAdd(int16(30000), int16(30000)) == int16(-5536))
```

## Possible issues

A value outside the range panics; it does not wrap:

```ant
int16(32768)  // panic: 32768 out of range for int16
int16(-32769) // panic: -32769 out of range for int16
```

A fraction panics instead of truncating:

```ant
int16(2.5) // panic: 2.5 has a fraction, so it does not convert to int16
```

`NaN` and the infinities panic:

```ant
int16(Math.NaN)      // panic: NaN out of range for int16
int16(Math.Infinity) // panic: Infinity out of range for int16
```

An integer literal is an `int`, not an `int16`:

```ant
let level: int16 = 5 // error: cannot use int where int16 is expected
```

Convert the literal:

```ant
let level: int16 = int16(5)
level // 5
```

An `int16` and an `int` do not mix in arithmetic:

```ant
let a = int16(1)
println(a + 1) // error: + needs both operands of one type, got int16 and int
```

The same holds for a comparison with a literal:

```ant
let a = int16(1)
println(a > 0) // error: > needs both operands of one type, got int16 and int
```

Convert the literal instead:

```ant
let a = int16(1)
a > int16(0) // true
```

Arithmetic on `int16` panics when the result leaves the range, even when the final value would fit
an `int`:

```ant
int16(30000) + int16(30000) // panic: integer overflow
int16(-32768) / int16(-1)   // panic: integer overflow
```

A sum of `int16` values is an `int16` too, so it overflows the same way:

```ant
[int16(30000), int16(20000)].sum() // panic: Array.sum: integer overflow
```

Widen first with `.map(int).sum()`, as in Applications.

## See also

[int](int.md) · [int8](int8.md) · [int32](int32.md) · [int64](int64.md) · [uint16](uint16.md) ·
[float64](float64.md) · [Math.trunc](../Math/trunc.md) · [Math.round](../Math/round.md) ·
[Math.clamp](../Math/clamp.md) · [Bits.wrapAdd](../Bits/wrapAdd.md) ·
[Arithmetic operators](../../language/operators/arithmetic.md)
