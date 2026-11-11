<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int64

```ant signature
func int64[T: Number](x: T): int64
```

int64 converts x to int64 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int64(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `int64` and `int` are one type under two names, so `int64(n)` on an `int` returns `n` unchanged,
  and Ant shows the type as `int`.
- To drop a fraction, round first with `Math.trunc`, `Math.floor`, `Math.ceil`, or `Math.round`.

## Basic examples

Widen a smaller integer to `int64`:

```ant
int64(int32(-5)) // -5
int64(byte(200)) // 200
```

Convert a float that holds a whole number:

```ant
int64(2.0)   // 2
int64(-1e12) // -1000000000000
```

Convert a value from another kind before mixing it with an `int64`:

```ant
let count: int64 = 10
let extra = int32(5)
count + int64(extra) // 15
```

## Scope

Every numeric kind converts, as long as the value fits:

```ant
int64(int8(-128))            // -128
int64(uint32(4000000000))    // 4000000000
int64(uint64(1) + uint64(2)) // 3
int64(float32(1.5e9))        // 1500000000
int64(-0.0)                  // 0
```

The range ends at `Math.minInt` and `Math.maxInt`:

```ant
int64(Math.minInt)                               // -9223372036854775808
int64(Math.maxUint64 / uint64(2))                // 9223372036854775807
int64(Math.maxUint64 / uint64(2)) == Math.maxInt // true
```

A value of a named primitive type loses its tag:

```ant
type Cents = int64
let price = Cents(250)
price        // Cents(250)
int64(price) // 250
```

`int64` is a function value, so it can be passed to `Array.map`:

```ant
let small: Array[int16] = [int16(1), int16(-2), int16(300)]
small | Array.map(int64) // [1, -2, 300]
```

## Applications

Add up `int32` readings without overflow by widening them first:

```ant
let readings: Array[int32] = [int32(2000000000), int32(2000000000)]
readings | Array.map(int64) | Array.sum() // 4000000000
```

Round a measurement to the nearest whole number, then convert:

```ant
let meters = 19.6
int64(Math.round(meters)) // 20
int64(Math.trunc(meters)) // 19
```

Total the prices of a cart kept as a named primitive type:

```ant
type Cents = int64
let cart = [Cents(250), Cents(199), Cents(1000)]
println("total: ${(cart | Array.map(int64) | Array.sum()) ?? 0} cents")
```

```text
total: 1449 cents
```

Parse an unsigned number and use it where an `int` is expected:

```ant
func port(text: string): int {
  let n = String.toUint(text) ?? return -1
  if n > uint64(65535) {
    return -1
  }
  int64(n)
}

println(port("8080"))
println(port("http"))
println(port("18446744073709551615"))
```

```text
8080
-1
-1
```

## Properties and relations

Widening an integer and narrowing it back gives the same value:

```ant
for n in [int32(-7), int32(0), int32(123456)] {
  assert(int32(int64(n)) == n)
}
```

`int64` and `int` name one type, so an `int64` goes wherever an `int` does:

```ant
let n: int = int64(int8(3))
assert(type(int64) == type(int))
assert(n == 3)
```

An integer from -2^53 to 2^53 survives a round trip through `float64`:

```ant
for n in [0, -1, 9007199254740992] {
  assert(int64(float64(n)) == n)
}
```

`int64(Math.trunc(x))` drops the fraction toward zero:

```ant
int64(Math.trunc(2.7))  // 2
int64(Math.trunc(-2.7)) // -2
```

## Possible issues

A float with a fraction panics, including one that float arithmetic produced:

```ant
int64(2.5)       // panic: 2.5 has a fraction, so it does not convert to int
int64(0.1 + 0.2) // panic: 0.30000000000000004 has a fraction, so it does not convert to int
```

A value outside the range of `int64` panics:

```ant
int64(Math.maxUint64) // panic: 18446744073709551615 out of range for int
int64(1e19)           // panic: 10000000000000000000.0 out of range for int
```

`NaN` and the infinities have no integer value:

```ant
int64(Math.NaN)      // panic: NaN does not convert to int
int64(Math.Infinity) // panic: Infinity does not convert to int
```

To get `null` instead of a panic, use [Math.toInt](../Math/toInt.md):

```ant
Math.toInt(2.5) // null
Math.toInt(2.0) // 2
```

`float64(Math.maxInt)` rounds up to 2^63, which is one past `Math.maxInt`, so converting it back
panics:

```ant
int64(float64(Math.maxInt)) // panic: 9223372036854776000.0 out of range for int
```

A `float64` holds every integer only up to 2^53. Past that, the value already changed when it became
a `float64`, and `int64` returns the rounded number without a panic:

```ant
int64(float64(9007199254740993)) // 9007199254740992
```

`int64` does not parse text. Use [String.toInt](../String/toInt.md):

```ant
let n = int64("42")
// error: int converts a number or a named primitive type over a number, got string
```

Converting the result does not stop an overflow that already happened in a smaller kind. Widen
before the arithmetic:

```ant
let a = int32(2000000000)
int64(a + a)        // panic: integer overflow
int64(a) + int64(a) // 4000000000
```

## See also

[int](int.md) · [int32](int32.md) · [uint64](uint64.md) · [float64](float64.md) ·
[Math.trunc](../Math/trunc.md) · [Math.round](../Math/round.md) · [Math.toInt](../Math/toInt.md) ·
[String.toInt](../String/toInt.md) · [String.toUint](../String/toUint.md)
