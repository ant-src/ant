<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# int8

```ant signature
func int8[T: Number](x: T): int8
```

int8 converts x to int8 exactly. It panics if x is out of range, has a fraction, or is NaN or
infinite; int8(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- `int8` holds whole numbers from -128 to 127.
- An integer literal is always an `int`, so an `int8` value always starts from a conversion such as
  `int8(5)`.
- Arithmetic on two `int8` values gives an `int8` and panics when the result leaves the range;
  [Bits.wrapAdd](../Bits/wrapAdd.md), [Bits.wrapSub](../Bits/wrapSub.md), and
  [Bits.wrapMul](../Bits/wrapMul.md) wrap instead.

## Basic examples

```ant
int8(42)   // 42
int8(-128) // -128
int8(127)  // 127
```

A float converts when it is a whole number:

```ant
int8(3.0)  // 3
int8(-0.0) // 0
```

A value outside the range panics:

```ant
int8(128)  // panic: 128 out of range for int8
int8(-129) // panic: -129 out of range for int8
```

## Scope

Any numeric type converts, as long as the value fits:

```ant
int8(byte(100))    // 100
int8(uint64(7))    // 7
int8(int16(-50))   // -50
int8(float32(2.0)) // 2
int8(byte(200))    // panic: 200 out of range for int8
```

A float with a fraction, `NaN`, and the infinities panic:

```ant
int8(2.5)           // panic: 2.5 has a fraction, so it does not convert to int8
int8(Math.NaN)      // panic: NaN does not convert to int8
int8(Math.Infinity) // panic: Infinity does not convert to int8
```

A named primitive over a number is unwrapped first:

```ant
type Celsius = int
let today = Celsius(21)
int8(today) // 21
```

`int8` is a function value, so it can be passed where a function is expected:

```ant
Array.map([1, -2, 3], int8) // [1, -2, 3]
```

In a pipe, the value is the argument:

```ant
100.0 | int8() // 100
```

## Applications

Read a signed byte. A `Bytes` element is a `byte` from 0 to 255; values from 128 up stand for
negative numbers in two's complement:

```ant
func signed(b: byte): int8 {
  if b < byte(128) { int8(b) } else { int8(int(b) - 256) }
}

let data = Bytes.fromArray([0x05, 0xff, 0x80, 0x7f])
for b in data {
  println(signed(b))
}
```

```text
5
-1
-128
127
```

Clamp a value into the range before converting it, so that the conversion cannot panic:

```ant
func saturate(n: int): int8 {
  int8(Math.clamp(n, -128, 127))
}

println(saturate(300))
println(saturate(-1000))
println(saturate(12))
```

```text
127
-128
12
```

Truncate or round a measurement first, then convert:

```ant
let reading = -12.8
println(int8(Math.trunc(reading)))
println(int8(Math.round(reading)))
```

```text
-12
-13
```

## Properties and relations

Every `int` from -128 to 127 survives the round trip through `int8`:

```ant
for n in Iter.range(-128, 128) {
  assert(int(int8(n)) == n)
}
```

Converting a whole float gives the same `int8` as converting the matching `int`:

```ant
for n in [-128, -1, 0, 64, 127] {
  assert(int8(float64(n)) == int8(n))
}
```

`Bits.wrapAdd` agrees with `+` whenever `+` does not overflow:

```ant
let a = int8(100)
let b = int8(27)
assert(Bits.wrapAdd(a, b) == a + b)
```

## Possible issues

An integer literal is an `int`, never an `int8`, so it does not fit an `int8` slot. Convert it:

```ant
let level: int8 = 5 // error: cannot use int where int8 is expected
```

```ant
let level: int8 = int8(5)
level // 5
```

An array literal of integers is an `Array[int]`, so it does not fit an `Array[int8]` slot:

```ant
let xs: Array[int8] = [1, 2, 3] // error: cannot use int where int8 is expected
```

Convert each element instead:

```ant
let xs = Array.map([1, 2, 3], int8)
xs // [1, 2, 3]
```

A hex literal is an `int` too, so `0xff` is 255, not -1, and does not fit. To read a byte as a
signed number, see the signed-byte example under Applications:

```ant
int8(0xff) // panic: 255 out of range for int8
```

`int8` and `int` do not mix in arithmetic. Convert one side so that both have the same type:

```ant
let step = int8(3)
let next = step + 1 // error: + needs both operands of one type, got int8 and int
```

```ant
let step = int8(3)
step + int8(1) // 4
int(step) + 1  // 4
```

Arithmetic overflows quickly in eight bits, and overflow panics:

```ant
let a = int8(100)
a + a              // panic: integer overflow
Bits.wrapAdd(a, a) // -56
```

The range is not symmetric: -128 has no positive counterpart, so negating it overflows:

```ant
let low = int8(-128)
-low           // panic: integer overflow
low / int8(-1) // panic: integer overflow
```

## See also

[int16](int16.md) · [int](int.md) · [uint8](uint8.md) · [byte](byte.md) · [float64](float64.md) ·
[Math.trunc](../Math/trunc.md) · [Math.clamp](../Math/clamp.md) · [Bits.wrapAdd](../Bits/wrapAdd.md)
· [Numbers](../../language/literals/numbers.md) ·
[Arithmetic](../../language/operators/arithmetic.md)
