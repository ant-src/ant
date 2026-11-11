<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# float32

```ant signature
func float32[T: Number](x: T): float32
```

float32 converts x to float32, rounding to the nearest value it holds; it never panics. x may also
be a named primitive over a number.

<!-- end generated -->

## Details

- Ant has no `float32` literal. Every float literal is a `float64`; convert it with `float32`.
- A value too large for `float32` rounds to an infinity, and one too small rounds to `0.0` or
  `-0.0`, keeping its sign.
- A `float32` does not mix with a `float64` in arithmetic, comparison, or `==`, and the `Math`
  functions over `float64` do not take it. Convert one side.

## Basic examples

```ant
float32(1.5) // 1.5
float32(3)   // 3.0
float32(0.1) // 0.1
```

A `float32` renders with the fewest digits that read back as the same `float32`. Converted to
`float64`, the same value shows all the digits that `float64` keeps:

```ant
float64(float32(0.1)) // 0.10000000149011612
```

## Scope

Every numeric kind converts, signed or unsigned, of any size:

```ant
float32(int8(-7))       // -7.0
float32(byte(255))      // 255.0
float32(Math.maxUint64) // 18446744000000000000.0
float32(float64(2.25))  // 2.25
float32(float32(2.25))  // 2.25
```

An integer with more significant digits than `float32` holds rounds to the nearest value it holds:

```ant
float32(16777216)  // 16777216.0
float32(16777217)  // 16777216.0
float32(123456789) // 123456790.0
```

Values beyond the range round to an infinity, and values too small to hold round to zero, keeping
the sign:

```ant
float32(1e300)  // Infinity
float32(-1e300) // -Infinity
float32(1e-50)  // 0.0
float32(-1e-50) // -0.0
```

`NaN` and the infinities pass through:

```ant
float32(Math.NaN)       // NaN
float32(-Math.Infinity) // -Infinity
```

A named primitive over a number is unwrapped and converted, whatever its base kind:

```ant
type Meters = float64
type Count = int
float32(Meters(2.5)) // 2.5
float32(Count(4))    // 4.0
```

To build a named primitive over `float32`, convert the literal first:

```ant
type Celsius = float32
let room = Celsius(float32(21.5))
room          // Celsius(21.5)
float32(room) // 21.5
```

`float32` is a value, so it can be passed to a function:

```ant
[1, 2, 3].map(float32) // [1.0, 2.0, 3.0]
```

In a pipe, the value is the argument:

```ant
Math.PI | float32() // 3.1415927
```

## Applications

Store samples compactly, and widen them back to `float64` for the `Math` functions:

```ant
let samples = [0.5, 1.25, 2.0].map(float32)
let roots = samples.map({ Math.sqrt(float64(x)) })
println(samples)
println(roots)
```

```text
[0.5, 1.25, 2.0]
[0.7071067811865476, 1.118033988749895, 1.4142135623730951]
```

Average an array of `float32` readings. Every operand of the arithmetic is a `float32`, including
the count:

```ant
func mean(xs: Array[float32]): float32 {
  var sum = float32(0.0)
  for x in xs {
    sum = sum + x
  }
  sum / float32(xs.length())
}

println(mean([float32(19.5), float32(21.0), float32(20.25)]))
```

```text
20.25
```

Read a `float32` from text. `String.toFloat` returns a `float64`, which `float32` then rounds:

```ant
func parse32(s: string): float32 | null {
  let f = s.toFloat() ?? return null
  float32(f)
}

println(parse32("3.14159265358979"))
println(parse32("abc"))
```

```text
3.1415927
null
```

## Properties and relations

A `float32` widens to `float64` exactly, so converting back gives the same value:

```ant
for x in [float32(0.1), float32(1e30), float32(-7.0)] {
  assert(float32(float64(x)) == x)
}
```

An integer up to 2^24 in magnitude converts exactly:

```ant
for n in [0, 1, -42, 16777216, -16777216] {
  assert(int(float32(n)) == n)
}
```

`float32` arithmetic rounds at `float32` precision, so the result can differ from the same sum in
`float64`:

```ant
assert(float32(0.1) + float32(0.2) == float32(0.3))
assert(0.1 + 0.2 != 0.3)
```

`NaN` is not equal to itself, in `float32` as in `float64`:

```ant
let n = float32(Math.NaN)
assert(n != n)
```

## Possible issues

A float literal is a `float64`, so it is a compile error where a `float32` is expected:

```ant
let ratio: float32 = 0.5 // error: cannot use float64 where float32 is expected
```

Convert it:

```ant
let ratio: float32 = float32(0.5)
ratio // 0.5
```

A `float32` and a `float64` do not mix in arithmetic:

```ant
let width = float32(1.5)
println(width * 2.0) // error: * needs both operands of one type, got float32 and float64
```

Nor in `==`:

```ant
let x = float32(0.1)
println(x == 0.1)
// error: float32 and float64 cannot be compared with ==: neither is assignable to the other
```

Convert one side. Widening the `float32` does not give back the digits of the literal, so compare in
`float32`:

```ant
let x = float32(0.1)
float64(x) == 0.1 // false
x == float32(0.1) // true
```

The `Math` functions over `float64` do not take a `float32`:

```ant
let side = float32(2.0)
println(Math.sqrt(side)) // error: cannot use float32 where float64 is expected
```

Widen the argument, and narrow the result if you need a `float32`:

```ant
let side = float32(2.0)
float32(Math.sqrt(float64(side))) // 1.4142135
```

The generic `Math` functions take a `float32` directly:

```ant
Math.abs(float32(-1.5))              // 1.5
Math.max(float32(2.5), float32(3.0)) // 3.0
```

`String.fromFloat` takes a `float64`. Widen first, and expect the `float32` rounding to show:

```ant
String.fromFloat(float64(float32(2.675)), 2) // "2.67"
"${float32(2.675)}"                          // "2.675"
```

Above 2^24, `float32` no longer holds every integer, so adding a small value to a large one can
leave it unchanged:

```ant
let big = float32(16777216)
big + float32(1.0) == big // true
```

`float32` never panics, but converting a `float32` with a fraction to an integer kind does. Round it
first:

```ant
int(float32(2.5))                      // panic: 2.5 has a fraction, so it does not convert to int
int(Math.trunc(float64(float32(2.5)))) // 2
```

## See also

[float64](float64.md) · [int](int.md) · [int32](int32.md) · [String.toFloat](../String/toFloat.md) ·
[String.from](../String/from.md) · [String.fromFloat](../String/fromFloat.md) ·
[Math.trunc](../Math/trunc.md) · [Math.abs](../Math/abs.md) · [Number](../Math/types/Number.md) ·
[Float types](../../language/types/float.md) · [Numbers guide](../guides/numbers.md)
