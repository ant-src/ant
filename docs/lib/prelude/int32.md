<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int32

```ant signature
func int32[T: Number](x: T): int32
```

int32 converts x to int32 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int32(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- Every `int32` value lies between -2147483648 and 2147483647.
- Integer literals are always `int`. Write `int32(5)`, not `5`, wherever an `int32` is expected.
- Arithmetic on two `int32` values gives an `int32` and panics when the result leaves the range. Use
  [Bits.wrapAdd](../Bits/wrapAdd.md), [Bits.wrapSub](../Bits/wrapSub.md), and
  [Bits.wrapMul](../Bits/wrapMul.md) for wrapping arithmetic.

## Basic examples

```ant
int32(42)  // 42
int32(-7)  // -7
int32(3.0) // 3
```

The result is an `int32`, and arithmetic between two of them stays an `int32`. Integer `/` truncates
toward zero:

```ant
let width = int32(640)
let height = int32(480)
width * height     // 307200
width / int32(100) // 6
```

## Scope

Any integer kind converts, as long as the value fits:

```ant
int32(byte(200))       // 200
int32(int16(-300))     // -300
int32(uint64(4000000)) // 4000000
int32(2147483647)      // 2147483647
int32(-2147483648)     // -2147483648
```

A float converts when it has no fraction. Negative zero gives `0`:

```ant
int32(1500.0)          // 1500
int32(float32(-2.5e3)) // -2500
int32(-0.0)            // 0
```

A value of a named primitive type over a number is unwrapped first:

```ant
type Celsius = int
let today = Celsius(21)
int32(today) // 21
```

`int32` is a function value, so it can be passed to `Array.map`:

```ant
[1, 2, 3].map(int32)      // [1, 2, 3]
[10, 20].map(int32).sum() // 30
```

## Applications

Compute a 32-bit string hash, the one Java's `String.hashCode` gives. The wrapping functions let the
hash overflow on purpose:

```ant
func hashCode(s: string): int32 {
  var h = int32(0)
  for c in s.toCodePoints() {
    h = Bits.wrapAdd(Bits.wrapMul(int32(31), h), int32(c))
  }
  h
}

println(hashCode("hello"))
println(hashCode("Ant"))
```

```text
99162322
65991
```

Check that an `int` fits before converting it, and return `null` when it does not:

```ant
func toInt32(n: int): int32 | null {
  if n < -2147483648 || n > 2147483647 {
    return null
  }
  int32(n)
}

println(toInt32(1_000_000))
println(toInt32(3_000_000_000))
```

```text
1000000
null
```

Read an `int32` from text in two steps. `String.toInt` returns `int | null`, so unwrap it first,
then convert. `int32` panics if the parsed value is out of range:

```ant
let port = "8080".toInt() ?? 0
println(int32(port))
```

```text
8080
```

Round a measurement to the nearest whole number before storing it as an `int32`:

```ant
let readings = [12.4, 12.6, -0.5]
println(readings.map({ int32(Math.round(x)) }))
```

```text
[12, 13, -1]
```

## Properties and relations

Converting an `int32` back to `int` and again to `int32` gives the same value:

```ant
for n in [int32(0), int32(-2147483648), int32(2147483647)] {
  assert(int32(int(n)) == n)
  assert(int32(float64(n)) == n)
}
```

`int32(Math.trunc(x))` truncates toward zero, for negative numbers too:

```ant
for x in [2.9, -2.9, 0.5] {
  let t = int32(Math.trunc(x))
  assert(float64(t) == Math.trunc(x))
}
```

## Possible issues

An integer literal is an `int`, never an `int32`, so it does not fit where an `int32` is expected:

```ant
let n: int32 = 5     // error: cannot use int where int32 is expected
let m = int32(5) + 1 // error: + needs both operands of one type, got int32 and int
```

Convert the literal too:

```ant
let n = int32(5)
n + int32(1) // 6
```

A value out of range panics; the conversion never wraps:

```ant
int32(2147483648)         // panic: 2147483648 out of range for int32
int32(uint64(4000000000)) // panic: 4000000000 out of range for int32
```

A float with a fraction panics instead of rounding or truncating:

```ant
int32(1.5)             // panic: 1.5 has a fraction, so it does not convert to int32
int32(Math.trunc(1.5)) // 1
```

`NaN` and the infinities panic too:

```ant
int32(Math.NaN)      // panic: NaN does not convert to int32
int32(Math.Infinity) // panic: Infinity does not convert to int32
```

Arithmetic that leaves the range panics, even when the same values would fit an `int`:

```ant
let big = int32(2147483647)
big + int32(1)              // panic: integer overflow
int(big) + 1                // 2147483648
Bits.wrapAdd(big, int32(1)) // -2147483648
```

## Neat examples

Some strings hash to the most negative `int32`:

```ant
func hashCode(s: string): int32 {
  var h = int32(0)
  for c in s.toCodePoints() {
    h = Bits.wrapAdd(Bits.wrapMul(int32(31), h), int32(c))
  }
  h
}

println(hashCode("polygenelubricants"))
```

```text
-2147483648
```

## See also

[int](int.md) · [int16](int16.md) · [int64](int64.md) · [uint32](uint32.md) · [float64](float64.md)
· [Math.trunc](../Math/trunc.md) · [Math.round](../Math/round.md) · [Math.toInt](../Math/toInt.md) ·
[Bits.wrapAdd](../Bits/wrapAdd.md) · [Bits.wrapSub](../Bits/wrapSub.md) ·
[Bits.wrapMul](../Bits/wrapMul.md) · [String.toInt](../String/toInt.md) ·
[Arithmetic operators](../../language/operators/arithmetic.md)
